package wallet

import (
	"errors"
	"fmt"

	"github.com/tyler-smith/go-bip39"

	"github.com/spf13/cobra"
)

var length int

func newCreateMnemonic() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createmnemonic",
		Short: "Creates and returns a mnemonic sentence of the length specified",
		RunE:  runCreateMnemonic(),
	}

	cmd.Flags().IntVarP(&length, "length", "l", 24, "Number of words in the mnemonic")

	return cmd
}

func runCreateMnemonic() runEFunc {
	return func(cmd *cobra.Command, args []string) error {
		bitSize, err := lengthToBitSize(length)
		if err != nil {
			return err
		}

		entropy, err := bip39.NewEntropy(bitSize)
		if err != nil {
			return err
		}

		mnemonic, err := bip39.NewMnemonic(entropy)
		if err != nil {
			return err
		}

		fmt.Println(mnemonic)
		return nil
	}
}

func lengthToBitSize(length int) (int, error) {
	switch length {
	case 12:
		return 128, nil
	case 15:
		return 160, nil
	case 18:
		return 192, nil
	case 21:
		return 224, nil
	case 24:
		return 256, nil

	default:
		return 0, errors.New("invalid mnemonic length")
	}
}
