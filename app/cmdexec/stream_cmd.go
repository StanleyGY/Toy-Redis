package cmdexec

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
	"github.com/stanleygy/toy-redis/app/resp"
)

type streamCmdExecutor struct{}

func (e streamCmdExecutor) generateStreamId(stream *Stream, id *string) error {
	if *id == "*" {
		// Generate a stream ID
		unixMs := time.Now().UnixMilli()
		if unixMs == stream.LastId.Ms {
			err := stream.LastId.Incr()
			if err != nil {
				return err
			}
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

/*
Syntax: XADD key <*> field value [field value ...]
Example:
  - XADD mystream *

Reply:
  - Bulk string reply: the ID of the added entry
*/
func (e streamCmdExecutor) parseXAddCmdArgs(cmdArgs []*resp.RespValue, key *string, id *string, fieldValues *[]string) error {
	if len(cmdArgs) < 4 {
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

func (e streamCmdExecutor) executeXAddCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key         string
		id          string
		fieldValues []string
	)
	err := e.parseXAddCmdArgs(cmdArgs, &key, &id, &fieldValues)
	if err != nil {
		AddErrorReplyEvent(c, err)
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
		AddErrorReplyEvent(c, err)
		return
	}

	stream.Radix.Insert(id, fieldValues)
	stream.Radix.Visualize()
	AddBulkStringReplyEvent(c, id)

	NotifyBlockedClientsOnKeySpace(&BlockKey{Source: BlockOnStream, Key: key})
}

/*
Syntax: XRANGE key start end [COUNT count]
Reply:
  - Array reply: a list of stream entries with IDs matching the specified range
*/
func (e streamCmdExecutor) parseXRangeCmdArgs(cmdArgs []*resp.RespValue, key *string, start *string, end *string, count *int) error {
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

func (e streamCmdExecutor) generateSearchResultsReplyEvent(c *ClientInfo, searchResults []*algo.RadixSearchResult) {
	resps := make([]*resp.RespValue, len(searchResults))

	for i, result := range searchResults {
		fieldValues := result.Node.Value.([]string)
		values := make([]*resp.RespValue, len(fieldValues))
		for j, v := range fieldValues {
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
	AddArrayReplyEvent(c, resps)
}

func (e streamCmdExecutor) executeXRangeCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key   string
		start string
		end   string
		count int = math.MaxInt
	)

	err := e.parseXRangeCmdArgs(cmdArgs, &key, &start, &end, &count)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	// Loop up the stream at key
	stream, found := db.StreamStore[key]
	if !found {
		AddArrayReplyEvent(c, []*resp.RespValue{})
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
	e.generateSearchResultsReplyEvent(c, searchResults)
}

/*
Syntax: XREAD [COUNT count] [BLOCK milliseconds] STREAMS key [key...] id [id...]
Reply:
  - Array reply: a list of stream entries with IDs matching the specified range
*/
func (e streamCmdExecutor) parseXReadCmdArgs(cmdArgs []*resp.RespValue, count *int, timeout *int, keys *[]string, ids *[]string) error {
	var err error
	i := 0

	// Parse options
	for ; i < len(cmdArgs) && cmdArgs[i].BulkStr != "STREAMS"; i++ {
		option := strings.ToUpper(cmdArgs[i].BulkStr)
		if option == "COUNT" || option == "BLOCK" {
			if i+1 == len(cmdArgs) {
				return ErrInvalidArgs
			}
			if option == "COUNT" {
				*count, err = strconv.Atoi(cmdArgs[i+1].BulkStr)
			} else {
				*timeout, err = strconv.Atoi(cmdArgs[i+1].BulkStr)
			}
			if err != nil {
				return err
			}
			i++
		}
	}
	if i >= len(cmdArgs) {
		return ErrInvalidArgs
	}

	// Parse keys and ids
	i++
	keysAndIds := make([]string, 0)
	for ; i < len(cmdArgs); i++ {
		keysAndIds = append(keysAndIds, cmdArgs[i].BulkStr)
	}
	if len(keysAndIds)%2 != 0 {
		return ErrInvalidArgs
	}
	*keys = keysAndIds[:len(keysAndIds)/2]
	*ids = keysAndIds[len(keysAndIds)/2:]
	return nil
}

func (e streamCmdExecutor) executeXReadCmd(c *ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		count    int = math.MaxInt
		timeout  int = -1
		keys     []string
		startIds []string
	)
	err := e.parseXReadCmdArgs(cmdArgs, &count, &timeout, &keys, &startIds)
	if err != nil {
		AddErrorReplyEvent(c, err)
		return
	}

	var searchResults []*algo.RadixSearchResult
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		startId := startIds[i]

		// Loop up the stream at key
		stream, found := db.StreamStore[key]
		if !found {
			AddEmptyArrayReplyEvent(c)
			return
		}

		// The start id is exclusive, so incr the start id to make it inclusive for the search
		sid, err := ParseStreamID(startId)
		if err != nil {
			AddErrorReplyEvent(c, err)
			return
		}
		err = sid.Incr()
		if err != nil {
			AddErrorReplyEvent(c, err)
			return
		}
		searchResults = append(searchResults, stream.Radix.SearchByRange(sid.ToString(), ":", count)...)
	}

	if len(searchResults) == 0 {
		if timeout != -1 {
			// If no results are returned, and client specify block option,
			// Block the client until a key space occurs or client times out
			var bkeys []*BlockKey
			for _, key := range keys {
				bkeys = append(bkeys, &BlockKey{Source: BlockOnStream, Key: key})
			}
			BlockClientForKeys(c, bkeys, timeout)
		} else {
			// Immediately reply to client
			AddEmptyArrayReplyEvent(c)
		}
		return
	}

	UnblockClient(c)
	e.generateSearchResultsReplyEvent(c, searchResults)
}

func (e streamCmdExecutor) Execute(c *ClientInfo, cmdName string, cmdArgs []*resp.RespValue) {
	switch cmdName {
	case "XADD":
		e.executeXAddCmd(c, cmdArgs)
	case "XRANGE":
		e.executeXRangeCmd(c, cmdArgs)
	case "XREAD":
		e.executeXReadCmd(c, cmdArgs)
	}
}
