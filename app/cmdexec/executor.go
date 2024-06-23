package cmdexec

import (
	"errors"
	"strings"

	"github.com/stanleygy/toy-redis/app/resp"
)

type CmdExecutor interface {
	Execute([]*resp.RespValue) (*resp.RespValue, error)
}

/*
 * syntax: PING [message]
 */
type PingCmdExecutor struct{}

func (PingCmdExecutor) Execute([]*resp.RespValue) (*resp.RespValue, error) {
	return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "PONG"}, nil
}

/*
 * syntax: ECHO message
 */
type EchoCmdExecutor struct{}

func (EchoCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	if len(args) == 0 {
		return nil, errors.New("echo: missing arg")
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: args[0].BulkStr}, nil
}

/*
 * syntax: SET key value [NX | XX] [GET] [EX seconds | PX milliseconds |
 *  EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
 */
type SetCmdExecutor struct{}

func (SetCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	key := args[0].BulkStr
	val := args[1].BulkStr
	db.KvStore[key] = val
	return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "OK"}, nil
}

/*
 * syntax: GET key
 */
type GetCmdExecutor struct{}

func (GetCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	key := args[0].BulkStr
	val, ok := db.KvStore[key]

	if ok {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: val}, nil
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
}

var CmdLookupTable = map[string]CmdExecutor{
	"COMMAND": &PingCmdExecutor{},
	"PING":    &PingCmdExecutor{},
	"ECHO":    &EchoCmdExecutor{},
	"SET":     &SetCmdExecutor{},
	"GET":     &GetCmdExecutor{},
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
