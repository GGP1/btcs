package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetTransaction() *cobra.Command {
	return &cobra.Command{
		Use:   "gettransaction <id>",
		Short: "Get information about a specific transaction",
		RunE:  runGetTransaction(),
	}
}

func runGetTransaction() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		txID := strings.Join(args, " ")
		if txID == "" {
			return errors.New("transaction id not specified. Use 'gettransaction <id>'")
		}

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		id, err := hex.DecodeString(txID)
		if err != nil {
			return err
		}

		block, tx, err := client.GetTransaction(id)
		if err != nil {
			return err
		}

		bestHeight, err := client.GetBestHeight()
		if err != nil {
			return err
		}

		fmt.Printf(`Block: %x
Confirmations: %d

%v
`, block.Hash, bestHeight-block.Height, tx)
		return nil
	}
}
