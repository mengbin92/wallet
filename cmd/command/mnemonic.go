package command

import (
	"fmt"

	"github.com/mengbin92/wallet/kms"
	"github.com/mengbin92/wallet/storage"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// mnemonicCmd 方法定义了一个用于管理助记词的命令，该命令包含创建和加载助记词的子命令。
func (c *WalletCommand) mnemonicCmd() *cobra.Command {
	genMnemonicCmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Manage mnemonic",
		Long:  "Manage mnemonic",
	}

	genMnemonicCmd.AddCommand(c.createMnemonic())
	// genMnemonicCmd.AddCommand(c.saveMnemonic())
	genMnemonicCmd.AddCommand(c.loadMnemonic())

	return genMnemonicCmd
}

// createMnemonic 方法定义了一个创建新助记词的命令，该命令会生成一个新的助记词并保存到文件中。
func (c *WalletCommand) createMnemonic() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new mnemonic and save it to file",
		Long:  `Create a new mnemonic and save it to file, example: ./wallet mnemonic create ./mnemonic.txt password
		The password is optional, if not provided, the program will generate a random password.`,
		RunE:  c.runCreateMnemonic,
	}
}

// saveMnemonic 方法定义了一个保存助记词到文件的命令。
func (c *WalletCommand) saveMnemonic() *cobra.Command {
	return &cobra.Command{
		Use:   "save",
		Short: "Save a mnemonic to file",
		Long:  "Save a mnemonic to file",
		RunE:  c.runSaveMnemonic,
	}
}

// loadMnemonic 方法定义了一个从文件加载助记词的命令。
func (c *WalletCommand) loadMnemonic() *cobra.Command {
	return &cobra.Command{
		Use:   "load",
		Short: "Load a mnemonic from file",
		Long:  "Load a mnemonic from file",
		RunE:  c.runLoadMnemonic,
	}
}

// runCreateMnemonic 方法实现了创建新助记词的逻辑，包括生成助记词和保存助记词到文件。
func (c *WalletCommand) runCreateMnemonic(cmd *cobra.Command, args []string) error {
	fmt.Println("Create a new mnemonic")
	if len(args) < 1 {
		return errors.New("Please provide the file path to save the mnemonic, e.g. ./mnemonic.txt password")
	}

	mnemonic, err := kms.GenMnemonic()
	if err != nil {
		return errors.Wrap(err, "generate mnemonic failed")
	}

	fmt.Println("Your new mnemonic is: ", mnemonic)
	c.mnemonic = mnemonic

	return c.runSaveMnemonic(cmd, args)
}

// runSaveMnemonic 方法实现了保存助记词到文件的逻辑，包括加密助记词和保存加密后的内容到文件。
func (c *WalletCommand) runSaveMnemonic(cmd *cobra.Command, args []string) error {
	store := storage.NewLocalStorage(args[0])
	var password string
	var err error

	if len(args) == 2 {
		password = args[1]
	} else {
		password, err = utils.CreatePassphrase(12)
		if err != nil {
			return errors.Wrap(err, "create password failed")
		}
	}
	// encrypt the mnemonic with password
	encryptedMnemonic, err := utils.AesEncrypt([]byte(c.mnemonic), password)
	if err != nil {
		return errors.Wrap(err, "encrypt mnemonic failed")
	}
	// save the mnemonic to file
	err = store.Save(encryptedMnemonic)
	if err != nil {
		return errors.Wrap(err, "save mnemonic failed")
	}

	fmt.Println("Your mnemonic is saved to file with password: ", password)
	return nil
}

// runLoadMnemonic 方法实现了从文件加载助记词的逻辑，包括从文件读取加密的助记词和解密助记词。
func (c *WalletCommand) runLoadMnemonic(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return errors.New("Please provide the file path and password")
	}

	store := storage.NewLocalStorage(args[0])
	encryptedMnemonic, err := store.Load()
	if err != nil {
		return errors.Wrap(err, "load mnemonic failed")
	}

	// decrypt the mnemonic with password
	mnemonic, err := utils.AesDecrypt(encryptedMnemonic, args[1])
	if err != nil {
		return errors.Wrap(err, "decrypt mnemonic failed")
	}

	c.mnemonic = string(mnemonic)
	fmt.Println("Your mnemonic is: ", c.mnemonic)
	return nil
}
