package cmdexec

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randString(n int) string {
	var letterRunes = []rune("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, n)
	for i := range s {
		s[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(s)
}

func TestAddRemoveScale(t *testing.T) {
	numElems := 100000
	members := make(map[string]int)

	for i := 0; i < numElems; i++ {
		m := randString(20)
		members[m] = i
	}

	sl := MakeSkipList(time.Now().Unix())

	// Check `Add`
	for m, score := range members {
		sl.Add(m, score, false)
	}
	assert.Equal(t, numElems, sl.Size())

	// Check `GetScore`
	for m, score := range members {
		s := sl.GetScore(m)
		assert.Equal(t, score, s)
	}

	// Intrusively check `Add`
	idx := 0
	for curr := sl.Head.NextNodes[0]; curr != sl.Tail; curr = curr.NextNodes[0] {
		assert.Equal(t, curr.Score, idx)
		idx++
	}

	// Check `Remove`
	for m := range members {
		sl.Remove(m)
	}
	assert.Zero(t, sl.Size())
}

func TestRangeBasic(t *testing.T) {
	sl := MakeSkipList(time.Now().Unix())

	sl.Add("1", 1, true)
	sl.Add("3", 3, true)
	sl.Add("5", 5, true)
	sl.Add("8", 8, true)
	sl.Add("12", 12, true)
	sl.Add("14", 14, true)

	// Test where ranges include some nodes
	assert.Equal(t, 3, sl.CountByRange(4, 13))
	assert.Equal(t, 1, sl.CountByRange(1, 1))
	assert.Equal(t, 1, sl.CountByRange(14, 14))

	nodes := sl.FindByRange(5, 12)
	assert.Equal(t, 5, nodes[0].Score)
	assert.Equal(t, 8, nodes[1].Score)
	assert.Equal(t, 12, nodes[2].Score)

	// Test where ranges include all nodes
	assert.Equal(t, 6, sl.CountByRange(1, 100))

	// Test where ranges are invalid
	assert.Equal(t, 0, sl.CountByRange(15, 100))
	assert.Equal(t, 0, sl.CountByRange(-15, 0))
	assert.Equal(t, 0, sl.CountByRange(15, 14))
}

func TestRankBasic(t *testing.T) {
	sl := MakeSkipList(1)

	sl.Add("1", 1, true)
	sl.Add("3", 3, true)
	sl.Add("8", 8, true)
	sl.Add("5", 5, true)
	sl.Add("12", 12, true)
	sl.Add("14", 14, true)

	// Test GetRank
	n, r := sl.GetRank("5")
	assert.NotNil(t, n)
	assert.Equal(t, 2, r)

	n, r = sl.GetRank("1")
	assert.NotNil(t, n)
	assert.Equal(t, 0, r)

	n, r = sl.GetRank("14")
	assert.NotNil(t, n)
	assert.Equal(t, 5, r)

	n, _ = sl.GetRank("999")
	assert.Nil(t, n)

	// Test FindByRank
	n = sl.FindByRank(0)
	assert.Equal(t, "1", n.Member)
	n = sl.FindByRank(5)
	assert.Equal(t, "14", n.Member)
	n = sl.FindByRank(6)
	assert.Nil(t, n)
	n = sl.FindByRank(10)
	assert.Nil(t, n)

	// Test FindByRanks
	nodes := sl.FindByRanks(2, 4)
	assert.Equal(t, 3, len(nodes))

	// Test Remove
	sl.Remove("8")
	nodes = sl.FindByRanks(2, 4)
	assert.Equal(t, 3, len(nodes))

	sl.Remove("1")
	n = sl.FindByRank(0)
	assert.Equal(t, "3", n.Member)

	sl.Remove("5")
	n = sl.FindByRank(1)
	assert.Equal(t, "12", n.Member)
}

func TestRankScale(t *testing.T) {
	numElems := 100000

	members := make([]int, numElems)
	for i := 0; i < numElems; i++ {
		members[i] = i
	}
	rand.Shuffle(numElems, func(i, j int) { members[i], members[j] = members[j], members[i] })

	sl := MakeSkipList(9)
	for i := 0; i < numElems; i++ {
		sl.Add(fmt.Sprintf("%03d", members[i]), members[i], true)
	}

	numRemoved := numElems / 3
	for i := 0; i < numRemoved; i++ {
		pos := rand.Intn(numElems)
		sl.Remove(fmt.Sprintf("%03d", members[pos]))
		members = append(members[:pos], members[pos+1:]...)
		numElems--
	}

	sort.Ints(members)
	for i := 0; i < numElems; i++ {
		n := sl.FindByRank(i)
		assert.Equal(t, fmt.Sprintf("%03d", members[i]), n.Member)
	}
}
