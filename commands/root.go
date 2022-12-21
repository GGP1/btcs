package commands

import (
	"github.com/GGP1/btcs/commands/wallet"

	"github.com/spf13/cobra"
)

// RunEFunc is a cobra function returning an error.
type RunEFunc func(cmd *cobra.Command, args []string) error

// NewRoot returns the command that is the parent of all the other commands.
func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "btcs",
		SilenceErrors: true,
		SilenceUsage:  true,
		CompletionOptions: cobra.CompletionOptions{
			HiddenDefaultCmd: true,
		},
	}

	cmd.AddCommand(
		newAddNode(),
		newDisconnectNode(),
		newGetBalance(),
		newGetBlockCount(),
		newGetBlock(),
		newGetBlockchainInfo(),
		newGetDifficulty(),
		newGetPeerInfo(),
		newGetRawMempool(),
		newGetTransaction(),
		newPing(),
		newSendTx(),
		newStartNode(),
		newStopNode(),
		wallet.NewCmd(),
	)

	return cmd
}
