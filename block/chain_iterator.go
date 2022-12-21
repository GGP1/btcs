package block

import (
	"github.com/GGP1/btcs/encoding/gob"

	bolt "go.etcd.io/bbolt"
)

// ChainIterator is used to iterate over blockchain blocks
type ChainIterator struct {
	db          *bolt.DB
	currentHash []byte
}

// BlocksCount returns the number of blocks stored in the databse.
func (i *ChainIterator) BlocksCount() (int, error) {
	var count int

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		count = b.Stats().KeyN
		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, err
}

// ForEach runs the f function on each iteration of the blockchain blocks.
func (i *ChainIterator) ForEach(f func(block Block) error) error {
	tx, err := i.db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	b := tx.Bucket([]byte(blocksBucket))
	c := b.Cursor()

	for k, _ := c.Last(); k != nil; k, _ = c.Prev() {
		encodedBlock := b.Get(i.currentHash)
		block, err := gob.Decode[Block](encodedBlock)
		if err != nil {
			return err
		}

		if err := f(block); err != nil {
			return err
		}

		i.currentHash = block.PrevBlockHash
		if i.currentHash == nil {
			break
		}
	}

	return nil
}

// Next returns the next block starting from the tip.
func (i *ChainIterator) Next(tx *bolt.Tx) (Block, error) {
	if len(i.currentHash) == 0 {
		return Block{}, nil
	}

	b := tx.Bucket([]byte(blocksBucket))
	encodedBlock := b.Get(i.currentHash)

	block, err := gob.Decode[Block](encodedBlock)
	if err != nil {
		return Block{}, err
	}
	i.currentHash = block.PrevBlockHash

	return block, nil
}
