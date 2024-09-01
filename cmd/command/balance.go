package command

import (
	"fmt"

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

	utxos, err := getUTXOs(args[1], args[0])
	if err != nil {
		return errors.Wrap(err, "get utxos failed")
	}
	balance := 0.0
	for _, utxo := range utxos {
		balance += utxo.Amount
	}
	fmt.Printf("Address: %s Balance: %.9f BTC\n", args[1], balance)
	return nil
}
