package command

import (
	"strings"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/mengbin92/wallet/kms"
	"github.com/spf13/cobra"
)

// WalletCommand 结构体表示钱包命令，包含命令名称、助记词以及根命令
type WalletCommand struct {
	name      string
	mnemonic  string
	masterKey *hdkeychain.ExtendedKey

	rootCmd *cobra.Command
}

// NewWalletCommand 函数创建并初始化一个新的 WalletCommand 实例
func NewWalletCommand(name string) *WalletCommand {
	c := &WalletCommand{
		name: strings.ToLower(name),
	}
	c.init()
	return c
}

// init 方法初始化 WalletCommand 的根命令及其子命令
func (c *WalletCommand) init() {
	c.rootCmd = &cobra.Command{
		Use:   cmdName,
		Short: longName,
		Args:  cobra.MinimumNArgs(1),
	}

	// mnemonics subcommand
	c.rootCmd.AddCommand(c.mnemonicCmd())
	// key subcommand
	c.rootCmd.AddCommand(c.keyCmd())
}

// Execute 方法执行 WalletCommand 的根命令
func (c *WalletCommand) Execute() error {
	return c.rootCmd.Execute()
}

// GenMasterKey 方法生成主密钥
func (c *WalletCommand) genMasterKey(password, network string) (*hdkeychain.ExtendedKey, error) {
	return kms.GenMasterKey(c.mnemonic, password, network)
}
