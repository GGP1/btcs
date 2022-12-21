package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

func newCreateAccount() *cobra.Command {
	return &cobra.Command{
		Use:   "createaccount <name>",
		Short: "Creates a new account",
		RunE:  runCreateAccount(),
	}
}

func runCreateAccount() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidAccountName
		}

		wallet, err := wallet.Load()
		if err != nil {
			return err
		}

		if wallet.AccountExists(name) {
			return fmt.Errorf("account %s already exists", name)
		}

		if _, err := wallet.NewAccount(name); err != nil {
			return err
		}

		fmt.Printf("%q account created\n", name)
		return wallet.Save()
	}
}
