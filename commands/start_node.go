package commands

import (
	"errors"
	"os"
	"strings"

	"github.com/GGP1/btcs/logger"
	"github.com/GGP1/btcs/node"

	"github.com/spf13/cobra"
)

var (
	miner, debug bool
	nodes        []string
	address      string

	seedNodes = []string{
		"node1:3000",
		"node2:4000",
		"node3:5000",
	}
)

func newStartNode() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "startnode <account>",
		Short:   "Start bitcoin node server",
		Example: "startnode satoshi --address localhost:8838 --miner",
		RunE:    runStartNode(),
	}

	f := cmd.Flags()
	f.StringVarP(&address, "address", "a", "", "node server address")
	f.StringSliceVarP(&nodes, "nodes", "n", seedNodes, "nodes addresses to connect to")
	f.BoolVarP(&miner, "miner", "m", false, "whether the node will perform mining operations")
	f.BoolVar(&debug, "debug", false, "set the logger mode to debug")

	return cmd
}

func runStartNode() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		accountName := strings.Join(args, " ")
		if accountName == "" {
			return errors.New("account name not specified")
		}

		if address == "" {
			address = os.Getenv("NODE_ADDR")
			if address == "" {
				return errors.New("no address was provided")
			}
		}

		logger.SetDevelopment(debug)

		node, err := node.New(address, nodes, miner)
		if err != nil {
			return err
		}

		return node.Run(accountName)
	}
}
