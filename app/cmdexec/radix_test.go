package cmdexec

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadixAddRemoveBasic(t *testing.T) {
	r := MakeRadix()
	r.Insert("abc")
	r.Insert("ab")
	r.Insert("abd")
	r.Insert("abcc")
	r.Insert("abb")
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

	r := MakeRadix()

	for _, t := range texts {
		r.Insert(t)
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

	r := MakeRadix()

	// Insert strings into radix tree
	for i := 0; i < numElems; i++ {
		assert.True(t, r.Insert(texts[i]))
	}
	assert.Equal(t, numElems, r.NumElems)

	// Search for strings that do exist
	for i := 0; i < numElems; i++ {
		assert.True(t, r.Search(texts[i]))
	}

	// Search for strings that do not exist
	for i := 0; i < numElems; i++ {
		assert.False(t, r.Search(randString(15)))
	}

	// Remove all keys
	for i := 0; i < numElems; i++ {
		assert.True(t, r.Remove(texts[i]))
	}
	assert.Equal(t, 0, len(r.Head.Edges))

	// Search for strings that do exist. Now they cannot be found
	for i := 0; i < numElems; i++ {
		assert.False(t, r.Search(texts[i]))
	}
}
