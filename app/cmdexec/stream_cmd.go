package cmdexec

import (
	"math"
	"strconv"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
	"github.com/stanleygy/toy-redis/app/event"
	"github.com/stanleygy/toy-redis/app/resp"
)

type StreamCmdExecutor struct{}

/*
Syntax: XADD key [MAXLEN ~ threshold [LIMIT count]] <*> field value [field value ...]
Example:
  - XADD mystream MAXLEN ~ 1000 *

Reply:
  - Bulk string reply: the ID of the added entry
*/
func (e StreamCmdExecutor) parseXAddCmdArgs(cmdArgs []*resp.RespValue, key *string, id *string, fieldValues *[]string) error {
	if len(cmdArgs) < 2 {
		return ErrInvalidArgs
	}

	*key = cmdArgs[0].BulkStr
	*id = cmdArgs[1].BulkStr

	if len(cmdArgs)%2 != 0 {
		// One field is missing value
		return ErrInvalidArgs
	}

	for i := 2; i < len(cmdArgs); i += 2 {
		*fieldValues = append(*fieldValues, cmdArgs[i].BulkStr)
		*fieldValues = append(*fieldValues, cmdArgs[i+1].BulkStr)
	}
	return nil
}

func (e StreamCmdExecutor) generateStreamId(stream *Stream, id *string) error {
	if *id == "*" {
		// Generate a stream ID
		unixMs := time.Now().UnixMilli()
		if unixMs == stream.LastId.Ms {
			if stream.LastId.Seq == math.MaxUint {
				return ErrOverflow
			}
			stream.LastId.Seq++
		} else {
			stream.LastId.Ms = unixMs
			stream.LastId.Seq = 0
		}
		*id = stream.LastId.ToString()
		return nil
	}
	// TODO: support customized stream ID
	return ErrInvalidArgs
}

func (e StreamCmdExecutor) executeXAddCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key         string
		id          string
		fieldValues []string
	)
	err := e.parseXAddCmdArgs(cmdArgs, &key, &id, &fieldValues)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	// Loop up the stream at key
	stream, found := db.StreamStore[key]
	if !found {
		// If stream key does not exist, create one
		stream = &Stream{
			Radix: algo.MakeRadixTree(),
			LastId: &StreamID{
				Ms:  0,
				Seq: 0,
			},
		}
		db.StreamStore[key] = stream
	}

	// Generate stream ID
	err = e.generateStreamId(stream, &id)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	stream.Radix.Insert(id, fieldValues)
	stream.Radix.Visualize()
	event.AddBulkStringReplyEvent(c, id)
}

/*
Syntax: XRANGE key start end [COUNT count]
Reply:
  - Array reply: a list of stream entries with IDs matching the specified range
*/
func (e StreamCmdExecutor) parseXRangeCmdArgs(cmdArgs []*resp.RespValue, key *string, start *string, end *string, count *int) error {
	if len(cmdArgs) < 3 || len(cmdArgs) > 4 {
		return ErrInvalidArgs
	}
	var err error

	*key = cmdArgs[0].BulkStr
	*start = cmdArgs[1].BulkStr
	*end = cmdArgs[2].BulkStr

	if len(cmdArgs) == 4 {
		*count, err = strconv.Atoi(cmdArgs[3].BulkStr)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e StreamCmdExecutor) generateXRangeCmdReplyEvent(c *event.ClientInfo, searchResults []*algo.RadixSearchResult) {
	resps := make([]*resp.RespValue, len(searchResults))

	for i, result := range searchResults {
		values := make([]*resp.RespValue, len(result.Node.Values))
		for j, v := range result.Node.Values {
			values[j] = resp.MakeBulkString(v)
		}

		entry := &resp.RespValue{
			DataType: resp.TypeArrays,
			Array: []*resp.RespValue{
				resp.MakeBulkString(result.Id),             // stream id
				{DataType: resp.TypeArrays, Array: values}, // field and values
			},
		}
		resps[i] = entry
	}
	event.AddArrayReplyEvent(c, resps)
}

func (e StreamCmdExecutor) executeXRangeCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key   string
		start string
		end   string
		count int = math.MaxInt
	)

	err := e.parseXRangeCmdArgs(cmdArgs, &key, &start, &end, &count)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	// Loop up the stream at key
	stream, found := db.StreamStore[key]
	if !found {
		event.AddArrayReplyEvent(c, []*resp.RespValue{})
		return
	}

	// Perform the search
	if start == "-" {
		// Ascii value lower than all valid stream Ids
		start = "0"
	}
	if end == "+" {
		// Ascii value greater than all valid stream Ids
		end = ":"
	}
	searchResults := stream.Radix.SearchByRange(start, end, count)
	e.generateXRangeCmdReplyEvent(c, searchResults)
}

/*
Syntax: XREAD [COUNT count] [BLOCK milliseconds] STREAMS key id  TODO: support multiple keys
Reply:
  - Array reply: a list of stream entries with IDs matching the specified range
*/
func (e StreamCmdExecutor) executeXReadCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		count int
		block int = -1
		key   string
		id    string
	)

	// Loop up the stream at key
	stream, found := db.StreamStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeArrays, Array: []*resp.RespValue{}}, nil
	}

	searchResults := stream.Radix.SearchByRange(id, ":", count)
	if len(searchResults) == 0 && block != -1 {
		// Block the client
	}

	return nil, nil
}

func (e StreamCmdExecutor) Execute(c *event.ClientInfo, cmdName string, cmdArgs []*resp.RespValue) {
	switch cmdName {
	case "XADD":
		e.executeXAddCmd(c, cmdArgs)
	case "XRANGE":
		e.executeXRangeCmd(c, cmdArgs)
	case "XREAD":
		e.executeXReadCmd(c, cmdArgs)
	}
}
