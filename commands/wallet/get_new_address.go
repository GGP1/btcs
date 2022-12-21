package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

func newGetNewAddress() *cobra.Command {
	return &cobra.Command{
		Use:   "getnewaddress <account>",
		Short: "Generates a new address for receiving payments",
		RunE:  runGetNewAddress(),
	}
}

func runGetNewAddress() runEFunc {
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

		if !wallet.AccountExists(name) {
			return fmt.Errorf("account %s does not exist", name)
		}

		account := wallet.Account(name)
		address, err := account.NewAddress(true)
		if err != nil {
			return err
		}

		fmt.Println(address)
		return nil
	}
}
