package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/wallet"
	"github.com/spf13/cobra"
)

func newValidateAddress() *cobra.Command {
	return &cobra.Command{
		Use:   "validateaddress <address>",
		Short: "Validates an address",
		RunE:  runValidateAddress(),
	}
}

func runValidateAddress() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		address := strings.Join(args, " ")
		if err := wallet.ValidateAddress(address); err != nil {
			return err
		}

		fmt.Println("Valid address")
		return nil
	}
}
