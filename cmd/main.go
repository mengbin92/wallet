package main

import (
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{Use: "wallet-tool"}
	rootCmd.AddCommand(NewWalletCmd())
	rootCmd.AddCommand(NewCollectorCmd())
	rootCmd.Execute()
}
