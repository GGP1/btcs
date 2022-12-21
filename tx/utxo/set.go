package utxo

import (
	"encoding/hex"
	"errors"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/encoding/base58"
	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/wallet"

	bolt "go.etcd.io/bbolt"
)

const utxoBucket = "chainstate"

// Set represents a UTXO set and holds all the unspent transaction outputs of an address.
type Set struct {
	Blockchain *block.Chain
}

// UTXO represents an output that has never been part of an input.
type UTXO struct {
	txID        string
	outsIndices []int
}

// AccountUTXOs returns an account's unspent outputs to be used in a new transaction.
//
// It returns an error if the account doesn't have enough funds.
func (s *Set) AccountUTXOs(account *wallet.Account, amount, fee int) (int, []UTXO, error) {
	boltTx, err := s.Blockchain.Begin(false)
	if err != nil {
		return 0, nil, err
	}
	defer boltTx.Rollback()

	b := boltTx.Bucket([]byte(utxoBucket))
	c := b.Cursor()
	utxos := make([]UTXO, 0, b.Stats().KeyN)
	pubKeyHashes := addressesToPubKeyHashes(account.UsedAddresses())
	targetAmount := amount + fee
	accumulated := 0

	for k, v := c.First(); k != nil; k, v = c.Next() {
		txID := hex.EncodeToString(k)
		outputs, err := gob.Decode[[]tx.Output](v)
		if err != nil {
			return 0, nil, err
		}

		indices := make([]int, 0)
		for i, out := range outputs {
			if accumulated >= targetAmount {
				// We have already collected enough outputs for the transaction
				if len(indices) > 0 {
					utxos = append(utxos, UTXO{
						txID:        txID,
						outsIndices: indices,
					})
				}

				return accumulated, utxos, nil
			}

			lockedByAccount := false
			for _, pubKeyHash := range pubKeyHashes {
				// Look for outputs that belong to the account addresses we have
				if out.IsLockedWithKey(pubKeyHash) {
					lockedByAccount = true
					break
				}
			}

			if !lockedByAccount {
				continue
			}

			accumulated += out.Value
			indices = append(indices, i)
		}

		utxos = append(utxos, UTXO{
			txID:        txID,
			outsIndices: indices,
		})
	}

	if accumulated < targetAmount {
		return 0, nil, errors.New("account has not enough funds")
	}

	return accumulated, utxos, nil
}

// FindUTXOs finds UTXO for a public key hash.
func (s *Set) FindUTXOs(pubKeyHash []byte) ([]tx.Output, error) {
	var utxos []tx.Output

	err := s.Blockchain.View(func(boltTx *bolt.Tx) error {
		b := boltTx.Bucket([]byte(utxoBucket))

		return b.ForEach(func(_, v []byte) error {
			outputs, err := gob.Decode[[]tx.Output](v)
			if err != nil {
				return err
			}

			for _, output := range outputs {
				if output.IsLockedWithKey(pubKeyHash) {
					utxos = append(utxos, output)
				}
			}

			return nil
		})
	})
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

// Reindex rebuilds the UTXO set.
func (s *Set) Reindex() error {
	utxo, err := s.Blockchain.FindUTXOs()
	if err != nil {
		return err
	}

	return s.Blockchain.Update(func(boltTx *bolt.Tx) error {
		bucketName := []byte(utxoBucket)
		if err := boltTx.DeleteBucket(bucketName); err != nil && err != bolt.ErrBucketNotFound {
			return err
		}
		b, err := boltTx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		for txID, outputs := range utxo {
			id, err := hex.DecodeString(txID)
			if err != nil {
				return err
			}

			encOutputs, err := gob.Encode(outputs)
			if err != nil {
				return err
			}

			if err := b.Put(id, encOutputs); err != nil {
				return err
			}
		}

		return nil
	})
}

// Update updates the UTXO set with transactions from the block received.
//
// The block is considered to be the tip of a blockchain.
func (s *Set) Update(block block.Block) error {
	return s.Blockchain.Update(func(boltTx *bolt.Tx) error {
		b := boltTx.Bucket([]byte(utxoBucket))

		for _, transaction := range block.Transactions {
			if !transaction.IsCoinbase() {
				for _, in := range transaction.Inputs {
					outsBytes := b.Get(in.PrevOutput.TxID)
					outs, err := gob.Decode[[]tx.Output](outsBytes)
					if err != nil {
						return err
					}

					// Find new unspent outputs
					unspentOuts := make([]tx.Output, 0)
					for outIdx, out := range outs {
						if outIdx != in.PrevOutput.Index {
							unspentOuts = append(unspentOuts, out)
						}
					}

					// Delete input with no unspent outputs
					if len(unspentOuts) == 0 {
						if err := b.Delete(in.PrevOutput.TxID); err != nil {
							return err
						}
						continue
					}

					encOuts, err := gob.Encode(unspentOuts)
					if err != nil {
						return err
					}

					if err := b.Put(in.PrevOutput.TxID, encOuts); err != nil {
						return err
					}
				}
			}

			encOuts, err := gob.Encode(transaction.Outputs)
			if err != nil {
				return err
			}

			if err := b.Put(transaction.ID, encOuts); err != nil {
				return err
			}
		}

		return nil
	})
}

func addressesToPubKeyHashes(addresses []string) [][]byte {
	pubKeyHashes := make([][]byte, 0, len(addresses))
	for _, addr := range addresses {
		pubKeyHash := base58.Decode([]byte(addr))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
		pubKeyHashes = append(pubKeyHashes, pubKeyHash)
	}
	return pubKeyHashes
}
