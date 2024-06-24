package cmdexec

import "github.com/stanleygy/toy-redis/app/resp"

type ZsetCmdExecutor struct{}

/*
 * syntax: ZADD key [NX] score member
 * syntax: ZREM key member [member ...]
 * syntax: ZSCORE key member
 * syntax: ZCOUNT key min max
 * syntax: ZRANGE key start stop [REV] [LIMIT offset count]  [WITHSCORES]
 */

func (e ZsetCmdExecutor) Execute(cmdName string, cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	return nil, nil
}
