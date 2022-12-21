package wallet

import (
	"errors"
	"os"

	"github.com/GGP1/btcs/encoding/gob"

	"github.com/tyler-smith/go-bip39"
)

const walletPath = "wallet.dat"

// Wallet represents a hierarchical deterministic wallet.
type Wallet struct {
	MasterKey *Key
	// map[name]Account
	Accounts       map[string]*Account
	NextChildIndex uint32
}

// NewWallet creates a new wallet.
func NewWallet(mnemonic, passphrase string) (*Wallet, error) {
	if _, err := os.Stat(walletPath); os.IsExist(err) {
		return nil, errors.New("wallet already exists")
	}

	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, errors.New("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, passphrase)
	masterKey, err := NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	wallet := &Wallet{
		MasterKey: masterKey,
		Accounts:  make(map[string]*Account),
	}

	return wallet, nil
}

// Load loads the wallet from persistent storage.
//
// Call Save to write the Wallet's changes to persistent storage when done.
func Load() (*Wallet, error) {
	fileContent, err := os.ReadFile(walletPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("no wallet found, use <btcs wallet create> to create it")
		}
		return nil, err
	}

	return gob.Decode[*Wallet](fileContent)
}

// Account returns an account with the given name.
func (w *Wallet) Account(name string) *Account {
	return w.Accounts[name]
}

// AccountExists returns whether the account exists or not.
func (w *Wallet) AccountExists(name string) bool {
	_, ok := w.Accounts[name]
	return ok
}

// AccountNames returns the wallet accounts.
func (w *Wallet) AccountNames() []string {
	accounts := make([]string, 0, len(w.Accounts))
	for name := range w.Accounts {
		accounts = append(accounts, name)
	}

	return accounts
}

// NewAccount creates a new account.
func (w *Wallet) NewAccount(name string) (*Account, error) {
	childKey, err := w.MasterKey.Child(w.NextChildIndex)
	if err != nil {
		return nil, err
	}
	w.NextChildIndex++

	account := NewAccount(childKey)
	w.Accounts[name] = account

	return account, nil
}

// Save stores the wallet structure into persistent storage.
// The file is left unencrypted on purpose so it's easier to read its content.
//
// Save should be always called after a change in it or in an account.
func (w *Wallet) Save() error {
	b, err := gob.Encode(w)
	if err != nil {
		return err
	}

	return os.WriteFile(walletPath, b, 0o644)
}
