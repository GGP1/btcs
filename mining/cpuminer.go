package mining

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/logger"
	"github.com/GGP1/btcs/mempool"
	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/wallet"
)

const maxNonce = math.MaxUint32

// CPUMiner mines blocks using the CPU.
type CPUMiner struct {
	txPool    *mempool.TxPool
	newBlocks <-chan block.Block
	// coinbaseAddr is the address where mining rewards will be send
	coinbaseAddr string
}

// NewCPUMiner returns an object that mines blocks with the CPU.
func NewCPUMiner(accountName string, txPool *mempool.TxPool, newBlocks <-chan block.Block) (Miner, error) {
	wallet, err := wallet.Load()
	if err != nil {
		return nil, err
	}

	if !wallet.AccountExists(accountName) {
		return nil, fmt.Errorf("account %q does not exist", accountName)
	}

	coinbaseAddr, err := wallet.Account(accountName).NewAddress(true)
	if err != nil {
		return nil, err
	}

	if err := wallet.Save(); err != nil {
		return nil, err
	}

	logger.Info("Mining rewards and fees will be send to: ", coinbaseAddr)

	return &CPUMiner{
		coinbaseAddr: coinbaseAddr,
		txPool:       txPool,
		newBlocks:    newBlocks,
	}, nil
}

// Mine solves a block's puzzle and returns the mined block if it suceeds.
//
// If lastBlock is nil, it will mine the genesis block.
//
// It must be called inside a goroutine.
func (c *CPUMiner) Mine(prevBlock *block.Block) (block.Block, error) {
	b, err := c.buildBlock(prevBlock)
	if err != nil {
		return block.Block{}, err
	}

	if err := c.mine(b); err != nil {
		return block.Block{}, err
	}

	// Remove the block's transactions from the mempool
	// Ignore the first transaction (coinbase)
	for _, tx := range b.Transactions[1:] {
		c.txPool.Remove(tx.ID)
	}

	return *b, nil
}

// buildBlock creates the block and populates it with transactions from the pool.
func (c *CPUMiner) buildBlock(prevBlock *block.Block) (*block.Block, error) {
	// TODO:
	// - Take transactions from the pool until block is full (reaches size limit).
	// - Prioritize transactions with higher SAT/bytes fees.
	transactions := make([]tx.Tx, 0, c.txPool.Count())
	fees := 0
	_ = c.txPool.ForEach(func(txID string, tx tx.Tx) error {
		transactions = append(transactions, tx)
		fees += tx.Fee
		return nil
	})

	// Create the transaction that sends us the subsidy and fees if we succeed
	coinbaseTx, err := tx.NewCoinbase(c.coinbaseAddr, "", fees, prevBlock.Height+1)
	if err != nil {
		return nil, err
	}

	transactions = append([]tx.Tx{*coinbaseTx}, transactions...)
	return block.NewBlock(prevBlock, transactions)
}

// mine hashes the block data and different nonces until it finds a hash lower than the target.
func (c *CPUMiner) mine(b *block.Block) error {
	data, err := b.PowData()
	if err != nil {
		return err
	}

	target := block.CompactToBig(b.Bits)

	var (
		hashInt big.Int
		hash    [sha256.Size]byte
	)

Loop:
	for b.Nonce < maxNonce {
		select {
		case <-c.newBlocks:
			// Stop mining if another node has already completed the task
			logger.Info("New block received")
			return nil

		default:
			n, err := block.IntToBytes(int64(b.Nonce))
			if err != nil {
				return err
			}

			fullData := append(data, n...)
			hash = sha256.Sum256(fullData)

			// The hash has to be lower than the target to be accepted
			hashInt.SetBytes(hash[:])
			if hashInt.Cmp(target) <= 0 {
				b.Hash = hash[:]
				break Loop
			}

			b.Nonce++
		}
	}

	logger.Infof("New block mined: %x", b.Hash)
	return nil
}
