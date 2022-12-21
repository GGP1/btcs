package mempool

import (
	"encoding/hex"
	"sync"

	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/tx"
)

// TODO:
// - keep the mempool memory below <n> MB (use SizeBytes).
// In case the limit is about to be reached, remove transactions with the lowest fees (minmempoolfee).
// Overtime it should self-adjust to accept transactions with lower fees if there's enough space.
// Perhaps just take up to <n> transactions to make things simple, prioritizing the ones with higher fees.

// TxPool contains valid transactions that may be included in the next block.
type TxPool struct {
	mu   *sync.RWMutex
	pool map[string]tx.Tx
}

// NewTxPool returns a new transaction pool.
func NewTxPool() *TxPool {
	return &TxPool{
		pool: make(map[string]tx.Tx),
		mu:   &sync.RWMutex{},
	}
}

// Add adds a transaction to the pool.
func (t *TxPool) Add(tx tx.Tx) {
	txID := hex.EncodeToString(tx.ID)
	t.mu.Lock()
	t.pool[txID] = tx
	t.mu.Unlock()
}

// Contains returns whether the txID is in the pool or not.
func (t *TxPool) Contains(txID []byte) bool {
	t.mu.RLock()
	_, ok := t.pool[hex.EncodeToString(txID)]
	t.mu.RUnlock()
	return ok
}

// Count returns the number of transactions in the pool.
func (t *TxPool) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return len(t.pool)
}

// ForEach iterates over the pool executing f on each transaction.
func (t *TxPool) ForEach(f func(txID string, tx tx.Tx) error) error {
	t.mu.Lock()
	for id, tx := range t.pool {
		if err := f(id, tx); err != nil {
			return err
		}
	}
	t.mu.Unlock()

	return nil
}

// Get retrieves a transaction from the pool.
func (t *TxPool) Get(txID []byte) tx.Tx {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.pool[hex.EncodeToString(txID)]
}

// Remove deletes a transaction from the pool.
func (t *TxPool) Remove(txID []byte) {
	t.mu.Lock()
	delete(t.pool, hex.EncodeToString(txID))
	t.mu.Unlock()
}

// SizeBytes returns the size of the mempool in bytes.
func (t *TxPool) SizeBytes() (int, error) {
	t.mu.RLock()
	b, err := gob.Encode(t.pool)
	if err != nil {
		return 0, err
	}
	t.mu.RUnlock()

	return len(b), nil
}
