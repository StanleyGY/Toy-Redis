package cmdexec

import (
	"errors"
	"strings"

	"github.com/stanleygy/toy-redis/app/resp"
)

var (
	ErrInvalidArgs = errors.New("invalid args")
)

type CmdExecutor interface {
	Execute(cmdName string, cmdArgs []*resp.RespValue) (*resp.RespValue, error)
}

var CmdLookupTable = map[string]CmdExecutor{
	"COMMAND": &PingCmdExecutor{},
	"PING":    &PingCmdExecutor{},
	"ECHO":    &EchoCmdExecutor{},
	"SET":     &SetCmdExecutor{},
	"GET":     &SetCmdExecutor{},
	"ZADD":    &ZsetCmdExecutor{},
	"ZREM":    &ZsetCmdExecutor{},
	"ZSCORE":  &ZsetCmdExecutor{},
}

func Execute(val *resp.RespValue) (*resp.RespValue, error) {
	cmdName := strings.ToUpper(val.Array[0].BulkStr)
	cmd := CmdLookupTable[cmdName]
	if cmd == nil {
		return nil, errors.New("failed to look up command")
	}
	cmdArgs := val.Array[1:]
	return cmd.Execute(cmdName, cmdArgs)
}

/*
 * syntax: PING
 */
type PingCmdExecutor struct{}

func (PingCmdExecutor) Execute(_ string, _ []*resp.RespValue) (*resp.RespValue, error) {
	return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "PONG"}, nil
}

/*
 * syntax: ECHO message
 */
type EchoCmdExecutor struct{}

func (EchoCmdExecutor) Execute(_ string, args []*resp.RespValue) (*resp.RespValue, error) {
	if len(args) == 0 {
		return nil, ErrInvalidArgs
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: args[0].BulkStr}, nil
}
