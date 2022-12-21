package wallet

import (
	"fmt"
	"strings"

	"github.com/GGP1/btcs/node/rpc"
	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

var unit string

func newGetAccountBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getaccountbalance <account>",
		Short: "Get account balance",
		Long: `Get account balance
		
Bitcoin denominations

SAT: Satoshi
uBTC: Microbit
mBTC: Millibit
BTC: Bitcoin`,
		RunE: runGetAccountBalance(),
	}

	cmd.Flags().StringVarP(&unit, "unit", "u", "BTC", "balance unit")

	return cmd
}

func runGetAccountBalance() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		name := strings.Join(args, " ")
		if name == "" {
			return errInvalidAccountName
		}

		w, err := wallet.Load()
		if err != nil {
			return err
		}
		if !w.AccountExists(name) {
			return fmt.Errorf("account %s does not exist", name)
		}
		account := w.Account(name)

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		utxosMap, err := client.GetAddressesUTXOs(account.UsedAddresses())
		if err != nil {
			return err
		}

		totalBalance := 0
		fmt.Printf("%q account unspent outputs\n\n", name)

		for address, outs := range utxosMap {
			addrBalance := 0
			for _, out := range outs {
				addrBalance += out.Value
			}

			totalBalance += addrBalance
			if addrBalance > 0 {
				fmt.Println(address+":", wallet.FormatBalance(addrBalance, unit))
			}
		}

		fmt.Printf("\nTotal balance: %s\n", wallet.FormatBalance(totalBalance, unit))
		return nil
	}
}
