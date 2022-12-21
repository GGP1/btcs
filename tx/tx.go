package tx

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/GGP1/btcs/encoding/gob"
	"github.com/GGP1/btcs/logger"

	"github.com/btcsuite/btcd/btcec/v2"
)

const (
	// baseSubsidy is the initial amount a miner receives for mining a block.
	//
	// Represented in satoshis (50 BTC).
	baseSubsidy = 5000000000

	// subsidyReductionPeriod represents the number of blocks until the subisidy is halved.
	//
	// In the Bitcoin mainnet, it is 210,000 blocks.
	subsidyReductionPeriod = 21
)

// Tx represents a transaction.
//
// Every new transaction must have at least one input and output, except coinbase.
type Tx struct {
	ID      []byte
	Inputs  []Input
	Outputs []Output
	// In Bitcoin, a transaction's fee is equal to the difference between the amount of coins
	// locked in the inputs' referenced outputs and the ones in the new outputs.
	//
	// To make things simpler we just include it as a field, the input surplus will still exist.
	Fee int
}

// New returns a new transaction with the inputs and outputs provided.
func New(inputs []Input, outputs []Output, fee int) (*Tx, error) {
	tx := &Tx{
		Inputs:  inputs,
		Outputs: outputs,
		Fee:     fee,
	}

	encTx, err := gob.Encode(tx)
	if err != nil {
		return nil, err
	}

	// Add a nonce to avoid duplicated coinbase transactions.
	//
	// In Bitcoin, this is done through the transaction scriptSig.
	// See BIP34: https://github.com/bitcoin/bips/blob/master/bip-0034.mediawiki
	nonce := make([]byte, 24)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	hash := sha256.Sum256(append(encTx, nonce...))
	doubleHash := sha256.Sum256(hash[:])
	tx.ID = doubleHash[:]

	return tx, nil
}

// NewCoinbase returns a new coinbase transaction.
func NewCoinbase(toAddr, data string, fees int, nextBlockHeight int32) (*Tx, error) {
	txin := Input{
		PubKey: []byte(data),
		PrevOutput: OutPoint{
			Index: -1,
		},
	}
	subsidy := calculateBlockSubsidy(nextBlockHeight)
	txOut := NewOutput(subsidy+fees, toAddr)
	logger.Debugf("Block %d subsidy: %d, fees: %d", nextBlockHeight, subsidy, fees)

	return New([]Input{txin}, []Output{txOut}, 0)
}

// IsCoinbase checks whether the transaction is coinbase
func (tx *Tx) IsCoinbase() bool {
	return len(tx.Inputs) == 1 &&
		len(tx.Inputs[0].PrevOutput.TxID) == 0 &&
		tx.Inputs[0].PrevOutput.Index == -1
}

// Sign signs the inputs of a transaction.
func (tx *Tx) Sign(privKey *ecdsa.PrivateKey, prevTxs map[string]Tx) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, vin := range tx.Inputs {
		if prevTxs[hex.EncodeToString(vin.PrevOutput.TxID)].ID == nil {
			return errors.New("previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	// Every transaction input is signed by the one who created the transaction
	for i, vin := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(vin.PrevOutput.TxID)]

		// Set the input public key to the public key hash of the referenced output
		// for calculating the data and then remove it
		txCopy.Inputs[i].PubKey = prevTx.Outputs[vin.PrevOutput.Index].PubKeyHash
		data, err := gob.Encode(txCopy)
		if err != nil {
			return err
		}
		txCopy.Inputs[i].PubKey = nil

		signature, err := ecdsa.SignASN1(rand.Reader, privKey, data)
		if err != nil {
			return err
		}

		tx.Inputs[i].Signature = signature
	}

	return nil
}

// String returns a human-readable representation of a transaction.
func (tx Tx) String() string {
	lines := make([]string, 0, 1+len(tx.Inputs)+len(tx.Outputs))

	lines = append(lines, fmt.Sprintf("--- Transaction %x ---\nFee: %d SAT", tx.ID, tx.Fee))

	inFormat := `  Input %d:
	TxID: 	 	%x
	Out:		%d
	Signature: 	%x
	PubKey:	 	%x`
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf(inFormat,
			i, input.PrevOutput.TxID, input.PrevOutput.Index,
			input.Signature, input.PubKey,
		))
	}

	outFormat := `  Output %d:
	Value: 	 %d
	PubKey Script:	 %x`
	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf(outFormat, i, output.Value, output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing.
//
// https://en.bitcoin.it/w/images/en/7/70/Bitcoin_OpCheckSig_InDetail.png
func (tx *Tx) TrimmedCopy() Tx {
	inputs := make([]Input, 0, len(tx.Inputs))
	for _, in := range tx.Inputs {
		// Signature and public key omitted since we don't need to sign them
		inputs = append(inputs, Input{
			PrevOutput: OutPoint{
				TxID:  in.PrevOutput.TxID,
				Index: in.PrevOutput.Index,
			},
		})
	}

	return Tx{
		ID:      tx.ID,
		Inputs:  inputs,
		Outputs: tx.Outputs,
	}
}

// Verify validates signatures of transaction inputs.
func (tx *Tx) Verify(prevTxs map[string]Tx) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	txCopy := tx.TrimmedCopy()
	// The curve must be KoblitzCurve and not elliptic.P256()
	// in order for the verification to succeed
	curve := btcec.S256()

	for i, vin := range tx.Inputs {
		prevTx := prevTxs[hex.EncodeToString(vin.PrevOutput.TxID)]

		// Same as in Sign()
		txCopy.Inputs[i].PubKey = prevTx.Outputs[vin.PrevOutput.Index].PubKeyHash
		data, err := gob.Encode(txCopy)
		if err != nil {
			return false, err
		}
		txCopy.Inputs[i].PubKey = nil

		// Get the public key from the input PubKey field
		x := big.Int{}
		y := big.Int{}
		x.SetBytes(vin.PubKey[:len(vin.PubKey)/2])
		y.SetBytes(vin.PubKey[len(vin.PubKey)/2:])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}

		if !ecdsa.VerifyASN1(&rawPubKey, data, vin.Signature) {
			return false, nil
		}
	}

	return true, nil
}

// calculateBlockSubsidy returns the subsidy for the miner depending on the height of the
// block being mined.
//
// The subsidy halves every subsidyReductionPeriod blocks.
func calculateBlockSubsidy(nextBlockHeight int32) int {
	halvings := uint(nextBlockHeight / subsidyReductionPeriod)
	// Force block reward to zero when right shift is undefined.
	if halvings >= 64 {
		return 0
	}

	if nextBlockHeight%subsidyReductionPeriod == 0 && nextBlockHeight != 0 {
		logger.Debug("New block subsidy halving")
	}

	// Halve baseSubsidy every halvingPeriod blocks
	return baseSubsidy >> halvings
}
