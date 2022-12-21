package commands

import (
	"fmt"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetPeerInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "getpeerinfo",
		Short: "Get information about each connected node",
		RunE:  runGetPeerInfo(),
	}
}

func runGetPeerInfo() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		addresses, err := client.GetPeerInfo()
		if err != nil {
			return err
		}

		fmt.Println("Connected nodes\n---------------")
		for _, addr := range addresses {
			fmt.Println(addr)
		}
		return nil
	}
}
