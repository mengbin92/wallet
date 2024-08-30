package command

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/mengbin92/wallet/address"
	"github.com/mengbin92/wallet/storage"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// addressCmd 地址命令主入口，包含创建地址和查看地址列表功能
func (c *WalletCommand) addressCmd() *cobra.Command {
	addressCmd := &cobra.Command{
		Use:   "address",
		Short: "Manage btc address",
		Long:  "Manage btc address",
	}
	addressCmd.AddCommand(
		c.newAddressCmd(),
		c.listAddressCmd(),
	)
	return addressCmd
}

// newAddressCmd 从提供 wif 私钥创建 btc 地址
func (c *WalletCommand) newAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new",
		Short: "Create a new address from wif key, example: ./wallet address create wif network[testnet|mainnet]",
		Long:  "Create a new address from wif key, example: ./wallet address create wif network[testnet|mainnet]",
		RunE:  c.runNewAddressCmd,
	}
}

// runNewAddressCmd 从提供 wif 私钥创建 btc 地址
func (c *WalletCommand) runNewAddressCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("new btc address from wif key")
	// TODO: 校验参数合法性
	// 校验并解析 wif 私钥
	wif, err := btcutil.DecodeWIF(string(args[0]))
	if err != nil {
		return errors.Wrap(err, "decode wif failed")
	}
	// 生成 bech32 地址
	bech32Addr, err := address.NewBTCAddressFromWIF(wif).GenBech32Address(utils.GetNetwork(args[1]))
	if err != nil {
		return errors.Wrap(err, "generate bech32 address failed")
	}
	fmt.Println("address: ", bech32Addr)
	return nil
}

// listAddressCmd 列出所有地址，需要提供 key 文件和密码
func (c *WalletCommand) listAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all addresses, example: ./wallet address list ./key.key password network[testnet|mainnet]",
		Long:  "List all addresses, example: ./wallet address list ./key.key password network[testnet|mainnet]",
		RunE:  c.runListAddressCmd,
	}
}

// runListAddressCmd 列出所有地址，需要提供 key 文件和密码
func (c *WalletCommand) runListAddressCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("address list")
	// TODO: 校验参数合法性
	// 解析 key 文件
	store := storage.NewLocalStorage(args[0])
	// 获取 key 文件中所有的私钥
	keys, err := store.ListKeys()
	if err != nil {
		return errors.Wrap(err, "list keys failed")
	}
	for _, key := range keys {
		// 解密私钥
		decryptedKey, err := utils.AesDecrypt(key, args[1])
		if err != nil {
			return errors.Wrap(err, "decrypt key failed")
		}
		fmt.Println("key: ", string(decryptedKey))
		wif, err := btcutil.DecodeWIF(string(decryptedKey))
		if err != nil {
			return errors.Wrap(err, "decode wif failed")
		}
		// 生成 bech32 地址
		bech32Addr, err := address.NewBTCAddressFromWIF(wif).GenBech32Address(utils.GetNetwork(args[2]))
		if err != nil {
			return errors.Wrap(err, "generate bech32 address failed")
		}
		fmt.Println("address: ", bech32Addr)
	}
	return nil
}
