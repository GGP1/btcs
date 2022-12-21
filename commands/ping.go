package commands

import (
	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newPing() *cobra.Command {
	return &cobra.Command{
		Use:   "ping",
		Short: "Ping all nodes",
		RunE:  runPing(),
	}
}

func runPing() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		return client.SendPing()
	}
}
