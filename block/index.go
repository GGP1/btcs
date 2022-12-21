package block

var blockIndex = index{blocks: map[int32]int64{}}

type index struct {
	// map[height]timestamp
	blocks map[int32]int64
}

// TODO: in btcs the index contains the total work in the chain so it can figure out
// which node to connect to (shall we do the same? should rebalance index after it)
// type node struct {
// 	parent    *node
// 	work      *big.Int
// 	timestamp int64
// 	height    int32
// }

// addNode adds a node to the index.
func (i *index) addNode(height int32, timestamp int64) {
	i.blocks[height] = timestamp
}

// nodeTimestamp returns the timestamp of a block at a certain height.
func (i *index) nodeTimestamp(height int32) int64 {
	return i.blocks[height]
}
