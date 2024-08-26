package command

import (
	"strings"

	"github.com/spf13/cobra"
)

type WalletCommand struct {
	name     string
	mnemonic string

	rootCmd *cobra.Command
}

func NewWalletCommand(name string) *WalletCommand {
	c := &WalletCommand{
		name: strings.ToLower(name),
	}
	c.name = strings.ToLower(name)
	c.init()
	return c
}

func (c *WalletCommand) init() {
	c.rootCmd = &cobra.Command{
		Use:   cmdName,
		Short: longName,
		Args:  cobra.MinimumNArgs(1),
	}

	// mnemonics subcommand
	c.rootCmd.AddCommand(c.mnemonicCmd())
}

func (c *WalletCommand) Execute() error {
	return c.rootCmd.Execute()
}
