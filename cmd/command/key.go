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

// keyCmd 定义了一个用于管理密钥的命令，包括创建、列出、导入和导出密钥
func (c *WalletCommand) keyCmd() *cobra.Command {
	keyCmd := &cobra.Command{
		Use:   "key",
		Short: "Manage key",
		Long:  "Manage key, including create, list",
	}

	keyCmd.AddCommand(c.keyCreateCmd())
	keyCmd.AddCommand(c.keyListCmd())
	keyCmd.AddCommand(c.importKeyCmd())
	return keyCmd
}

// keyCreateCmd 定义了一个用于创建新密钥的命令
func (c *WalletCommand) keyCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new key, example: ./wallet key create ./key.key password network account address_index",
		Long:  "Create a new key, example: ./wallet key create ./key.key password network account address_index",
		RunE:  c.runKeyCreateCmd,
	}
}

// runKeyCreateCmd 执行创建新密钥的命令
// 此代码段负责创建一个密钥，并将其存储在本地存储中。
// 首先，它会检查主密钥是否已经存在，如果不存在，则加载助记词并生成主密钥。
// 然后，解析账户和地址索引，并从主密钥派生子密钥。
// 接着，将子密钥转换为WIF格式，并使用AES加密算法对其进行加密。
// 最后，将加密后的密钥保存到本地存储中，并输出成功信息。
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

	child, err := kms.DeriveChildKey(c.masterKey, 0, uint32(account), uint32(addressIndex))
	if err != nil {
		return errors.Wrap(err, "derive child key failed")
	}

	wif, err := kms.GetWIFFromExtendedKey(child, args[2])
	if err != nil {
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

// keyListCmd 定义了一个用于列出所有密钥的命令
func (c *WalletCommand) keyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all keys, example: ./wallet key list ./key.key password",
		Long:  "List all keys, example: ./wallet key list ./key.key password",
		RunE:  c.runListKeys,
	}
}

// runListKeys 执行列出所有密钥的命令
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

// importKeyCmd 定义了一个用于导入密钥的命令
func (c *WalletCommand) importKeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import a key, example: ./wallet key import ./key.key password wif",
		Long:  "Import a key, example: ./wallet key import ./key.key password wif",
		RunE:  c.runImportKeyCmd,
	}
}

// runImportKeyCmd 执行导入密钥的命令
func (c *WalletCommand) runImportKeyCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("key import")
	store := storage.NewLocalStorage(args[0])
	encryptedKey, err := utils.AesEncrypt([]byte(args[2]), args[1])
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
