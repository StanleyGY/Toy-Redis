package cmdexec

import (
	"fmt"
	"math"
	"math/rand"
)

const (
	SkipListDefaultMaxHeight = 36
)

type Node struct {
	Score     int
	Height    int
	Spans     []int
	Member    string
	PrevNodes []*Node
	NextNodes []*Node
}

func MakeNode(member string, score int, height int) *Node {
	node := &Node{
		Member:    member,
		Score:     score,
		Height:    height,
		Spans:     make([]int, height),
		NextNodes: make([]*Node, height),
		PrevNodes: make([]*Node, height),
	}
	for i := 0; i < height; i++ {
		node.Spans[i] = 1
	}
	return node
}

type SkipList struct {
	MemberMap map[string]*Node
	Rand      *rand.Rand
	NumElems  int
	Head      *Node
	Tail      *Node
}

func MakeSkipList(seed int64) *SkipList {
	headNode := MakeNode("head", math.MinInt, SkipListDefaultMaxHeight)
	tailNode := MakeNode("tail", math.MaxInt, 1)
	headNode.NextNodes[0] = tailNode
	tailNode.PrevNodes[0] = headNode
	tailNode.Spans[0] = 0

	return &SkipList{
		MemberMap: make(map[string]*Node),
		Head:      headNode,
		Tail:      tailNode,
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
	return l.Head.NextNodes[0]
}

func (l *SkipList) Back() *Node {
	if l.NumElems == 0 {
		return nil
	}
	return l.Tail.PrevNodes[0]
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

// findNodeAtLRange searches for the smallest node where `node.Score >= score`
func (l *SkipList) findNodeAtLeftRange(score int) *Node {
	if l.NumElems == 0 {
		return nil
	}
	if score <= l.Front().Score {
		return l.Front()
	}
	if score > l.Back().Score {
		return nil
	}

	var ans *Node
	h := l.Head.Height - 1
	curr := l.Head

	// Keep pushing to the right to increase `curr.Score` and decrease `ans.Score`
	for h >= 0 {
		next := curr.NextNodes[h]
		if next == nil {
			h--
		} else if next.Score < score {
			curr = next
		} else {
			// next.Score >= score
			ans = next
			h--
		}
	}
	return ans
}

// findNodeAtLRange searches for the biggest node where `node.Score <= score`
func (l *SkipList) findNodeAtRightRange(score int) *Node {
	// Idea: keep push to the right for a larger value
	if l.NumElems == 0 {
		return nil
	}
	if score >= l.Back().Score {
		return l.Back()
	}
	if score < l.Front().Score {
		return nil
	}

	var ans *Node
	h := l.Head.Height - 1
	curr := l.Head

	// Keep pushing to the right to increase `curr.Score` and increase `ans.Score`
	for h >= 0 {
		ans = curr
		next := curr.NextNodes[h]
		if next == nil {
			h--
		} else if next.Score <= score {
			curr = next
		} else {
			// next.Score > score
			ans = curr
			h--
		}
	}
	return ans
}

func (l *SkipList) CountByRange(min int, max int) int {
	if min > max {
		return 0
	}

	start := l.findNodeAtLeftRange(min)
	end := l.findNodeAtRightRange(max)

	if start == nil || end == nil {
		return 0
	}

	c := 0
	for curr := start; curr != end; curr = curr.NextNodes[0] {
		c++
	}
	c++
	return c
}

func (l *SkipList) FindByRange(min int, max int) []*Node {
	if min > max {
		return nil
	}

	start := l.findNodeAtLeftRange(min)
	end := l.findNodeAtRightRange(max)

	if start == nil || end == nil {
		return nil
	}

	res := make([]*Node, 0)
	for curr := start; curr != end; curr = curr.NextNodes[0] {
		res = append(res, curr)
	}
	res = append(res, end)
	return res
}

func (l *SkipList) GetRank(member string) (int, bool) {
	target, found := l.MemberMap[member]
	if !found {
		return 0, false
	}

	rank := 0
	h := l.Head.Height - 1
	curr := l.Head
	for curr != target {
		next := curr.NextNodes[h]
		if next == nil || next.Score > target.Score {
			h--
		} else {
			rank += curr.Spans[h]
			curr = next
		}
	}
	return rank, true
}

func (l *SkipList) FindByRank(rank int) *Node {
	if rank <= 0 || rank > l.NumElems {
		return nil
	}

	currRank := 0
	h := l.Head.Height - 1
	curr := l.Head

	for currRank != rank {
		next := curr.NextNodes[h]
		if currRank+curr.Spans[h] <= rank {
			currRank += curr.Spans[h]
			curr = next
		} else {
			h--
		}
	}

	return curr
}

func (l *SkipList) FindByRanks(start int, end int) []*Node {
	if start > end {
		return nil
	}

	startNode := l.FindByRank(start)
	endNode := l.FindByRank(end)

	if startNode == nil || endNode == nil {
		return nil
	}

	res := make([]*Node, 0)
	for curr := startNode; curr != endNode; curr = curr.NextNodes[0] {
		res = append(res, curr)
	}
	res = append(res, endNode)
	return res
}

func (l *SkipList) Remove(member string) bool {
	curr, found := l.MemberMap[member]
	if !found {
		return false
	}

	for i := 0; i < curr.Height; i++ {
		prev := curr.PrevNodes[i]
		next := curr.NextNodes[i]

		prev.NextNodes[i] = next
		if next != nil {
			next.PrevNodes[i] = prev
		}
	}

	l.NumElems--
	delete(l.MemberMap, member)
	return true
}

func (l *SkipList) findInsertionPos(score int, height int) []*Node {
	prevs := make([]*Node, height)

	h := l.Head.Height - 1
	curr := l.Head
	for h >= 0 {
		if score > curr.Score {
			// Track nodes at each level that comes before the current node
			if h < height {
				prevs[h] = curr
			}

			// Keep searching
			next := curr.NextNodes[h]
			if next == nil || score < next.Score {
				curr.Spans[h]++
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

	// Create a new node
	newHeight := l.getHeight()
	newNode := MakeNode(member, score, newHeight)

	l.NumElems++
	l.MemberMap[member] = newNode

	prevs := l.findInsertionPos(score, newHeight)

	for i := 0; i < newHeight; i++ {
		var prev *Node = prevs[i]
		var next *Node = prev.NextNodes[i]

		// Link with prev/next nodes
		prev.NextNodes[i] = newNode
		if next != nil {
			next.PrevNodes[i] = newNode
		}
		newNode.PrevNodes[i] = prev
		newNode.NextNodes[i] = next

		// Update spans for prev nodes
		prev.Spans[i]--
	}

	return true
}

func (l *SkipList) Visualize() {
	if l.NumElems == 0 {
		return
	}
	for curr := l.Head; curr != nil; curr = curr.NextNodes[0] {
		fmt.Printf("%s,%d ", curr.Member, curr.Score)
	}
	fmt.Println()
}

func (l *SkipList) VisualizeSpans() {
	if l.NumElems == 0 {
		return
	}

	fmt.Println()
	for i := SkipListDefaultMaxHeight - 1; i >= 0; i-- {
		for curr := l.Head; curr != nil; curr = curr.NextNodes[0] {
			if i >= len(curr.Spans) {
				fmt.Printf("    ")
			} else {
				fmt.Printf("%03d ", curr.Spans[i])
			}
		}
		fmt.Println()
	}
}
