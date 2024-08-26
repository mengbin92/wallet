package command

import (
	"fmt"

	"github.com/mengbin92/wallet/kms"
	"github.com/mengbin92/wallet/storage"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// mnemonic command
func (c *WalletCommand) mnemonicCmd() *cobra.Command {
	genMnemonicCmd := &cobra.Command{
		Use:   "mnemonic",
		Short: "Manage mnenomic",
		Long:  "Manage mnenomic",
	}

	genMnemonicCmd.AddCommand(c.createMnemonic())
	// genMnemonicCmd.AddCommand(c.saveMnemonic())
	genMnemonicCmd.AddCommand(c.loadMnemonic())

	return genMnemonicCmd
}

// create new mnemonic command
func (c *WalletCommand) createMnemonic() *cobra.Command {
	createMnemonicCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new mnemonic and save it to file",
		Long:  "Create a new mnemonic and save it to file, example: ./wallet mnemonic create ./mnemonic.txt password",
		RunE:  c.runCreateMnemonic,
	}
	return createMnemonicCmd
}

// save mnemonic command
func (c *WalletCommand) saveMnemonic() *cobra.Command {
	saveMnemonicCmd := &cobra.Command{
		Use:   "save",
		Short: "Save a mnemonic to file",
		Long:  "Save a mnemonic to file",
		RunE:  c.runSaveMnemonic,
	}
	return saveMnemonicCmd
}

// load mnemonic command
func (c *WalletCommand) loadMnemonic() *cobra.Command {
	loadMnemonicCmd := &cobra.Command{
		Use:   "load",
		Short: "Load a mnemonic from file",
		Long:  "Load a mnemonic from file",
		RunE:  c.runLoadMnemonic,
	}
	return loadMnemonicCmd
}

func (c *WalletCommand) runCreateMnemonic(cmd *cobra.Command, args []string) error {
	fmt.Println("Create a new mnemonic")
	if len(args) < 1{
		return errors.New("Please provide the file path to save the mnemonic, e.g. ./mnemonic.txt")
	}
	mnemonic, err := kms.GenMnemonic()
	if err != nil {
		return errors.Wrap(err, "generate mnemonic failed")
	}
	fmt.Println("Your new mnemonic is: ", mnemonic)
	c.mnemonic = mnemonic

	// save the mnemonic to file
	c.runSaveMnemonic(cmd, args)
	return nil
}

func (c *WalletCommand) runSaveMnemonic(cmd *cobra.Command, args []string) error {
	fmt.Println("Save a mnemonic to file: ",c.mnemonic)
	// args[0] is the file path
	store := storage.NewLocalStorage(args[0])
	// args[1] is the password
	// if len(args) < 2, create a new password
	var password string
	var err error
	if len(args) ==2{
		password = args[1]
	} else{
		password,err = utils.CreatePassphrase(12)
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

func (c *WalletCommand) runLoadMnemonic(cmd *cobra.Command, args []string) error {
	fmt.Println("Load a mnemonic from file")
	// args[0] is the file path
	store := storage.NewLocalStorage(args[0])
	// args[1] is the password
	// load the encrypted mnemonic from file
	encryptedMnemonic,err := store.Load()
	if err != nil{
		return errors.Wrap(err, "load mnemonic failed")
	}

	// decrypt the mnemonic with password
	mnemonic, err := utils.AesDecrypt(encryptedMnemonic, args[1])
	if err !=nil{
		return errors.Wrap(err, "decrypt mnemonic failed")
	}
	c.mnemonic = string(mnemonic)
	fmt.Println("Your mnemonic is: ", c.mnemonic)
	return nil
}
