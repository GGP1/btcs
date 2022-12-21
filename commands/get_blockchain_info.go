package commands

import (
	"fmt"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetBlockchainInfo() *cobra.Command {
	return &cobra.Command{
		Use:   "getblockchaininfo",
		Short: "Get information about the blockchain state",
		RunE:  runGetBlockchainInfo(),
	}
}

func runGetBlockchainInfo() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		blocks, err := client.ListBlocks()
		if err != nil {
			return err
		}

		for i := len(blocks) - 1; i >= 0; i-- {
			b := blocks[i]
			fmt.Printf(`============ Block %x ============
Height: %d
Previous block: %x
Nonce: %d
Timestamp: %v
`,
				b.Hash, b.Height, b.PrevBlockHash,
				b.Nonce, b.Timestamp)

			fmt.Print("\n")
			for _, tx := range b.Transactions {
				fmt.Println(tx)
			}
			fmt.Print("\n")
		}

		return nil
	}
}
