package algo

import (
	"fmt"
	"sort"
)

type RadixSearchResult struct {
	Node *RaxNode
	Id   string
}

type RaxNodeEdge struct {
	Label    byte   // Used for quickly checking if a node splitting should occur
	Prefix   string // The actual prefix associated with the edge
	DestNode *RaxNode
}

type RaxNode struct {
	From     *RaxNode
	FromEdge *RaxNodeEdge

	Edges  []*RaxNodeEdge
	Values []string
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
	if len(n.Edges) != 1 || n.Values != nil {
		return
	}

	// Consolidate current node with my only child node
	child := n.Edges[0].DestNode
	n.Edges = child.Edges
	n.Values = child.Values
	n.FromEdge.Prefix = n.FromEdge.Prefix + child.FromEdge.Prefix

	for _, edge := range child.Edges {
		edge.DestNode.From = n
	}
}

type RadixTree struct {
	Head     *RaxNode
	NumElems int
}

func MakeRadixTree() *RadixTree {
	return &RadixTree{
		Head: &RaxNode{
			Edges:  make([]*RaxNodeEdge, 0),
			Values: nil,
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

func (r *RadixTree) searchByRange(startId string, endId string, limit int, curr *RaxNode, currPrefix string, results *[]*RadixSearchResult) {
	if curr == nil {
		return
	}

	if curr.Values != nil {
		*results = append(*results, &RadixSearchResult{
			Id:   currPrefix,
			Node: curr,
		})
	}
	if len(*results) >= limit {
		return
	}

	for _, edge := range curr.Edges {
		var (
			startIdPrefix string = startId
			endIdPrefix   string = endId
		)
		nextPrefix := currPrefix + edge.Prefix
		if len(nextPrefix) < len(startId) {
			startIdPrefix = startId[:len(nextPrefix)]
		}
		if len(nextPrefix) < len(endId) {
			endIdPrefix = endId[:len(nextPrefix)]
		}

		if nextPrefix < startIdPrefix {
			continue
		}
		if nextPrefix > endIdPrefix {
			break
		}
		r.searchByRange(startId, endId, limit, edge.DestNode, nextPrefix, results)
	}
}

func (r *RadixTree) SearchByRange(startId string, endId string, limit int) []*RadixSearchResult {
	// Search range is inclusive at both ends
	var results []*RadixSearchResult
	r.searchByRange(startId, endId, limit, r.Head, "", &results)
	return results
}

func (r *RadixTree) searchNode(id string) *RaxNode {
	curr := r.Head
	for len(id) > 0 {
		edge := curr.GetEdge(id)
		if edge == nil {
			return nil
		}

		commonPrefixIdx := getLongestCommonPrefix(edge.Prefix, id)
		if commonPrefixIdx+1 < len(edge.Prefix) {
			return nil
		}

		id = id[commonPrefixIdx+1:]
		curr = edge.DestNode
	}
	return curr
}

func (r *RadixTree) Search(id string) []string {
	curr := r.searchNode(id)
	if curr.Values == nil {
		return nil
	}
	return curr.Values
}

func (r *RadixTree) Remove(id string) bool {
	curr := r.searchNode(id)
	if curr.Values == nil {
		// Nothing to remove if this not a value node
		return false
	}

	curr.Values = nil

	if len(curr.Edges) == 1 && curr.From != nil {
		// If curr node has only one edge, consolidate with its child
		curr.Consolidate()
	} else if len(curr.Edges) == 0 && curr.From != nil {
		// If curr node has no children, remove this node from its parent
		prev := curr.From
		prev.DeleteEdge(curr.FromEdge.Prefix)

		// If parent node is a non-value node and has only one edge, consolidate with its child.
		// The invariant is that no two non-value nodes can be neighbors if both of them has at most one edge.
		if len(prev.Edges) == 1 && prev.Values == nil && prev.From != nil {
			prev.Consolidate()
		}
	}
	return true
}

func (r *RadixTree) Insert(id string, values []string) bool {
	curr := r.Head
	for {
		edge := curr.GetEdge(id)

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
			curr.AddEdge(id, &RaxNode{Values: values})
			r.NumElems++
			return true
		}

		commonPrefixIdx := getLongestCommonPrefix(id, edge.Prefix)
		commonPrefix := id[:commonPrefixIdx+1]

		if commonPrefixIdx+1 == len(edge.Prefix) {
			// If the prefix fully matches and exhausts this edge's prefix
			id = id[commonPrefixIdx+1:]

			// If prefix to search is exhausted
			if len(id) == 0 {
				if curr.Values != nil {
					curr.Values = values
					return false
				}

				curr.Values = values
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
		if commonPrefixIdx+1 == len(id) && commonPrefixIdx+1 < len(edge.Prefix) {
			// Handle case 2
			newNode := &RaxNode{Values: values}
			newNode.AddEdge(edge.Prefix[commonPrefixIdx+1:], edge.DestNode)
			curr.DeleteEdge(edge.Prefix)
			curr.AddEdge(commonPrefix, newNode)
			r.NumElems++
			return true
		}

		// Handle case 1
		splitNode := &RaxNode{}
		splitNode.AddEdge(edge.Prefix[commonPrefixIdx+1:], edge.DestNode)
		splitNode.AddEdge(id[commonPrefixIdx+1:], &RaxNode{Values: values})
		curr.DeleteEdge(edge.Prefix)
		curr.AddEdge(commonPrefix, splitNode)
		r.NumElems++
		return true
	}
}

func (r *RadixTree) visualizeDFS(n *RaxNode, level int) {
	for _, edge := range n.Edges {
		for i := 0; i < level; i++ {
			fmt.Printf("  ")
		}
		if edge.DestNode.Values != nil {
			fmt.Printf("[%s]\n", edge.Prefix)
		} else {
			fmt.Println(edge.Prefix)
		}
		r.visualizeDFS(edge.DestNode, level+1)
	}
}

func (r *RadixTree) Visualize() {
	fmt.Println("=============== Visualize ===============")
	r.visualizeDFS(r.Head, 0)
	fmt.Println()
}
