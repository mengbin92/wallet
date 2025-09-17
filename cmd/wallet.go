package main

import (
	"log"

	"github.com/mengbin92/wallet/internal/wallet"
	"github.com/mengbin92/wallet/utils"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewWalletCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wallet",
		Short: "Manage wallets (create)",
	}

	// -------------------- 注册子命令 --------------------
	cmd.AddCommand(createCmd())
	return cmd
}

func createCmd() *cobra.Command {
	// -------------------- create 子命令 --------------------
	var count int
	var out string
	var chain string
	var mnemonic string
	var password string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new wallets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if mnemonic == "" {
				var err error
				mnemonic, err = wallet.GenerateMnemonic()
				if err != nil {
					return errors.Wrapf(err, "failed to generate mnemonic (count: %d, chain: %s)", count, chain)
				}
			}

			// 加密助记词
			if password != "" {
				encMnemonic, err := utils.AesEncrypt([]byte(mnemonic), password)
				if err != nil {
					return err
				}
				log.Printf("[INFO] Encrypted mnemonic: %s\n", encMnemonic)
			} else {
				log.Printf("[INFO] Mnemonic: %s\n", mnemonic)
			}

			addresses, err := wallet.DeriveBatchEVM(chain, mnemonic, password, count)
			if err != nil {
				return errors.Wrap(err, "failed to derive addresses")
			}

			if err := wallet.SaveToKeystore(addresses, password, out); err != nil {
				return errors.Wrap(err, "failed to save to keystore")
			}
			log.Printf("[SUCCESS] Created %d addresses on chain %s, saved to %s\n", count, chain, out)
			return nil
		},
	}

	cmd.Flags().IntVarP(&count, "count", "c", 1, "Number of addresses to create")
	cmd.Flags().StringVarP(&out, "out", "o", "keystore", "v3 keystore output directory")
	cmd.Flags().StringVarP(&chain, "chain", "", "bsc", "Chain type (eth/bsc)")
	cmd.Flags().StringVarP(&mnemonic, "mnemonic", "m", "", "Use existing mnemonic")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Password to encrypt mnemonic/private key")

	return cmd
}
