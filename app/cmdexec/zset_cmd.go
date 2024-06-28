package cmdexec

import (
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
	"github.com/stanleygy/toy-redis/app/event"
	"github.com/stanleygy/toy-redis/app/resp"
)

type zsetCmdExecutor struct{}

/*
 * syntax: ZADD key [NX] score member [score member ...]
 */
func (e zsetCmdExecutor) parseZAddCmdArgs(cmdArgs []*resp.RespValue, key *string, members *[]string, scores *[]int, nxFlag *bool) error {
	*key = cmdArgs[0].BulkStr

	for i := 1; i < len(cmdArgs); i++ {
		if cmdArgs[i].BulkStr == "NX" {
			*nxFlag = true
		} else {
			score, err := strconv.Atoi(cmdArgs[i].BulkStr)
			if err != nil {
				return err
			}
			*scores = append(*scores, score)
			*members = append(*members, cmdArgs[i+1].BulkStr)
			i++
		}
	}
	return nil
}

func (e zsetCmdExecutor) executeZAddCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key     string
		members []string = make([]string, 0)
		scores  []int    = make([]int, 0)
		nxFlag  bool
	)

	err := e.parseZAddCmdArgs(cmdArgs, &key, &members, &scores, &nxFlag)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	// If a new key is requested, a new skip list will be created
	sortedSet, found := db.SortedSetStore[key]
	if !found {
		db.SortedSetStore[key] = algo.MakeSkipList(time.Now().Unix())
		sortedSet = db.SortedSetStore[key]
	}

	numAdded := 0
	for i := 0; i < len(members); i++ {
		if sortedSet.Add(members[i], scores[i], nxFlag) {
			numAdded++
		}
	}

	event.AddIntegerReplyEvent(c, numAdded)
}

/*
 * syntax: ZREM key member [member ...]
 */
func (e zsetCmdExecutor) parseZRemCmdArgs(cmdArgs []*resp.RespValue, key *string, members *[]string) {
	*key = cmdArgs[0].BulkStr
	for i := 1; i < len(cmdArgs); i++ {
		*members = append(*members, cmdArgs[i].BulkStr)
	}
}

func (e zsetCmdExecutor) executeZRemCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key     string
		members []string = make([]string, 0)
	)
	e.parseZRemCmdArgs(cmdArgs, &key, &members)

	// Check if the key exists
	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddIntegerReplyEvent(c, 0)
		return
	}

	numRemoved := 0
	for i := 0; i < len(members); i++ {
		if sortedSet.Remove(members[i]) {
			numRemoved++
		}
	}
	event.AddIntegerReplyEvent(c, numRemoved)
}

/*
 * syntax: ZSCORE key member
 */
func (e zsetCmdExecutor) parseZScoreCmdArgs(cmdArgs []*resp.RespValue, key *string, member *string) error {
	if len(cmdArgs) != 2 {
		return ErrInvalidArgs
	}
	*key = cmdArgs[0].BulkStr
	*member = cmdArgs[1].BulkStr
	return nil
}

func (e zsetCmdExecutor) executeZScoreCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key    string
		member string
	)

	err := e.parseZScoreCmdArgs(cmdArgs, &key, &member)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddNullBulkStringReplyEvent(c)
		return
	}

	score := sortedSet.GetScore(member)
	event.AddBulkStringReplyEvent(c, strconv.Itoa(score))
}

/*
 * syntax: ZCOUNT key min max
 */
func (e zsetCmdExecutor) parseZCountCmdArgs(cmdArgs []*resp.RespValue, key *string, min *int, max *int) error {
	if len(cmdArgs) != 3 {
		return ErrInvalidArgs
	}

	var err error

	*key = cmdArgs[0].BulkStr
	*min, err = strconv.Atoi(cmdArgs[1].BulkStr)
	if err != nil {
		return err
	}

	*max, err = strconv.Atoi(cmdArgs[2].BulkStr)
	if err != nil {
		return err
	}
	return nil
}

func (e zsetCmdExecutor) executeZCountCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key string
		min int
		max int
	)

	err := e.parseZCountCmdArgs(cmdArgs, &key, &min, &max)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddNullBulkStringReplyEvent(c)
		return
	}

	numElems := sortedSet.CountByRange(min, max)
	event.AddIntegerReplyEvent(c, numElems)
}

/*
 * syntax: ZRANGEBYSCORE key min max [WITHSCORES]
 */
func (e zsetCmdExecutor) parseZRangeByScoreCmdArgs(cmdArgs []*resp.RespValue, key *string, min *int, max *int, withScoresFlag *bool) error {
	if len(cmdArgs) < 3 || len(cmdArgs) > 4 {
		return ErrInvalidArgs
	}
	var err error
	*key = cmdArgs[0].BulkStr
	*min, err = strconv.Atoi(cmdArgs[1].BulkStr)
	if err != nil {
		return err
	}

	*max, err = strconv.Atoi(cmdArgs[2].BulkStr)
	if err != nil {
		return err
	}

	if len(cmdArgs) == 4 {
		if strings.ToUpper(cmdArgs[3].BulkStr) == "WITHSCORES" {
			*withScoresFlag = true
		} else {
			return ErrInvalidArgs
		}
	}
	return nil
}

