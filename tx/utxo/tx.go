package utxo

import (
	"encoding/hex"

	"github.com/GGP1/btcs/tx"
	"github.com/GGP1/btcs/wallet"
)

// NewTx creates a new transaction.
func NewTx(account *wallet.Account, to string, amount, fee int, set *Set) (*tx.Tx, error) {
	accumulated, utxos, err := set.AccountUTXOs(account, amount, fee)
	if err != nil {
		return nil, err
	}

	inputs := make([]tx.Input, 0)
	// Reuse object
	input := tx.Input{PubKey: account.PublicKey()}

	for _, spendableOutput := range utxos {
		txID, err := hex.DecodeString(spendableOutput.txID)
		if err != nil {
			return nil, err
		}

		input.PrevOutput.TxID = txID
		for _, outIdx := range spendableOutput.outsIndices {
			input.PrevOutput.Index = outIdx
			inputs = append(inputs, input)
		}
	}

	// The amount will now be locked with the receiver address,
	// this is how coins are transferred.
	outputs := []tx.Output{tx.NewOutput(amount, to)}
	if accumulated > amount+fee {
		// Create output for the change
		changeAddr, err := account.NewAddress(false)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, tx.NewOutput(accumulated-amount-fee, changeAddr))
	}

	tx, err := tx.New(inputs, outputs, fee)
	if err != nil {
		return nil, err
	}

	if err := set.Blockchain.SignTransaction(tx, account.PrivateKey()); err != nil {
		return nil, err
	}

	return tx, nil
}
