package wallet

import (
	"errors"

	"github.com/spf13/cobra"
)

var errInvalidAccountName = errors.New("account name not specified")

type runEFunc func(cmd *cobra.Command, args []string) error

// NewCmd returns a new wallet command.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Wallet operations",
	}

	cmd.AddCommand(
		newCreate(),
		newCreateAccount(),
		newCreateMnemonic(),
		newGetAccountBalance(),
		newGetAccount(),
		newGetNewAddress(),
		newListAccounts(),
		newListAddresses(),
		newValidateAddress(),
	)

	return cmd
}
