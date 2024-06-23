package cmdexec

import "time"

type KvStoreValue struct {
	Value      string
	WillExpire bool
	ExpireTime time.Time
}

type RedisDb struct {
	KvStore map[string]*KvStoreValue
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		KvStore: make(map[string]*KvStoreValue),
	}
}
