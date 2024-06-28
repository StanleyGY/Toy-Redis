package cmdexec

import (
	"fmt"
	"time"

	"github.com/stanleygy/toy-redis/app/algo"
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
	Radix  *algo.RadixTree
	LastId *StreamID
}

type RedisDb struct {
	DictStore      map[string]*DictStoreValue
	SortedSetStore map[string]*algo.SkipList
	StreamStore    map[string]*Stream
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		DictStore:      make(map[string]*DictStoreValue),
		SortedSetStore: make(map[string]*algo.SkipList),
		StreamStore:    make(map[string]*Stream),
	}
}
