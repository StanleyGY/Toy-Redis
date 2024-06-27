package cmdexec

import (
	"fmt"
	"sort"
)

type RaxNodeEdge struct {
	Label    byte   // Used for quickly checking if a node splitting should occur
	Prefix   string // The actual prefix associated with the edge
	DestNode *RaxNode
}

type RaxNode struct {
	From     *RaxNode
	FromEdge *RaxNodeEdge

	Edges    []*RaxNodeEdge
	HasValue bool
}

func (n *RaxNode) GetEdge(prefix string) *RaxNodeEdge {
	// Find the smallest index where `n.Edges[i].Label >= target[0]`
	pos := sort.Search(len(n.Edges), func(i int) bool {
		return n.Edges[i].Label >= prefix[0]
	})
	if pos < len(n.Edges) && n.Edges[pos].Label == prefix[0] {
		return n.Edges[pos]
	}
	return nil
}

func (n *RaxNode) DeleteEdge(prefix string) bool {
	// Find target index
	pos := sort.Search(len(n.Edges), func(i int) bool {
		return n.Edges[i].Label >= prefix[0]
	})
	if pos >= len(n.Edges) || n.Edges[pos].Prefix != prefix {
		return false
	}

	n.Edges = append(n.Edges[:pos], n.Edges[pos+1:]...)
	return true
}

func (n *RaxNode) AddEdge(prefix string, destNode *RaxNode) {
	e := &RaxNodeEdge{
		Label:    prefix[0],
		Prefix:   prefix,
		DestNode: destNode,
	}

	// Find insertion index
	pos := sort.Search(len(n.Edges), func(i int) bool {
		return n.Edges[i].Label >= prefix[0]
	})
	if len(n.Edges) > 0 && pos == len(n.Edges) && prefix[0] < n.Edges[0].Label {
		pos = 0
	}

	n.Edges = append(n.Edges, nil)
	copy(n.Edges[pos+1:], n.Edges[pos:])
	n.Edges[pos] = e

	// Track parent
	destNode.From = n
	destNode.FromEdge = e
}

func (n *RaxNode) Consolidate() {
	if len(n.Edges) != 1 || n.HasValue {
		return
	}

	// Consolidate current node with my only child node
	child := n.Edges[0].DestNode
	n.Edges = child.Edges
	n.HasValue = child.HasValue
	n.FromEdge.Prefix = n.FromEdge.Prefix + child.FromEdge.Prefix

	for _, edge := range child.Edges {
		edge.DestNode.From = n
	}
}

type Radix struct {
	Head     *RaxNode
	NumElems int
}

func MakeRadix() *Radix {
	return &Radix{
		Head: &RaxNode{
			Edges:    make([]*RaxNodeEdge, 0),
			HasValue: false,
		},
		NumElems: 0,
	}
}

func getLongestCommonPrefix(s1, s2 string) int {
	idx := 0
	for ; idx < len(s1) && idx < len(s2); idx++ {
		if s1[idx] != s2[idx] {
			break
		}
	}
	return idx - 1
}

func (r *Radix) Search(text string) bool {
	curr := r.Head
	for len(text) > 0 {
		edge := curr.GetEdge(text)
		if edge == nil {
			return false
		}

		commonPrefixIdx := getLongestCommonPrefix(edge.Prefix, text)
		if commonPrefixIdx+1 < len(edge.Prefix) {
			return false
		}

		text = text[commonPrefixIdx+1:]
		curr = edge.DestNode
	}
	return curr.HasValue
}

