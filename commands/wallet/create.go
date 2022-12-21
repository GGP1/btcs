package wallet

import (
	"fmt"

	"github.com/GGP1/btcs/wallet"
	"github.com/tyler-smith/go-bip39"

	"github.com/spf13/cobra"
)

var mnemonic, passphrase string

func newCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a new wallet",
		RunE:  runCreate(),
	}

	f := cmd.Flags()
	f.StringVarP(&passphrase, "passphrase", "p", "", "Seed passphrase")
	f.StringVarP(&mnemonic, "mnemonic", "m", "", "Seed mnemonic, should be enclosed by doublequotes. If it's not specified a new one will be provided")

	return cmd
}

func runCreate() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		if mnemonic == "" {
			entropy, err := bip39.NewEntropy(256)
			if err != nil {
				return err
			}

			mnemonic, err = bip39.NewMnemonic(entropy)
			if err != nil {
				return err
			}

			fmt.Println("Mnemonic:", mnemonic)
		}

		wallet, err := wallet.NewWallet(mnemonic, passphrase)
		if err != nil {
			return err
		}

		if err := wallet.Save(); err != nil {
			return err
		}

		fmt.Println("Wallet created")
		return nil
	}
}
