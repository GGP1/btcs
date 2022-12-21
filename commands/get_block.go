package commands

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GGP1/btcs/node/rpc"

	"github.com/spf13/cobra"
)

func newGetBlock() *cobra.Command {
	return &cobra.Command{
		Use:   "getblock <hash>",
		Short: "Get information about a specific block",
		RunE:  runGetBlock(),
	}
}

func runGetBlock() RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		blockHash := strings.Join(args, " ")
		if blockHash == "" {
			return errors.New("block hash not specified. Use 'getblock <hash>'")
		}

		client, err := rpc.NewClient()
		if err != nil {
			return err
		}
		defer client.Close()

		hash, err := hex.DecodeString(blockHash)
		if err != nil {
			return err
		}

		block, err := client.GetBlock(hash)
		if err != nil {
			return err
		}

		bestHeight, err := client.GetBestHeight()
		if err != nil {
			return err
		}

		fmt.Printf(`--- %x ---
Previous block: %x
Height: %d
Merkle root hash: %x
Timestamp: %s
Nonce: %d
Number of confirmations: %d`,
			block.Hash,
			block.PrevBlockHash,
			block.Height,
			block.MerkleRootHash,
			time.Unix(block.Timestamp, 0).Format(time.RFC3339Nano),
			block.Nonce,
			bestHeight-block.Height,
		)

		fmt.Print("\n")
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		return nil
	}
}