func (r *Radix) Remove(text string) bool {
	curr := r.Head
	for len(text) > 0 {
		edge := curr.GetEdge(text)

		if edge == nil {
			return false
		}

		commonPrefixIdx := getLongestCommonPrefix(edge.Prefix, text)
		if commonPrefixIdx+1 < len(edge.Prefix) {
			return false
		}

		text = text[commonPrefixIdx+1:]
		curr = edge.DestNode
	}
	if !curr.HasValue {
		return false
	}

	curr.HasValue = false

	if len(curr.Edges) == 1 && curr.From != nil {
		// If curr node has only one edge, consolidate with its child
		curr.Consolidate()
	} else if len(curr.Edges) == 0 && curr.From != nil {
		// If curr node has no children, remove this node from its parent
		prev := curr.From
		prev.DeleteEdge(curr.FromEdge.Prefix)

		// If parent node is a non-value node and has only one edge, consolidate with its child.
		// The invariant is that no two non-value nodes can be neighbors if each of them only has one edge
		if len(prev.Edges) == 1 && !prev.HasValue && prev.From != nil {
			prev.Consolidate()
		}
	}
	return true
}

func (r *Radix) Insert(text string) bool {
	curr := r.Head
	for {
		edge := curr.GetEdge(text)

		/*
			If this node `curr` doesn't have any edge that matches any character of `searchedPrefix`.
			create a new edge and add it to `curr`.

			Assume we have the following radix tree
				* - ABC - {*} - (other branches ....)

			Inserting another string "ABCD" will cause the tree to look like:
				* - ABC - {*} - (other branches ....)
						   \
							 D - [*]
		*/
		if edge == nil {
			curr.AddEdge(text, &RaxNode{HasValue: true})
			r.NumElems++
			return true
		}

		commonPrefixIdx := getLongestCommonPrefix(text, edge.Prefix)
		commonPrefix := text[:commonPrefixIdx+1]

		if commonPrefixIdx+1 == len(edge.Prefix) {
			// If the prefix fully matches and exhausts this edge's prefix
			text = text[commonPrefixIdx+1:]

			// If prefix to search is exhausted
			if len(text) == 0 {
				if curr.HasValue {
					return false
				}

				curr.HasValue = true
				r.NumElems++
				return true
			}

			curr = edge.DestNode
			continue
		}

		/*
			Handle node splitting. Assume we have the following radix tree:

					    DEF - [*]
					   /
			{*-ABC} - *
					   \
					    EF - [*]

			Case 1. To insert another string "ABQ", a node splitting will cause the tree to look like this:

					          DEF - [*]
					         /
			{*-AB} - * - C - *
					 |        \
					 Q         EF - [*]
					 |
					[*]

			Case 2. To insert another string "AB", a node splitting will cause the tree to look like this:

					          * - DEF - [*]
				 	         /
			{*-AB} - [*] - C
					         \
					          * - EF - [*]
		*/
		if commonPrefixIdx+1 == len(text) && commonPrefixIdx+1 < len(edge.Prefix) {
			// Handle case 2
			newNode := &RaxNode{HasValue: true}
			newNode.AddEdge(edge.Prefix[commonPrefixIdx+1:], edge.DestNode)
			curr.DeleteEdge(edge.Prefix)
			curr.AddEdge(commonPrefix, newNode)
			r.NumElems++
			return true
		}

		// Handle case 1
		splitNode := &RaxNode{}
		splitNode.AddEdge(edge.Prefix[commonPrefixIdx+1:], edge.DestNode)
		splitNode.AddEdge(text[commonPrefixIdx+1:], &RaxNode{
			Edges:    nil,
			HasValue: true,
		})
		curr.DeleteEdge(edge.Prefix)
		curr.AddEdge(commonPrefix, splitNode)
		r.NumElems++
		return true
	}
}

func (r *Radix) visualizeDFS(n *RaxNode, level int) {
	for _, edge := range n.Edges {
		for i := 0; i < level; i++ {
			fmt.Printf("  ")
		}
		if edge.DestNode.HasValue {
			fmt.Printf("[%s]\n", edge.Prefix)
		} else {
			fmt.Println(edge.Prefix)
		}
		r.visualizeDFS(edge.DestNode, level+1)
	}
}

func (r *Radix) Visualize() {
	fmt.Println("=============== Visualize ===============")
	r.visualizeDFS(r.Head, 0)
	fmt.Println()
}
