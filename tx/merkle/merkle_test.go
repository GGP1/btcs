package merkle

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
	}

	// Level 1
	n1 := newNode(nil, nil, data[0])
	n2 := newNode(nil, nil, data[1])
	n3 := newNode(nil, nil, data[2])
	n4 := newNode(nil, nil, data[2])

	// Level 2
	n5 := newNode(n1, n2, nil)
	n6 := newNode(n3, n4, nil)

	// Level 3
	n7 := newNode(n5, n6, nil)

	assert.Equal(
		t,
		"64b04b718d8b7c5b6fd17f7ec221945c034cfce3be4118da33244966150c4bd4",
		hex.EncodeToString(n5.Hash),
		"Level 1 hash 1 is correct",
	)
	assert.Equal(
		t,
		"08bd0d1426f87a78bfc2f0b13eccdf6f5b58dac6b37a7b9441c1a2fab415d76c",
		hex.EncodeToString(n6.Hash),
		"Level 1 hash 2 is correct",
	)
	assert.Equal(
		t,
		"4e3e44e55926330ab6c31892f980f8bfd1a6e910ff1ebc3f778211377f35227e",
		hex.EncodeToString(n7.Hash),
		"Root hash is correct",
	)
}

func TestNewTree(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
		[]byte("node4"),
		[]byte("node5"),
	}

	// Level 1
	n11 := newNode(nil, nil, data[0])
	n12 := newNode(nil, nil, data[1])
	n13 := newNode(nil, nil, data[2])
	n14 := newNode(nil, nil, data[3])
	n15 := newNode(nil, nil, data[4])

	// Level 2
	n21 := newNode(n11, n12, nil)
	n22 := newNode(n13, n14, nil)
	n23 := newNode(n15, n15, nil)

	// Level 3
	n31 := newNode(n21, n22, nil)
	n32 := newNode(n23, n23, nil)

	// Level 4
	n41 := newNode(n31, n32, nil)

	rootHash := hex.EncodeToString(n41.Hash)
	mTree := NewTree(data)

	assert.Equal(t, rootHash, hex.EncodeToString(mTree.Root.Hash), "Merkle tree root hash is correct")
}
