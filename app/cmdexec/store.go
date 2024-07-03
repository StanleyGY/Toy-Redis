package cmdexec

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
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
	Seq int
}

func (s *StreamID) Incr() error {
	// Increment the stream ID by one
	if s.Seq == math.MaxInt {
		return ErrOverflow
	}
	s.Seq++
	return nil
}

func (s StreamID) ToString() string {
	return fmt.Sprintf("%d-%d", s.Ms, s.Seq)
}

func ParseStreamID(id string) (*StreamID, error) {
	var ms int64
	var seq int
	var err error

	parts := strings.Split(id, "-")
	if len(parts) < 1 || len(parts) > 2 {
		return nil, errors.New("failed to parse stream id")
	}

	ms, err = strconv.ParseInt(parts[0], 10, 64)
	if len(parts) == 1 {
		return &StreamID{Ms: ms}, err
	}

	seq, err = strconv.Atoi(parts[1])
	return &StreamID{Ms: ms, Seq: seq}, err
}

type Stream struct {
	Radix  *algo.RadixTree
	LastId *StreamID
}

type GeoStoreValue struct {
	Coord algo.GeoCoord
	Hash  string
}

type RedisDb struct {
	DictStore      map[string]*DictStoreValue
	SortedSetStore map[string]*algo.SkipList
	StreamStore    map[string]*Stream
	GeoStore       map[string]map[string]*GeoStoreValue
}

var db *RedisDb

func InitRedisDb() {
	db = &RedisDb{
		DictStore:      make(map[string]*DictStoreValue),
		SortedSetStore: make(map[string]*algo.SkipList),
		StreamStore:    make(map[string]*Stream),
		GeoStore:       make(map[string]map[string]*GeoStoreValue),
	}
}
