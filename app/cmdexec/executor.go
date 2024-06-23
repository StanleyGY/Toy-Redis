package cmdexec

import (
	"errors"
	"strings"

	"github.com/stanleygy/toy-redis/app/resp"
)

type CmdExecutor interface {
	Execute([]*resp.RespValue) (*resp.RespValue, error)
}

type PingCmdExecutor struct{}

func (PingCmdExecutor) Execute([]*resp.RespValue) (*resp.RespValue, error) {
	return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "PONG"}, nil
}

type EchoCmdExecutor struct{}

func (EchoCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	if len(args) == 0 {
		return nil, errors.New("echo: missing arg")
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: args[0].BulkStr}, nil
}

var CmdLookupTable = map[string]CmdExecutor{
	"COMMAND": &PingCmdExecutor{},
	"PING":    &PingCmdExecutor{},
	"ECHO":    &EchoCmdExecutor{},
}

func Execute(val *resp.RespValue) (*resp.RespValue, error) {
	cmdName := val.Array[0].BulkStr
	cmd := CmdLookupTable[strings.ToUpper(cmdName)]
	if cmd == nil {
		return nil, errors.New("failed to look up command")
	}

	cmdArgs := val.Array[1:len(val.Array)]
	return cmd.Execute(cmdArgs)
}
