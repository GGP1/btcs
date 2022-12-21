package commands

import (
	"fmt"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newStopNode() *cobra.Command {
	return &cobra.Command{
		Use:   "stopnode",
		Short: "Stop bitcoin node server",
		RunE:  runStopNode(),
	}
}

func runStopNode() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		if err := client.Stop(); err != nil {
			return err
		}

		fmt.Println("Node stopped")
		return nil
	}
}
