package block

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/tx"

	bolt "go.etcd.io/bbolt"
)

const (
	dbPath       = "blockchain.db"
	blocksBucket = "blocks"
)

var (
	// ErrBlockchainNotFound is thrown when the blockchain database file is not found.
	ErrBlockchainNotFound = errors.New("blockchain not found")
	errEmptyBlockchain    = errors.New("empty blockchain")

	lastHashKey = []byte("l")
)

// Chain allows to read/write the blockchain file.
type Chain struct {
	*bolt.DB
	// tip contains the last block hash
	tip []byte
}

// NewChain creates a new blockchain and adds the genesis block.
func NewChain() (*Chain, error) {
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		return nil, errors.New("blockchain already exists")
	}

	db, err := bolt.Open(dbPath, 0o600, nil)
	if err != nil {
		return nil, err
	}

	genesis, err := NewGenesis()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		encodedBlock, err := gob.Encode(genesis)
		if err != nil {
			return err
		}

		if err := b.Put(genesis.Hash, encodedBlock); err != nil {
			return err
		}

		return b.Put(lastHashKey, genesis.Hash)
	})
	if err != nil {
		return nil, err
	}

	return &Chain{
		tip: genesis.Hash,
		DB:  db,
	}, nil
}

// LoadChain reads the blockchain file and loads the tip of the chain.
//
// Call Close to release the Chain's associated resources when done.
func LoadChain() (*Chain, error) {
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrBlockchainNotFound
		}
		return nil, err
	}

	db, err := bolt.Open(dbPath, 0o600, nil)
	if err != nil {
		return nil, err
	}

	var tip []byte
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get(lastHashKey)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Chain{
		tip: tip,
		DB:  db,
	}, nil
}

// NewIterator returns a BlockchainIterat
func (c *Chain) NewIterator() *ChainIterator {
	return &ChainIterator{
		db:          c.DB,
		currentHash: c.tip,
	}
}

// AddBlock adds the block to the chain.
func (c *Chain) AddBlock(block Block) error {
	if !block.IsValid() {
		return errors.New("invalid block")
	}

	for _, tx := range block.Transactions {
		if err := c.VerifyTx(tx); err != nil {
			return err
		}
	}

	return c.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if exists := b.Get(block.Hash); exists != nil {
			return errors.New("block already exists")
		}

		blockData, err := gob.Encode(block)
		if err != nil {
			return err
		}

		if err := b.Put(block.Hash, blockData); err != nil {
			return err
		}

		if err := b.Put(lastHashKey, block.Hash); err != nil {
			return err
		}

		c.tip = block.Hash
		return nil
	})
}

// BestHeight returns the height of the latest block.
func (c *Chain) BestHeight() (int32, error) {
	block, err := c.LastBlock()
	if err != nil {
		// If the blockchain has no blocks (new validator blockchain)
		// return -1 so nodes with at least one block send us it
		if err != errEmptyBlockchain {
			return 0, err
		}
		return -1, nil
	}

	return block.Height, nil
}

// Block finds a block by its hash and returns it
func (c *Chain) Block(hash []byte) (Block, error) {
	var block Block

	err := c.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		blockData := b.Get(hash)
		if blockData == nil {
			return errors.New("block not found")
		}

		decodedBlock, err := gob.Decode[Block](blockData)
		if err != nil {
			return err
		}

		block = decodedBlock
		return nil
	})
	if err != nil {
		return Block{}, err
	}

	return block, nil
}

// BlocksHashes returns a list of hashes of all the blocks in the chain.
func (c *Chain) BlocksHashes() ([][]byte, error) {
	bci := c.NewIterator()
	blocksCount, err := bci.BlocksCount()
	if err != nil {
		return nil, err
	}
	blocks := make([][]byte, 0, blocksCount)

	err = bci.ForEach(func(block Block) error {
		blocks = append(blocks, block.Hash)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return blocks, nil
}

// FindTransaction looks for a transaction by its id.
//
// It returns the block containing the transaction and the transaction itself.
func (c *Chain) FindTransaction(id []byte) (Block, tx.Tx, error) {
	bci := c.NewIterator()

	boltTx, err := c.Begin(false)
	if err != nil {
		return Block{}, tx.Tx{}, err
	}
	defer boltTx.Rollback()

	for {
		block, err := bci.Next(boltTx)
		if err != nil {
			return Block{}, tx.Tx{}, err
		}

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, id) == 0 {
				return block, tx, nil
			}
		}

		// There are no more blocks
		if block.IsGenesis() {
			break
		}
	}

	return Block{}, tx.Tx{}, errors.New("transaction not found")
}

// FindUTXOs finds all unspent transaction outputs and returns transactions with spent outputs removed.
func (c *Chain) FindUTXOs() (map[string][]tx.Output, error) {
	utxo := make(map[string][]tx.Output)
	spentTXOs := make(map[string][]int)
	bci := c.NewIterator()

	err := bci.ForEach(func(block Block) error {
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				// Was the output spent?
				if spentOutputs, ok := spentTXOs[txID]; ok {
					for _, spentOutIdx := range spentOutputs {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				// Add the output to the other outputs of this transaction
				outputs := utxo[txID]
				outputs = append(outputs, out)
				utxo[txID] = outputs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.PrevOutput.TxID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.PrevOutput.Index)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return utxo, nil
}

// LastBlock returns the last block in the chain.
func (c *Chain) LastBlock() (Block, error) {
	if c.tip == nil {
		return Block{}, errEmptyBlockchain
	}
	return c.Block(c.tip)
}

// SignTransaction signs inputs of a Transaction.
func (c *Chain) SignTransaction(t *tx.Tx, privKey *ecdsa.PrivateKey) error {
	prevTxs := make(map[string]tx.Tx, len(t.Inputs))

	for _, in := range t.Inputs {
		_, prevTx, err := c.FindTransaction(in.PrevOutput.TxID)
		if err != nil {
			return err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return t.Sign(privKey, prevTxs)
}

// VerifyTx returns an error if a transaction is not valid.
//
// TODO:
//   - utxo should not belong to the genesis block
//   - the size of the serialized transaction should not exceed max block size
//   - inputs referenced outputs should be unspent and in the chain/mempool
//   - inputs referenced outputs and new outputs should have the same amount of coins
func (c *Chain) VerifyTx(t tx.Tx) error {
	if t.IsCoinbase() {
		return nil
	}

	if len(t.Inputs) == 0 {
		return errors.New("transaction has no inputs")
	}
	if len(t.Outputs) == 0 {
		return errors.New("transaction has no outputs")
	}

	// Reject transactions already in the blockchain
	if _, _, err := c.FindTransaction(t.ID); err == nil {
		return fmt.Errorf("transaction %x already exists", t.ID)
	}

	prevTxs := make(map[string]tx.Tx, len(t.Inputs))
	for _, in := range t.Inputs {
		_, prevTx, err := c.FindTransaction(in.PrevOutput.TxID)
		if err != nil {
			return err
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	ok, err := t.Verify(prevTxs)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("invalid transaction signature")
	}

	return nil
}
