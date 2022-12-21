package main

import (
	"fmt"
	"os"

	"github.com/GGP1/btcs/commands"
)

func main() {
	root := commands.NewRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
	}
}
