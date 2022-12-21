package commands

import (
	"fmt"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetBlockCount() *cobra.Command {
	return &cobra.Command{
		Use:   "getblockcount",
		Short: "Get blockchain block count",
		RunE:  runGetBlockCount(),
	}
}

func runGetBlockCount() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		bestHeight, err := client.GetBestHeight()
		if err != nil {
			return err
		}

		fmt.Println("Block count:", bestHeight+1)
		return nil
	}
}
