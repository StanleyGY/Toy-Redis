package cmdexec

import (
	"fmt"
	"time"
)

type DictStoreValue struct {
	Value      string
	WillExpire bool
	ExpireTime time.Time
}

type StreamID struct {
	Ms  int64
	Seq uint
}

func (s StreamID) ToString() string {
	return fmt.Sprintf("%d-%d", s.Ms, s.Seq)
}

type Stream struct {
	Radix  *RadixTree
	LastId *StreamID
}

type RedisDb struct {
	DictStore      map[string]*DictStoreValue
	SortedSetStore map[string]*SkipList
	StreamStore    map[string]*Stream
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		DictStore:      make(map[string]*DictStoreValue),
		SortedSetStore: make(map[string]*SkipList),
		StreamStore:    make(map[string]*Stream),
	}
}
