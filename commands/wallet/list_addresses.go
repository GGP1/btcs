package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

func newListAddresses() *cobra.Command {
	return &cobra.Command{
		Use:   "listaddresses <account>",
		Short: "List an account's used addresses and their balance",
		RunE:  runListAddresses(),
	}
}

func runListAddresses() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidAccountName
		}

		wallet, err := wallet.Load()
		if err != nil {
			return err
		}
		defer wallet.Save()

		account := wallet.Account(name)
		addresses, err := account.NewAddresses()
		if err != nil {
			return err
		}

		for _, address := range addresses {
			fmt.Println(address)
		}
		return nil
	}
}
