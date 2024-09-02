package command

import (
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mengbin92/wallet/address"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *WalletCommand) txCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Transaction operations",
		Long:  "Transaction operations",
	}
	cmd.AddCommand(
		c.sendCmd(),
		c.getTxCmd(),
	)
	return cmd
}

func (c *WalletCommand) sendCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "send",
		Short: "Send bitcoins, example: ./wallet tx send ./key.key password network from to amount",
		Long:  "Send bitcoins, example: ./wallet tx send ./key.key password network from to amount",
		RunE:  c.runSendCmd,
	}
}

func (c *WalletCommand) runSendCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("send btc")
	var wif *btcutil.WIF
	var err error

	// check from address
	keys, err := listKeys(args[0])
	if err != nil {
		return errors.Wrap(err, "list keys failed")
	}
	for _, key := range keys {
		// 解密私钥
		decryptedKey, err := utils.BIP38Decrypt(key, args[1], args[2])
		if err != nil {
			return errors.Wrap(err, "decrypt key failed")
		}
		wif, err = btcutil.DecodeWIF(string(decryptedKey))
		if err != nil {
			return errors.Wrap(err, "decode wif failed")
		}
		addr, err := address.NewBTCAddressFromWIF(wif).GenBech32Address(utils.GetNetwork(args[2]))
		if err != nil {
			return errors.Wrap(err, "generate bech32 address failed")
		}
		if addr == args[3] {
			break
		}
	}
	if wif == nil {
		return errors.New("from address not found")
	}

	amount, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return errors.Wrap(err, "parse account failed")
	}
	// 构建交易输出
	txOut, _, err := buildTxOut(args[4], args[2], amount)
	if err != nil {
		return errors.Wrap(err, "build tx out failed")
	}

	// 构建交易输入
	msgTx, err := buildTxIn(wif, amount, txOut, args[2])
	if err != nil {
		return errors.Wrap(err, "build tx in failed")
	}

	// 发送交易
	txHash, err := client.SendRawTransaction(msgTx, false)
	if err != nil {
		return errors.Wrap(err, "send raw transaction failed")
	}

	fmt.Println("txHash:", txHash)

	return nil
}

func (c *WalletCommand) getTxCmd()*cobra.Command {
	return &cobra.Command{
		Use: "gettx",
		Short: "Get transaction by txHash, example: ./wallet tx gettx txHash",
		Long: "Get transaction by txHash, example: ./wallet tx gettx txHash",
		RunE: c.runGetTxCmd,
	}
}

func (c *WalletCommand) runGetTxCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("get tx")
	hash,err := chainhash.NewHashFromStr(args[0])
	if err != nil {
		return errors.Wrap(err, "parse txHash failed")
	}
	tx, err := client.GetRawTransaction(hash)
	if err != nil {
		return errors.Wrap(err, "get raw transaction failed")
	}
	fmt.Println(tx)
	return nil
}