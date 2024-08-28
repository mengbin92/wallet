package command

import (
	"fmt"
	"strconv"

	"github.com/mengbin92/wallet/kms"
	"github.com/mengbin92/wallet/storage"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *WalletCommand) keyCmd() *cobra.Command {
	keyCmd := &cobra.Command{
		Use:   "key",
		Short: "Manage key",
		Long:  "Manage key, including create, list, import and export",
	}

	keyCmd.AddCommand(c.keyCreateCmd())
	keyCmd.AddCommand(c.keyListCmd())
	return keyCmd
}

func (c *WalletCommand) keyCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new key, example: ./wallet key create ./mnemonic.key password network account address_index",
		Long:  "Create a new key, example: ./wallet key create ./mnemonic.key password network account address_index",
		RunE:  c.runKeyCreateCmd,
	}
}

func (c *WalletCommand) runKeyCreateCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("key create")
	if c.masterKey == nil {
		err := c.runLoadMnemonic(cmd, args)
		if err != nil {
			return errors.Wrap(err, "load mnemonic failed")
		}
		masterKey, err := c.genMasterKey(args[1], args[2])
		if err != nil {
			return errors.Wrap(err, "generate master key failed")
		}
		c.masterKey = masterKey
	}

	account, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parse account failed")
	}

	addressIndex, err := strconv.ParseUint(args[4], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parse address index failed")
	}

	child,err := kms.DeriveChildKey(c.masterKey, 0, uint32(account), uint32(addressIndex))
	if err!= nil {
		return errors.Wrap(err, "derive child key failed")
	}

	wif,err := kms.GetWIFFromExtendedKey(child,args[2])
	if err!= nil {
		return errors.Wrap(err, "get wif failed")
	}
	fmt.Println("wif: ", wif.String())

	store := storage.NewLocalStorage(args[0])
	encryptedKey, err := utils.AesEncrypt([]byte(wif.String()), args[1])
	if err != nil {
		return errors.Wrap(err, "encrypt key failed")
	}
	err = store.SaveKey(encryptedKey)
	if err != nil {
		return errors.Wrap(err, "save key failed")
	}
	fmt.Println("key created successfully")
	return nil
}

func (c *WalletCommand) keyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all keys, example: ./wallet key list ./mnemonic.key password",
		Long:  "List all keys, example: ./wallet key list ./mnemonic.key password",
		RunE:  c.runListKeys,
	}
}

func (c *WalletCommand) runListKeys(cmd *cobra.Command, args []string) error {
	fmt.Println("key list")
	store := storage.NewLocalStorage(args[0])
	keys, err := store.ListKeys()
	if err != nil {
		return errors.Wrap(err, "list keys failed")
	}
	for _, key := range keys {
		decryptedKey, err := utils.AesDecrypt(key, args[1])
		if err != nil {
			return errors.Wrap(err, "decrypt key failed")
		}
		fmt.Println("key: ", string(decryptedKey))
	}
	return nil
}