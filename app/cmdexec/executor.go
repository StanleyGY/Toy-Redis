package cmdexec

import (
	"errors"
	"strings"

	"github.com/stanleygy/toy-redis/app/resp"
)

var (
	ErrInvalidArgs = errors.New("invalid args")
	ErrOverflow    = errors.New("overflow")
)

var CmdLookupTable = map[string]cmdExecutor{
	"COMMAND":       &pingCmdExecutor{},
	"PING":          &pingCmdExecutor{},
	"ECHO":          &echoCmdExecutor{},
	"SET":           &setCmdExecutor{},
	"GET":           &setCmdExecutor{},
	"ZADD":          &zsetCmdExecutor{},
	"ZREM":          &zsetCmdExecutor{},
	"ZSCORE":        &zsetCmdExecutor{},
	"ZCOUNT":        &zsetCmdExecutor{},
	"ZRANGEBYSCORE": &zsetCmdExecutor{},
	"ZRANK":         &zsetCmdExecutor{},
	"ZRANGE":        &zsetCmdExecutor{},
	"XADD":          &streamCmdExecutor{},
	"XRANGE":        &streamCmdExecutor{},
	"XREAD":         &streamCmdExecutor{},
	"GEOADD":        &geoCmdExecutor{},
	"GEODIST":       &geoCmdExecutor{},
	"GEOHASH":       &geoCmdExecutor{},
}

func Execute(c *ClientInfo, val *resp.RespValue) {
	cmdName := strings.ToUpper(val.Array[0].BulkStr)
	cmd := CmdLookupTable[cmdName]
	if cmd == nil {
		AddErrorReplyEvent(c, errors.New("failed to look up command"))
		return
	}
	cmdArgs := val.Array[1:]
	cmd.Execute(c, cmdName, cmdArgs)
}

type cmdExecutor interface {
	Execute(client *ClientInfo, cmdName string, cmdArgs []*resp.RespValue)
}

/*
 * syntax: PING
 */
type pingCmdExecutor struct{}

func (pingCmdExecutor) Execute(c *ClientInfo, _ string, _ []*resp.RespValue) {
	AddSimpleStringReplyEvent(c, "PONG")
}

/*
 * syntax: ECHO message
 */
type echoCmdExecutor struct{}

func (echoCmdExecutor) Execute(c *ClientInfo, _ string, args []*resp.RespValue) {
	if len(args) == 0 {
		AddErrorReplyEvent(c, ErrInvalidArgs)
	} else {
		AddBulkStringReplyEvent(c, "PONG")
	}
}
