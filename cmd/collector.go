package main

import (
	"log"

	"github.com/mengbin92/wallet/internal/collector"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCollectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect ERC20 tokens from all accounts in keystore",
	}

	// -------------------- 注册子命令 --------------------
	cmd.AddCommand(collectCmd())
	return cmd
}

func collectCmd() *cobra.Command {
	var (
		rpcURL      string
		keystoreDir string
		password    string
		tokenAddr   string
		targetAddr  string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run token collection from all keystore accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenAddr == "" || targetAddr == "" {
				return errors.Errorf("token address and target address are required (token: %s, target: %s, use --token and --to flags)", tokenAddr, targetAddr)
			}
			return collector.CollectTokens(rpcURL, keystoreDir, password, tokenAddr, targetAddr, log.Default())
		},
	}

	cmd.Flags().StringVar(&rpcURL, "rpc", "https://bsc-dataseed.binance.org", "RPC endpoint")
	cmd.Flags().StringVar(&keystoreDir, "keystore", "keystore", "keystore directory")
	cmd.Flags().StringVar(&password, "password", "", "keystore password")
	cmd.Flags().StringVar(&tokenAddr, "token", "", "token contract address")
	cmd.Flags().StringVar(&targetAddr, "to", "", "target address to collect tokens into")

	return cmd
}
