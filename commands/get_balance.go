package commands

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/node/rpc"
	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

var unit string

func newGetBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getbalance <address>",
		Short: "Get balance of an address",
		Long: `Get balance of an address.
To look for the balance of an account, use 'wallet getaccountbalance <account>'
		
Bitcoin denominations

SAT: Satoshi
uBTC: Microbit
mBTC: Millibit
BTC: Bitcoin`,
		RunE: runGetBalance(),
	}

	cmd.Flags().StringVarP(&unit, "unit", "u", "BTC", "balance unit")

	return cmd
}

func runGetBalance() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		address := strings.Join(args, " ")
		if err := wallet.ValidateAddress(address); err != nil {
			return err
		}

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		utxos, err := client.GetAddressUTXOs(address)
		if err != nil {
			return err
		}

		balance := 0
		for _, out := range utxos {
			balance += out.Value
		}

		fmt.Printf("%q balance: %s\n", address, wallet.FormatBalance(balance, unit))
		return nil
	}
}
