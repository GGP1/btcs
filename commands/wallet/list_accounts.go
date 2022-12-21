package wallet

import (
	"fmt"

	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

func newListAccounts() *cobra.Command {
	return &cobra.Command{
		Use:   "listaccounts",
		Short: "List the account names inside the wallet",
		RunE:  runListAccounts(),
	}
}

func runListAccounts() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		wallet, err := wallet.Load()
		if err != nil {
			return err
		}

		for _, name := range wallet.AccountNames() {
			fmt.Println(name)
		}
		return nil
	}
}
