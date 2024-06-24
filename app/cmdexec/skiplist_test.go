package cmdexec

import (
	"math/rand"
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

func TestAddRemove(t *testing.T) {
	numElems := 100
	members := make(map[string]int)

	for i := 0; i < numElems; i++ {
		m := randString(20)
		members[m] = i
	}

	sl := MakeSkipList(time.Now().Unix())

	// Check `Add`
	for m, score := range members {
		sl.Add(m, score)
	}
	assert.Equal(t, numElems, sl.Size())

	// Check `GetScore`
	for m, score := range members {
		s := sl.GetScore(m)
		assert.Equal(t, score, s)
	}

	// Intrusively check `Add`
	idx := 0
	for curr := sl.Head[0]; curr != nil; curr = curr.NextNodes[0] {
		assert.Equal(t, curr.Score, idx)
		idx++
	}

	// Check `Remove`
	for m := range members {
		sl.Remove(m)
	}
	assert.Zero(t, sl.Size())
}
