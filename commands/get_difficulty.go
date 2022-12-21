package commands

import (
	"fmt"

	"github.com/GGP1/btcs/block"
	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetDifficulty() *cobra.Command {
	return &cobra.Command{
		Use:   "getdifficulty",
		Short: "Get the proof-of-work difficulty",
		RunE:  runGetDifficulty(),
	}
}

func runGetDifficulty() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		b, err := client.GetLastBlock()
		if err != nil {
			return err
		}

		fmt.Println("Difficulty:", block.BigToCompact(block.MaxTarget)/block.CalculateNextDifficulty(b))
		return nil
	}
}
