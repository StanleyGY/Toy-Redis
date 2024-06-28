package cmdexec

import (
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
	"github.com/stanleygy/toy-redis/app/resp"
)

type ZsetCmdExecutor struct{}

/*
 * syntax: ZADD key [NX] score member [score member ...]
 */
func (e ZsetCmdExecutor) parseZAddCmdArgs(cmdArgs []*resp.RespValue, key *string, members *[]string, scores *[]int, nxFlag *bool) error {
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

func (e ZsetCmdExecutor) executeZAddCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key     string
		members []string = make([]string, 0)
		scores  []int    = make([]int, 0)
		nxFlag  bool
	)

	err := e.parseZAddCmdArgs(cmdArgs, &key, &members, &scores, &nxFlag)
	if err != nil {
		return nil, err
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

	return &resp.RespValue{DataType: resp.TypeIntegers, Int: numAdded}, nil
}

/*
 * syntax: ZREM key member [member ...]
 */
func (e ZsetCmdExecutor) parseZRemCmdArgs(cmdArgs []*resp.RespValue, key *string, members *[]string) {
	*key = cmdArgs[0].BulkStr
	for i := 1; i < len(cmdArgs); i++ {
		*members = append(*members, cmdArgs[i].BulkStr)
	}
}

func (e ZsetCmdExecutor) executeZRemCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key     string
		members []string = make([]string, 0)
	)
	e.parseZRemCmdArgs(cmdArgs, &key, &members)

	// Check if the key exists
	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeIntegers, Int: 0}, nil
	}

	numRemoved := 0
	for i := 0; i < len(members); i++ {
		if sortedSet.Remove(members[i]) {
			numRemoved++
		}
	}

	return &resp.RespValue{DataType: resp.TypeIntegers, Int: numRemoved}, nil
}

/*
 * syntax: ZSCORE key member
 */
func (e ZsetCmdExecutor) parseZScoreCmdArgs(cmdArgs []*resp.RespValue, key *string, member *string) error {
	if len(cmdArgs) != 2 {
		return ErrInvalidArgs
	}
	*key = cmdArgs[0].BulkStr
	*member = cmdArgs[1].BulkStr
	return nil
}

func (e ZsetCmdExecutor) executeZScoreCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key    string
		member string
	)

	err := e.parseZScoreCmdArgs(cmdArgs, &key, &member)
	if err != nil {
		return nil, err
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	score := sortedSet.GetScore(member)
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: strconv.Itoa(score)}, nil
}

/*
 * syntax: ZCOUNT key min max
 */
func (e ZsetCmdExecutor) parseZCountCmdArgs(cmdArgs []*resp.RespValue, key *string, min *int, max *int) error {
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

func (e ZsetCmdExecutor) executeZCountCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key string
		min int
		max int
	)

	err := e.parseZCountCmdArgs(cmdArgs, &key, &min, &max)
	if err != nil {
		return nil, err
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	numElems := sortedSet.CountByRange(min, max)
	return &resp.RespValue{DataType: resp.TypeIntegers, Int: numElems}, nil
}

/*
 * syntax: ZRANGEBYSCORE key min max [WITHSCORES]
 */
func (e ZsetCmdExecutor) parseZRangeByScoreCmdArgs(cmdArgs []*resp.RespValue, key *string, min *int, max *int, withScoresFlag *bool) error {
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

func (e ZsetCmdExecutor) executeZRangeByScoreCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key            string
		min            int
		max            int
		withScoresFlag bool
	)

	err := e.parseZRangeByScoreCmdArgs(cmdArgs, &key, &min, &max, &withScoresFlag)
	if err != nil {
		return nil, err
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	nodes := sortedSet.FindByRange(min, max)
	res := make([]*resp.RespValue, 0)
	for _, node := range nodes {
		res = append(res, &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: node.Member})
		if withScoresFlag {
			res = append(res, &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: strconv.Itoa(node.Score)})
		}
	}
	return &resp.RespValue{DataType: resp.TypeArrays, Array: res}, nil
}

/*
Syntax: ZRANK key member [WITHSCORE]
Reply:
  - Null reply: if key or member does not exist
  - Integer reply: the rank of the member when WITHSCORE is not used
  - Array reply: the rank of the member when WITHSCORE is used
*/
func (e ZsetCmdExecutor) parseZRankCmdArgs(cmdArgs []*resp.RespValue, key *string, member *string, withScoreFlag *bool) error {
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

func (e ZsetCmdExecutor) executeZRankCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key           string
		member        string
		withScoreFlag bool
	)
	err := e.parseZRankCmdArgs(cmdArgs, &key, &member, &withScoreFlag)
	if err != nil {
		return nil, err
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	node, rank := sortedSet.GetRank(member)
	if node == nil {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	if !withScoreFlag {
		return &resp.RespValue{DataType: resp.TypeIntegers, Int: rank}, nil
	}
	return &resp.RespValue{DataType: resp.TypeArrays, Array: []*resp.RespValue{
		{DataType: resp.TypeIntegers, Int: rank},
		{DataType: resp.TypeBulkStrings, BulkStr: strconv.Itoa(node.Score)},
	}}, nil
}

/*
Syntax: ZRANGE key start stop [WITHSCORES]
Reply:
  - Array reply: a list of members with, optionally, their scores
*/
func (e ZsetCmdExecutor) parseZRangeCmdArgs(cmdArgs []*resp.RespValue, key *string, start *int, stop *int, withScoreFlag *bool) error {
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

func (e ZsetCmdExecutor) executeZRangeCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key           string
		start         int
		stop          int
		withScoreFlag bool
	)
	err := e.parseZRangeCmdArgs(cmdArgs, &key, &start, &stop, &withScoreFlag)
	if err != nil {
		return nil, err
	}

	sortedSet, found := db.SortedSetStore[key]
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	nodes := sortedSet.FindByRanks(start, stop)
	if nodes == nil {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}

	res := make([]*resp.RespValue, 0)
	for _, node := range nodes {
		res = append(res, &resp.RespValue{
			DataType: resp.TypeBulkStrings,
			BulkStr:  node.Member,
		})
		if withScoreFlag {
			res = append(res, &resp.RespValue{
				DataType: resp.TypeBulkStrings,
				BulkStr:  strconv.Itoa(node.Score),
			})
		}
	}
	return &resp.RespValue{DataType: resp.TypeArrays, Array: res}, nil
}

func (e ZsetCmdExecutor) Execute(cmdName string, cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	switch cmdName {
	case "ZSCORE":
		return e.executeZScoreCmd(cmdArgs)
	case "ZADD":
		return e.executeZAddCmd(cmdArgs)
	case "ZREM":
		return e.executeZRemCmd(cmdArgs)
	case "ZCOUNT":
		return e.executeZCountCmd(cmdArgs)
	case "ZRANGEBYSCORE":
		return e.executeZRangeByScoreCmd(cmdArgs)
	case "ZRANK":
		return e.executeZRankCmd(cmdArgs)
	case "ZRANGE":
		return e.executeZRangeCmd(cmdArgs)
	}
	return nil, nil
}
