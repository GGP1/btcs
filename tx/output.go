package tx

import (
	"bytes"

	"github.com/GGP1/btcs/encoding/base58"
)

// Output represents a transaction output,
// they are indivisible and the place where coins are actually stored.
type Output struct {
	// Hashed public key of the coins' owner
	PubKeyHash []byte
	// Represented in satoshis
	Value int
}

// NewOutput create a new transaction output.
func NewOutput(value int, address string) Output {
	pubKeyHash := base58.Decode([]byte(address))
	return Output{
		Value:      value,
		PubKeyHash: pubKeyHash[1 : len(pubKeyHash)-4],
	}
}

// IsLockedWithKey checks if the output can be used by the owner of the public key.
func (o Output) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(o.PubKeyHash, pubKeyHash) == 0
}
