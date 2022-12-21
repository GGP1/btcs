package commands

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newAddNode() *cobra.Command {
	return &cobra.Command{
		Use:   "addnode <address>",
		Short: "Add a node to the peer list",
		RunE:  runAddNode(),
	}
}

func runAddNode() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		address := strings.Join(args, " ")
		if _, err := net.ResolveTCPAddr("tcp", address); err != nil {
			return errors.New("invalid address")
		}

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		peersNum, err := client.AddNode(address)
		if err != nil {
			return err
		}

		fmt.Println("Number of peers in the list:", peersNum)
		return nil
	}
}
