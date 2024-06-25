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

func (l *SkipList) GetRank(member string) (*Node, int) {
	target, found := l.MemberMap[member]
	if !found {
		return nil, 0
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
	return curr, rank - 1
}

func (l *SkipList) FindByRank(rank int) *Node {
	if rank < 0 || rank >= l.NumElems {
		return nil
	}

	currRank := 0
	h := l.Head.Height - 1
	curr := l.Head

	for currRank != rank+1 {
		next := curr.NextNodes[h]
		if currRank+curr.Spans[h] <= rank+1 {
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
	if start < 0 {
		start = 0
	}
	if end >= l.NumElems {
		end = l.NumElems - 1
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
	target, found := l.MemberMap[member]
	if !found {
		return false
	}

	// Find the `prevs` that are above the height of  `target` node.
	// Can stop the search once it reaches the height of `target` node
	// then we'll use `target.PrevNodes[i]`
	prevs := make([]*Node, SkipListDefaultMaxHeight)
	h := l.Head.Height - 1
	curr := l.Head
	for h >= target.Height {
		next := curr.NextNodes[h]
		if next == nil || next.Score > target.Score {
			prevs[h] = curr
			h--
		} else {
			curr = next
		}
	}
	for i := 0; i < target.Height; i++ {
		prevs[i] = target.PrevNodes[i]
	}

	// Relink prev/next nodes
	for i := 0; i < target.Height; i++ {
		prev := target.PrevNodes[i]
		next := target.NextNodes[i]

		prev.NextNodes[i] = next
		if next != nil {
			next.PrevNodes[i] = prev
		}
	}

	// Update spans
	for i := 0; i < SkipListDefaultMaxHeight; i++ {
		if i < target.Height {
			prevs[i].Spans[i] += (target.Spans[i] - 1)
		} else {
			prevs[i].Spans[i]--
		}
	}

	l.NumElems--
	delete(l.MemberMap, member)
	return true
}

func (l *SkipList) findInsertionPos(score int, height int) ([]*Node, []int) {
	prevs := make([]*Node, height)
	prevSpans := make([]int, height)

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
				// No need to track the distance traversed for `h > height`
				if h >= height {
					curr.Spans[h]++
				}
				h--
			} else {
				// Update the distance traversed since prevs[i], i > h
				for i := h + 1; i < height; i++ {
					prevSpans[i] += curr.Spans[h]
				}
				curr = next
			}
		}
	}
	return prevs, prevSpans
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

	prevs, prevSpans := l.findInsertionPos(score, newHeight)

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
		newNode.Spans[i] = (prev.Spans[i] + 1) - (prevSpans[i] + 1)
		prev.Spans[i] = prevSpans[i] + 1
	}
	return true
}

func (l *SkipList) Visualize() {
	if l.NumElems == 0 {
		return
	}
	for curr := l.Head.NextNodes[0]; curr != l.Tail; curr = curr.NextNodes[0] {
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
		for curr := l.Head; curr != l.Tail; curr = curr.NextNodes[0] {
			if i >= len(curr.Spans) {
				fmt.Printf("    ")
			} else {
				fmt.Printf("%03d ", curr.Spans[i])
			}
		}
		fmt.Println()
	}
}
