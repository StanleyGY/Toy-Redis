package cmdexec

import (
	"fmt"
	"math/rand"
)

const (
	SkipListDefaultMaxHeight = 32
)

type Node struct {
	Score     int
	Member    string
	PrevNodes []*Node
	NextNodes []*Node
}

func (n Node) GetHeight() int {
	return len(n.NextNodes)
}

type SkipList struct {
	MemberMap map[string]*Node
	Rand      *rand.Rand
	NumElems  int
	Head      []*Node
	Tail      *Node
}

func MakeSkipList(seed int64) *SkipList {
	return &SkipList{
		MemberMap: make(map[string]*Node),
		Head:      make([]*Node, SkipListDefaultMaxHeight),
		Rand:      rand.New(rand.NewSource(seed)),
	}
}

func (l *SkipList) getHeight() int {
	// Node with height 1 - probability: 100%
	// Node with height 2 - probability: 50%
	// Node with height 3 - probability: 25%
	// ...
	height := 1
	for ; height < SkipListDefaultMaxHeight; height++ {
		if l.Rand.Int31() > (1 << 30) {
			break
		}
	}
	return height
}

func (l *SkipList) Front() *Node {
	if l.NumElems == 0 {
		return nil
	}
	return l.Head[0]
}

func (l *SkipList) Back() *Node {
	if l.NumElems == 0 {
		return nil
	}
	return l.Tail
}

func (l *SkipList) Size() int {
	return l.NumElems
}

func (l *SkipList) GetScore(member string) int {
	val := l.MemberMap[member]
	if val != nil {
		return val.Score
	}
	return 0
}

func (l *SkipList) Remove(member string) bool {
	curr, found := l.MemberMap[member]
	if !found {
		return false
	}

	for i := 0; i < curr.GetHeight(); i++ {
		prev := curr.PrevNodes[i]
		next := curr.NextNodes[i]

		if l.Head[i] == curr {
			l.Head[i] = next
		}
		if prev != nil {
			prev.NextNodes[i] = next
		}
		if next != nil {
			next.PrevNodes[i] = prev
		}
	}
	if l.Tail == curr {
		l.Tail = curr.PrevNodes[0]
	}

	l.NumElems--
	delete(l.MemberMap, member)
	return true
}

func (l *SkipList) findInsertion(score int, height int) []*Node {
	prevs := make([]*Node, height)

	var curr *Node
	var h int

	for h = len(l.Head) - 1; h >= 0; h-- {
		if l.Head[h] != nil && score >= l.Head[h].Score {
			break
		}
	}

	if h < 0 {
		return prevs
	}

	curr = l.Head[h]
	for h >= 0 {
		if score > curr.Score {
			// Track nodes at each level that comes before the current node
			if h < height {
				prevs[h] = curr
			}

			// Keep searching
			next := curr.NextNodes[h]
			if next == nil || score < next.Score {
				h--
			} else {
				curr = next
			}
		}
	}
	return prevs
}

func (l *SkipList) Add(member string, score int, insertOnly bool) bool {
	node, found := l.MemberMap[member]
	if found {
		if insertOnly || score == node.Score {
			return false
		}
		l.Remove(member)
		delete(l.MemberMap, member)
		l.Add(member, score, true)
		return true
	}

	newHeight := l.getHeight()
	newNode := &Node{
		Member:    member,
		Score:     score,
		PrevNodes: make([]*Node, newHeight),
		NextNodes: make([]*Node, newHeight),
	}

	// fmt.Printf("adding node: %s, score: %d, height: %d\n", member, score, newHeight)

	l.NumElems++
	l.MemberMap[member] = newNode

	if l.NumElems == 1 {
		for i := 0; i < newHeight; i++ {
			l.Head[i] = newNode
		}
		l.Tail = newNode
		return true
	}

	// Find insertion point
	prevs := l.findInsertion(score, newHeight)

	// Link prev/next nodes together
	for i := 0; i < newHeight; i++ {
		var next *Node

		if prevs[i] == nil {
			if l.Head[i] != nil {
				next = l.Head[i]
			}
			l.Head[i] = newNode
		} else {
			next = prevs[i].NextNodes[i]
			prevs[i].NextNodes[i] = newNode
		}

		if next != nil {
			next.PrevNodes[i] = newNode
		}

		newNode.PrevNodes[i] = prevs[i]
		newNode.NextNodes[i] = next
	}

	if prevs[0] == l.Tail {
		l.Tail = newNode
	}

	return true
}

func (l *SkipList) Visualize() {
	if l.NumElems == 0 {
		return
	}
	for curr := l.Head[0]; curr != nil; curr = curr.NextNodes[0] {
		fmt.Printf("%s,%d ", curr.Member, curr.Score)
	}
	fmt.Println()
}
