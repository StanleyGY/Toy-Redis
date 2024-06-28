package algo

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadixAddRemoveBasic(t *testing.T) {
	r := MakeRadixTree()
	r.Insert("abc", []string{"abc"})
	r.Insert("ab", []string{"ab"})
	r.Insert("abd", []string{"abd"})
	r.Insert("abcc", []string{"abcc"})
	r.Insert("abb", []string{"abb"})
	r.Visualize()

	r.Remove("abd")
	r.Remove("abc")
	r.Visualize()

	r.Remove("abcc")
	r.Visualize()

	assert.Equal(t, r.NumElems, 5)
}

func TestRadixAddRemoveBasic2(t *testing.T) {
	texts := []string{
		"CACADCAD",
		"CDCDECFE",
		"CDBBAAFD",
	}

	r := MakeRadixTree()

	for _, t := range texts {
		r.Insert(t, []string{t})
	}

	for _, str := range texts {
		r.Remove(str)
		r.Visualize()
	}
	assert.Equal(t, 0, len(r.Head.Edges))
}

func TestRadixAddRemoveScale(t *testing.T) {
	randString := func(n int) string {
		var letterRunes = []rune("ABCDEFGHIJKL")
		s := make([]rune, n)
		for i := range s {
			s[i] = letterRunes[rand.Intn(len(letterRunes))]
		}
		return string(s)
	}

	numElems := 500000
	texts := make([]string, 0)
	textsSeen := make(map[string]bool)
	for i := 0; i < numElems; i++ {
		newText := randString(8)

		existed := textsSeen[newText]
		if !existed {
			texts = append(texts, newText)
			textsSeen[newText] = true
		}
	}
	numElems = len(texts)
	t.Log("actual number of texts to be tested: ", numElems)

	r := MakeRadixTree()

	// Insert strings into radix tree
	for i := 0; i < numElems; i++ {
		assert.True(t, r.Insert(texts[i], []string{texts[i]}))
	}
	assert.Equal(t, numElems, r.NumElems)

	// Search for strings that do exist
	for i := 0; i < numElems; i++ {
		vals := r.Search(texts[i])
		assert.NotNil(t, vals)
	}

	// Search for strings that do not exist
	for i := 0; i < numElems; i++ {
		vals := r.Search(randString(15))
		assert.Nil(t, vals)
	}

	// Remove all keys
	for i := 0; i < numElems; i++ {
		assert.True(t, r.Remove(texts[i]))
	}
	assert.Equal(t, 0, len(r.Head.Edges))

	// Search for strings that do exist. Now they cannot be found
	for i := 0; i < numElems; i++ {
		vals := r.Search(texts[i])
		assert.Nil(t, vals)
	}
}

func TestRadixSearchByRange(t *testing.T) {
	r := MakeRadixTree()

	r.Insert("AA-3", []string{})
	r.Insert("BB-9", []string{})
	r.Insert("AC-2", []string{})
	r.Insert("CC-1", []string{})
	r.Insert("ZZ-1", []string{})

	res := r.SearchByRange("AA-45", "BB-9", math.MaxInt)
	assert.Equal(t, 2, len(res))

	res = r.SearchByRange("A-1", "Z-9", math.MaxInt)
	assert.Equal(t, 4, len(res))

	res = r.SearchByRange("A-1", "ZZ-9", math.MaxInt)
	assert.Equal(t, 5, len(res))
}
