package cmdexec

import (
	"strconv"
	"strings"
	"time"

	"github.com/stanleygy/toy-redis/app/event"
	"github.com/stanleygy/toy-redis/app/resp"
)

/*
 * syntax: SET key value [NX] [EX seconds | PX milliseconds]
 */
type setCmdExecutor struct{}

func (e setCmdExecutor) parseSetCmdArgs(args []*resp.RespValue, key *string, val *string, nxFlag *bool, expiry *bool, ttl *time.Time) error {
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
				return ErrInvalidArgs
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

func (e setCmdExecutor) doesKeyExistOrUnexpire(key string) bool {
	val, found := db.DictStore[key]
	if found && val.WillExpire && val.ExpireTime.Compare(time.Now()) == -1 {
		delete(db.DictStore, key)
		found = false
	}
	return found
}

func (e setCmdExecutor) set(key string, val string, nxFlag bool, expiry bool, ttl time.Time) bool {
	if nxFlag && e.doesKeyExistOrUnexpire(key) {
		return false
	}
	kvStoreVal := &DictStoreValue{
		Value: val,
	}
	if expiry {
		kvStoreVal.ExpireTime = ttl
		kvStoreVal.WillExpire = true
	}
	db.DictStore[key] = kvStoreVal
	return true
}

func (e setCmdExecutor) executeSetCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	var (
		key    string
		val    string
		nxFlag bool = false
		expiry bool = false
		ttl    time.Time
	)
	e.parseSetCmdArgs(cmdArgs, &key, &val, &nxFlag, &expiry, &ttl)
	if !e.set(key, val, nxFlag, expiry, ttl) {
		event.AddNullBulkStringReplyEvent(c)
		return
	}
	event.AddSimpleStringReplyEvent(c, "OK")
}

func (e setCmdExecutor) executeGetCmd(c *event.ClientInfo, cmdArgs []*resp.RespValue) {
	key := cmdArgs[0].BulkStr
	if !e.doesKeyExistOrUnexpire(key) {
		event.AddNullBulkStringReplyEvent(c)
		return
	}
	val := db.DictStore[key]
	event.AddBulkStringReplyEvent(c, val.Value)
}

func (e setCmdExecutor) Execute(c *event.ClientInfo, cmdName string, cmdArgs []*resp.RespValue) {
	switch cmdName {
	case "SET":
		e.executeSetCmd(c, cmdArgs)
	case "GET":
		e.executeGetCmd(c, cmdArgs)
	}
}