func (e zsetCmdExecutor) executeZRangeByScoreCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key            string
		min            int
		max            int
		withScoresFlag bool
	)

	err := e.parseZRangeByScoreCmdArgs(cmdArgs, &key, &min, &max, &withScoresFlag)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddNullBulkStringReplyEvent(c)
		return
	}

	nodes := sortedSet.FindByRange(min, max)

	// Generate reply event
	res := make([]*resp.RespValue, 0)
	for _, node := range nodes {
		res = append(res, resp.MakeBulkString(node.Member))
		if withScoresFlag {
			res = append(res, resp.MakeBulkString(strconv.Itoa(node.Score)))
		}
	}
	event.AddArrayReplyEvent(c, res)
}

/*
Syntax: ZRANK key member [WITHSCORE]
Reply:
  - Null reply: if key or member does not exist
  - Integer reply: the rank of the member when WITHSCORE is not used
  - Array reply: the rank of the member when WITHSCORE is used
*/
func (e zsetCmdExecutor) parseZRankCmdArgs(cmdArgs []*resp.RespValue, key *string, member *string, withScoreFlag *bool) error {
	if len(cmdArgs) < 2 || len(cmdArgs) > 3 {
		return ErrInvalidArgs
	}

	*key = cmdArgs[0].BulkStr
	*member = cmdArgs[1].BulkStr

	if len(cmdArgs) == 3 {
		if cmdArgs[2].BulkStr == "WITHSCORE" {
			*withScoreFlag = true
		} else {
			return ErrInvalidArgs
		}
	}
	return nil
}

func (e zsetCmdExecutor) executeZRankCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key           string
		member        string
		withScoreFlag bool
	)
	err := e.parseZRankCmdArgs(cmdArgs, &key, &member, &withScoreFlag)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	// Look up sorted set at key
	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddNullBulkStringReplyEvent(c)
		return
	}
	node, rank := sortedSet.GetRank(member)
	if node == nil {
		event.AddNullBulkStringReplyEvent(c)
		return
	}

	// Generate reply events
	if !withScoreFlag {
		event.AddIntegerReplyEvent(c, rank)
		return
	}
	event.AddArrayReplyEvent(c, []*resp.RespValue{
		resp.MakeInt(rank),
		resp.MakeBulkString(strconv.Itoa(node.Score)),
	})
}

/*
Syntax: ZRANGE key start stop [WITHSCORES]
Reply:
  - Array reply: a list of members with, optionally, their scores
*/
func (e zsetCmdExecutor) parseZRangeCmdArgs(cmdArgs []*resp.RespValue, key *string, start *int, stop *int, withScoreFlag *bool) error {
	if len(cmdArgs) < 3 || len(cmdArgs) > 4 {
		return ErrInvalidArgs
	}
	var err error

	*key = cmdArgs[0].BulkStr

	*start, err = strconv.Atoi(cmdArgs[1].BulkStr)
	if err != nil {
		return err
	}

	*stop, err = strconv.Atoi(cmdArgs[2].BulkStr)
	if err != nil {
		return err
	}

	if len(cmdArgs) == 4 {
		if cmdArgs[3].BulkStr == "WITHSCORE" {
			*withScoreFlag = true
		} else {
			return ErrInvalidArgs
		}
	}
	return nil
}

func (e zsetCmdExecutor) executeZRangeCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key           string
		start         int
		stop          int
		withScoreFlag bool
	)
	err := e.parseZRangeCmdArgs(cmdArgs, &key, &start, &stop, &withScoreFlag)
	if err != nil {
		event.AddErrorReplyEvent(c, err)
		return
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		event.AddNullBulkStringReplyEvent(c)
		return
	}
	nodes := sortedSet.FindByRanks(start, stop)
	if nodes == nil {
		event.AddNullBulkStringReplyEvent(c)
		return
	}

	// Generate reply events
	res := make([]*resp.RespValue, 0)
	for _, node := range nodes {
		res = append(res, resp.MakeBulkString(node.Member))
		if withScoreFlag {
			res = append(res, resp.MakeBulkString(strconv.Itoa(node.Score)))
		}
	}
	event.AddArrayReplyEvent(c, res)
}

func (e zsetCmdExecutor) Execute(c *event.ClientInfo, cmdName string, cmdArgs []*resp.RespValue) {
	switch cmdName {
	case "ZSCORE":
		e.executeZScoreCmd(c, cmdArgs)
	case "ZADD":
		e.executeZAddCmd(c, cmdArgs)
	case "ZREM":
		e.executeZRemCmd(c, cmdArgs)
	case "ZCOUNT":
		e.executeZCountCmd(c, cmdArgs)
	case "ZRANGEBYSCORE":
		e.executeZRangeByScoreCmd(c, cmdArgs)
	case "ZRANK":
		e.executeZRankCmd(c, cmdArgs)
	case "ZRANGE":
		e.executeZRangeCmd(c, cmdArgs)
	}
}
