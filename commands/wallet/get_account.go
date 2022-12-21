package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

func newGetAccount() *cobra.Command {
	return &cobra.Command{
		Use:   "getaccount <name>",
		Short: "Get information about a specific account",
		RunE:  runGetAccount(),
	}
}

func runGetAccount() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidAccountName
		}

		wallet, err := wallet.Load()
		if err != nil {
			return err
		}

		if !wallet.AccountExists(name) {
			return fmt.Errorf("account %q does not exist", name)
		}
		account := wallet.Account(name)

		fmt.Printf(`Name: %s
Next key index: %d
xPub: %x

Receiving addresses
-------------------
`, name, account.NextKeyIndex, account.PubKey.Serialize())

		for addr := range account.ReceivingAddresses {
			fmt.Println(addr)
		}

		fmt.Println("\nChange addresses\n-------------------")

		for addr := range account.ChangeAddresses {
			fmt.Println(addr)
		}
		return nil
	}
}
