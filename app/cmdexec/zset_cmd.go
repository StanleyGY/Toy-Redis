package cmdexec

import (
	"strconv"
	"time"

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

	// Check if a new key is requested, so a new skip list is created
	sset, found := db.SortedSetStore[key]
	if !found {
		db.SortedSetStore[key] = MakeSkipList(time.Now().Unix())
		sset = db.SortedSetStore[key]
	}

	numAdded := 0
	for i := 0; i < len(members); i++ {
		if sset.Add(members[i], scores[i], nxFlag) {
			numAdded++
		}
	}

	return &resp.RespValue{DataType: resp.TypeIntegers, Int: numAdded}, nil
}

/*
 * syntax: ZREM key member [member ...]
 * syntax: ZSCORE key member
 * syntax: ZCOUNT key min max
 * syntax: ZRANGE key start stop [REV] [LIMIT offset count]  [WITHSCORES]
 */

func (e ZsetCmdExecutor) Execute(cmdName string, cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	switch cmdName {
	case "ZADD":
		return e.executeZAddCmd(cmdArgs)
	case "ZREM":
	}
	return nil, nil
}
