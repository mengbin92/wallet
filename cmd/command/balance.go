package command

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *WalletCommand) balanceCmd() *cobra.Command {
	balanceCmd := &cobra.Command{
		Use:   "balance",
		Short: "Get the balance of the wallet",
		Long:  "Get the balance of the wallet",
	}
	balanceCmd.AddCommand(c.getBalanceCmd())
	return balanceCmd
}

func (c *WalletCommand) getBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get the balance of the wallet",
		Long:  "Get the balance of the wallet, example: ./wallet balance network[testnet|mainnet] address",
		RunE:  c.runGetBalanceCmd,
	}
}

func (c *WalletCommand) runGetBalanceCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("get balance")
	// 解析比特币地址
	address, err := btcutil.DecodeAddress(args[1], utils.GetNetwork(args[0]))
	if err != nil {
		return errors.Wrap(err, "解析比特币地址失败")
	}

	// 使用SearchRawTransactionsVerbose获取与地址相关的所有交易
	transactions, err := client.SearchRawTransactionsVerbose(address, 0, 100, true, false, nil)
	if err != nil {
		return errors.Wrap(err, "获取与地址相关的所有交易失败")
	}

	var balance int64
	// 遍历所有交易
	for _, tx := range transactions {
		// 将交易ID字符串转换为链哈希对象
		txid, err := chainhash.NewHashFromStr(tx.Txid)
		if err != nil {
			return errors.Wrap(err, "将交易ID字符串转换为链哈希对象失败")
		}
		// 遍历交易的输出
		for _, vout := range tx.Vout {
			// 检查输出地址是否是我们关心的地址
			if vout.ScriptPubKey.Address != args[1] {
				continue
			}

			// 使用GetTxOut方法获取交易输出，确认该输出是否未花费
			utxo, err := client.GetTxOut(txid, vout.N, true)
			if err != nil {
				return errors.Wrap(err, "获取交易输出失败")
			}

			// 如果交易输出未花费，则将其添加到UTXO切片中
			if utxo != nil {
				balance += int64(vout.Value * decimals)
			}
		}
	}
	fmt.Printf("address: %s balance: %d Satoshi\n", args[1], balance)
	return nil
}
