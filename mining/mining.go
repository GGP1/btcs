package mining

import (
	"github.com/GGP1/btcs/block"
)

// Miner provides facilities for solving blocks.
type Miner interface {
	Mine(lastBlock *block.Block) (block.Block, error)
}
