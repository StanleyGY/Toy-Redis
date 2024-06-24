package cmdexec

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/resp"
)

/*
 * syntax: SET key value [NX] [EX seconds | PX milliseconds]
 */
type SetCmdExecutor struct{}

func (e SetCmdExecutor) parseSetCmdArgs(args []*resp.RespValue, key *string, val *string, nxFlag *bool, expiry *bool, ttl *time.Time) error {
	*key = args[0].BulkStr
	*val = args[1].BulkStr

	for i := 2; i < len(args); i++ {
		modifier := strings.ToUpper(args[i].BulkStr)

		if modifier == "NX" {
			// NX - only set the key if it does not already exist
			*nxFlag = true
		} else if modifier == "PX" || modifier == "EX" {
			// PX milliseconds - set the specified expire time in ms (a positive integer)
			// EX milliseconds - set the specified expire time in secs (a positive integer)
			expireTime, err := strconv.Atoi(args[i+1].BulkStr)
			if err != nil {
				return err
			}
			if expireTime < 0 {
				return errors.New("ttl must be a positive integer")
			}

			*expiry = true
			if modifier == "PX" {
				*ttl = time.Now().Add(time.Millisecond * time.Duration(expireTime))
			} else {
				*ttl = time.Now().Add(time.Second * time.Duration(expireTime))
			}
			i++
		}
	}
	return nil
}

func (e SetCmdExecutor) doesKeyExistOrUnexpire(key string) bool {
	val, found := db.KvStore[key]
	if found && val.WillExpire && val.ExpireTime.Compare(time.Now()) == -1 {
		delete(db.KvStore, key)
		found = false
	}
	return found
}

func (e SetCmdExecutor) set(key string, val string, nxFlag bool, expiry bool, ttl time.Time) bool {
	if nxFlag && e.doesKeyExistOrUnexpire(key) {
		return false
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

func (e SetCmdExecutor) executeSetCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	var (
		key    string
		val    string
		nxFlag bool = false
		expiry bool = false
		ttl    time.Time
	)
	e.parseSetCmdArgs(cmdArgs, &key, &val, &nxFlag, &expiry, &ttl)
	if e.set(key, val, nxFlag, expiry, ttl) {
		return &resp.RespValue{DataType: resp.TypeSimpleStrings, SimpleStr: "OK"}, nil
	}
	return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
}

func (e SetCmdExecutor) executeGetCmd(cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	key := cmdArgs[0].BulkStr
	if !e.doesKeyExistOrUnexpire(key) {
		return &resp.RespValue{DataType: resp.TypeBulkStrings, IsNullBulkStr: true}, nil
	}
	val := db.KvStore[key]
	return &resp.RespValue{DataType: resp.TypeBulkStrings, BulkStr: val.Value}, nil
}

func (e SetCmdExecutor) Execute(cmdName string, cmdArgs []*resp.RespValue) (*resp.RespValue, error) {
	if cmdName == "SET" {
		return e.executeSetCmd(cmdArgs)
	} else {
		return e.executeGetCmd(cmdArgs)
	}
}
