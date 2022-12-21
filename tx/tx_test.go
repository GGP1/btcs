package tx_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/GGP1/btcs/wallet"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestSigning(t *testing.T) {
	wallet := newWallet(t)

	account, err := wallet.NewAccount("test")
	assert.NoError(t, err)

	pubKeyBytes := account.PublicKey()
	x := &big.Int{}
	y := &big.Int{}
	x.SetBytes(pubKeyBytes[:len(pubKeyBytes)/2])
	y.SetBytes(pubKeyBytes[len(pubKeyBytes)/2:])
	rawPubKey := ecdsa.PublicKey{Curve: btcec.S256(), X: x, Y: y}

	data := []byte("Bitcoin")
	signature, err := ecdsa.SignASN1(rand.Reader, account.PrivateKey(), data)
	assert.NoError(t, err)

	ok := ecdsa.VerifyASN1(&rawPubKey, data, signature)
	assert.True(t, ok)
}

// func TestTxSigning(t *testing.T) {
// 	wallet := newWallet(t)

// 	account, err := wallet.NewAccount("test")
// 	assert.NoError(t, err)

// 	addr, err := account.NewAddress(true)
// 	assert.NoError(t, err)

// 	inputs := []tx.Input{
// 		{
// 			TxID:   []byte("txID"),
// 			PubKey: account.PublicKey(),
// 			Vout:   0,
// 		},
// 	}
// 	outputs := []tx.Output{tx.NewOutput(1, addr)}

// 	txx, err := tx.New(inputs, outputs)
// 	assert.NoError(t, err)

// 	prevTx, err := tx.New([]tx.Input{{}}, nil)
// 	assert.NoError(t, err)

// 	prevTxs := map[string]tx.Tx{
// 		hex.EncodeToString(prevTx.ID): *prevTx,
// 	}

// 	err = txx.Sign(account.PrivateKey(), prevTxs)
// 	assert.NoError(t, err)

// 	ok, err := txx.Verify(prevTxs)
// 	assert.NoError(t, err)

// 	assert.True(t, ok)
// }

func newWallet(t *testing.T) *wallet.Wallet {
	entropy, err := bip39.NewEntropy(256)
	assert.NoError(t, err)

	mnemonic, err := bip39.NewMnemonic(entropy)
	assert.NoError(t, err)

	wallet, err := wallet.NewWallet(mnemonic, "")
	assert.NoError(t, err)

	return wallet
}
