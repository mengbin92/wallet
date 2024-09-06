package command

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// balanceCmd 方法定义了一个获取钱包余额的命令，并添加了子命令来具体执行获取余额的操作。
func (c *WalletCommand) balanceCmd() *cobra.Command {
	balanceCmd := &cobra.Command{
		Use:   "balance",
		Short: "Get the balance of the wallet",
		Long:  "Get the balance of the wallet",
	}
	balanceCmd.AddCommand(c.getBalanceCmd())
	return balanceCmd
}

// getBalanceCmd 方法定义了一个具体的获取钱包余额的命令，包含了命令的使用说明和执行函数。
func (c *WalletCommand) getBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get the balance of the wallet",
		Long:  "Get the balance of the wallet, example: ./wallet balance network[testnet|mainnet] address",
		RunE:  c.runGetBalanceCmd,
	}
}

// runGetBalanceCmd 方法实现了获取钱包余额的具体逻辑，包括获取未花费交易输出（UTXOs）并计算余额。
func (c *WalletCommand) runGetBalanceCmd(cmd *cobra.Command, args []string) error {
	fmt.Println("get balance")
	var network, address string
	var err error

	if len(args) != 2 {
		// 未提供参数，需要手动输入
		network, err = askNetwork()
		if err != nil {
			return errors.Wrap(err, "ask network failed")
		}
		address, err = askAddress()
		if err != nil {
			return errors.Wrap(err, "ask address failed")
		}
	} else {
		network = args[0]
		address = args[1]
	}

	utxos, err := getUTXOs(address, network)
	if err != nil {
		return errors.Wrap(err, "get utxos failed")
	}
	balance := 0.0
	for _, utxo := range utxos {
		balance += utxo.Amount
	}
	fmt.Printf("Address: %s Balance: %.9f BTC\n", address, balance)
	return nil
}
