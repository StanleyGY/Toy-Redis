package cmdexec

import (
	"errors"
	"strings"

	"github.com/stanleygy/toy-redis/app/event"
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
	"XADD":          &StreamCmdExecutor{},
	"XRANGE":        &StreamCmdExecutor{},
	"XREAD":         &StreamCmdExecutor{},
}

func Execute(c *event.ClientInfo, val *resp.RespValue) {
	cmdName := strings.ToUpper(val.Array[0].BulkStr)
	cmd := CmdLookupTable[cmdName]
	if cmd == nil {
		event.AddErrorReplyEvent(c, errors.New("failed to look up command"))
		return
	}
	cmdArgs := val.Array[1:]
	cmd.Execute(c, cmdName, cmdArgs)
}

type cmdExecutor interface {
	Execute(client *event.ClientInfo, cmdName string, cmdArgs []*resp.RespValue)
}

/*
 * syntax: PING
 */
type pingCmdExecutor struct{}

func (pingCmdExecutor) Execute(c *event.ClientInfo, _ string, _ []*resp.RespValue) {
	event.AddSimpleStringReplyEvent(c, "PONG")
}

/*
 * syntax: ECHO message
 */
type echoCmdExecutor struct{}

func (echoCmdExecutor) Execute(c *event.ClientInfo, _ string, args []*resp.RespValue) {
	if len(args) == 0 {
		event.AddErrorReplyEvent(c, ErrInvalidArgs)
	} else {
		event.AddBulkStringReplyEvent(c, "PONG")
	}
}
