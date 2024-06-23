package cmdexec

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/resp"
)

type CmdExecutor interface {
	Execute([]*resp.RespValue) (*resp.RespValue, error)
}

/*
 * syntax: PING
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
 * syntax: SET key value [NX] [EX seconds | PX milliseconds]
 */
type SetCmdExecutor struct{}

func (e SetCmdExecutor) execute(key string, val string, nxFlag bool, expiry bool, ttl time.Time) bool {
	if nxFlag {
		val, found := db.KvStore[key]
		if found && val.WillExpire && val.ExpireTime.Compare(time.Now()) == -1 {
			delete(db.KvStore, key)
			found = false
		}
		if found {
			return false
		}
	}

	kvStoreVal := &KvStoreValue{
		Value: val,
	}
	if expiry {
		kvStoreVal.ExpireTime = ttl
		kvStoreVal.WillExpire = true
	}
	db.KvStore[key] = kvStoreVal
	return true
}

func (e SetCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	key := args[0].BulkStr
	val := args[1].BulkStr
	nxFlag := false
	expiry := false

	var ttl time.Time

	for i := 2; i < len(args); i++ {
		modifier := strings.ToUpper(args[i].BulkStr)

		if modifier == "NX" {
			// NX - only set the key if it does not already exist
			nxFlag = true
		} else if modifier == "PX" || modifier == "EX" {
			// PX milliseconds - set the specified expire time in ms (a positive integer)
			// EX milliseconds - set the specified expire time in secs (a positive integer)
			expireTime, err := strconv.Atoi(args[i+1].BulkStr)
			if err != nil {
				return nil, err
			}
			if expireTime < 0 {
				return nil, errors.New("ttl must be a positive integer")
			}

			expiry = true
			if modifier == "PX" {
				ttl = time.Now().Add(time.Millisecond * time.Duration(expireTime))
			} else {
				ttl = time.Now().Add(time.Second * time.Duration(expireTime))
			}

			i++
		}
	}
	if e.execute(key, val, nxFlag, expiry, ttl) {
		return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "OK"}, nil
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
}

/*
 * syntax: GET key
 */
type GetCmdExecutor struct{}

func (GetCmdExecutor) Execute(args []*resp.RespValue) (*resp.RespValue, error) {
	key := args[0].BulkStr
	val, found := db.KvStore[key]

	// Check expiry
	if found && val.WillExpire && val.ExpireTime.Compare(time.Now()) == -1 {
		delete(db.KvStore, key)
		found = false
	}
	if !found {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: val.Value}, nil
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
