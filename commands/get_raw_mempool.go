package commands

import (
	"fmt"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetRawMempool() *cobra.Command {
	return &cobra.Command{
		Use:   "getrawmempool",
		Short: "Get all transaction ids in the memory pool",
		RunE:  runGetRawMempool(),
	}
}

func runGetRawMempool() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		txIDs, err := client.GetRawMempool()
		if err != nil {
			return err
		}

		if len(txIDs) == 0 {
			fmt.Println("There are no transactions in the mempool")
			return nil
		}

		for _, txID := range txIDs {
			fmt.Println(txID)
		}
		return nil
	}
}
