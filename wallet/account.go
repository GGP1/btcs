package wallet

import (
	"crypto/ecdsa"

	"github.com/btcsuite/btcd/btcec/v2"
)

// The gap limit is the maximum number of consecutive unused addresses
// in your deterministic sequence of addresses.
const gapLimit = 20

// Account contains a master key and the name the user has chosen to identify it.
type Account struct {
	// Use one mutex for both maps as they are accessed
	PrivKey            *Key
	PubKey             *Key
	ChangeAddresses    map[string]struct{}
	ReceivingAddresses map[string]struct{}
	NextKeyIndex       uint32
}

// NewAccount returns a new account with the key provided.
func NewAccount(masterKey *Key) *Account {
	return &Account{
		PrivKey:            masterKey,
		PubKey:             masterKey.Public(),
		ChangeAddresses:    map[string]struct{}{},
		ReceivingAddresses: map[string]struct{}{},
	}
}

// NewAddress returns an address corresponding to the account that hasn't been used before.
//
// If receiving is false, the address will be marked as a change address.
func (a *Account) NewAddress(receiving bool) (string, error) {
	key, err := a.PubKey.Child(a.NextKeyIndex)
	if err != nil {
		return "", err
	}
	a.NextKeyIndex++

	address := key.Address()
	if receiving {
		a.addReceivingAddresses(address)
	} else {
		a.addChangeAddresses(address)
	}

	return address, nil
}

// NewAddresses returns a list of receiving addresses. Change addresses are handled internally.
func (a *Account) NewAddresses() ([]string, error) {
	addresses := make([]string, 0, gapLimit)

	lastKeyIdx := a.NextKeyIndex
	for i := lastKeyIdx; i < lastKeyIdx+gapLimit; i++ {
		address, err := a.NewAddress(true)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, address)
	}

	return addresses, nil
}

// PublicKey returns the public key of an account.
func (a *Account) PublicKey() []byte {
	priv := a.PrivateKey()
	return append(priv.PublicKey.X.Bytes(), priv.PublicKey.Y.Bytes()...)
}

// PrivateKey returns the private key of an account.
func (a *Account) PrivateKey() *ecdsa.PrivateKey {
	priv, _ := btcec.PrivKeyFromBytes(a.PrivKey.Key)
	return priv.ToECDSA()
}

// UsedAddresses returns receiving and change addresses that were utilizied.
func (a *Account) UsedAddresses() []string {
	addresses := make([]string, 0, len(a.ChangeAddresses)+len(a.ReceivingAddresses))

	for addr := range a.ChangeAddresses {
		addresses = append(addresses, addr)
	}

	for addr := range a.ReceivingAddresses {
		addresses = append(addresses, addr)
	}

	return addresses
}

// addChangeAddresses marks multiple change addresses as used.
func (a *Account) addChangeAddresses(addresses ...string) {
	for _, addr := range addresses {
		a.ChangeAddresses[addr] = struct{}{}
	}
}

// addReceivingAddresses marks multiple receiving addresses as used.
func (a *Account) addReceivingAddresses(addresses ...string) {
	for _, addr := range addresses {
		a.ReceivingAddresses[addr] = struct{}{}
	}
}
