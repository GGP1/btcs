package merkle

import (
	"crypto/sha256"
)

// Tree represent a Merkle tree.
type Tree struct {
	Root *node
}

// node represent a Merkle tree node.
type node struct {
	left  *node
	right *node
	Hash  []byte
}

// NewTree creates a new Merkle tree from a sequence of data.
func NewTree(data [][]byte) *Tree {
	nodes := make([]*node, 0, len(data))

	for _, dt := range data {
		nodes = append(nodes, newNode(nil, nil, dt))
	}

	for len(nodes) > 1 {
		newLevel := make([]*node, 0, len(nodes)/2)

		for i := 0; i < len(nodes); i += 2 {
			j := i + 1
			if j == len(nodes) {
				j--
			}
			node := newNode(nodes[i], nodes[j], nil)
			newLevel = append(newLevel, node)
		}

		nodes = newLevel
	}

	return &Tree{nodes[0]}
}

// newNode creates a new Merkle tree node.
func newNode(left, right *node, data []byte) *node {
	var hash [sha256.Size]byte
	if left == nil && right == nil {
		hash = sha256.Sum256(data)
	} else {
		prevHashes := append(left.Hash, right.Hash...)
		hash = sha256.Sum256(prevHashes)
	}

	return &node{
		Hash:  hash[:],
		left:  left,
		right: right,
	}
}
