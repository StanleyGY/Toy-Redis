package cmdexec

import "time"

type DictStoreValue struct {
	Value      string
	WillExpire bool
	ExpireTime time.Time
}

type RedisDb struct {
	DictStore      map[string]*DictStoreValue
	SortedSetStore map[string]*SkipList
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		DictStore:      make(map[string]*DictStoreValue),
		SortedSetStore: make(map[string]*SkipList),
	}
}
