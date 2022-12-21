package commands

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GGP1/btcs/node"
	"github.com/GGP1/btcs/node/rpc"
	"github.com/GGP1/btcs/wallet"

	"github.com/spf13/cobra"
)

var (
	to, txUnit  string
	amount, fee int
)

func newSendTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sendtx <account>",
		Short:   "Create, sign and broadcast a transaction",
		Example: "sendtx satoshi --to 1MVBUT4h8q7c5xAuKiEwqCY6xexN6cWUTV --amount 1 --unit BTC --fee 20000",
		RunE:    runSendTx(),
	}

	f := cmd.Flags()
	f.StringVarP(&to, "to", "t", "", "to address")
	f.IntVarP(&amount, "amount", "a", 0, "transaction amount")
	f.IntVarP(&fee, "fee", "f", 0, "transaction fee (denominated in SAT)")
	f.StringVarP(&txUnit, "unit", "u", "BTC", "transaction amount unit")
	cmd.MarkFlagRequired("to")
	cmd.MarkFlagRequired("amount")

	return cmd
}

func runSendTx() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		accountName := strings.Join(args, " ")
		if accountName == "" {
			return errors.New("account name not specified")
		}

		if err := wallet.ValidateAddress(to); err != nil {
			return fmt.Errorf("recipient %w", err)
		}

		if amount <= 0 {
			return errors.New("invalid amount, must be higher than zero")
		}

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		params := node.SendTxParams{
			AccountName: accountName,
			To:          to,
			Amount:      amountToSats(txUnit, amount),
			Fee:         fee,
		}
		txID, err := client.SendTx(params)
		if err != nil {
			return err
		}

		fmt.Printf("Transaction sent. ID: %x\n", txID)
		return nil
	}
}

func amountToSats(unit string, amount int) int {
	switch strings.ToLower(unit) {
	case "ubtc":
		return amount * 100
	case "mbtc":
		return amount * 100000
	case "btc":
		return amount * 100000000
	default:
		return amount
	}
}
