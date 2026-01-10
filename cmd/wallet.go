package main

import (
	"fmt"
	"log"
	"os"

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

			// Securely handle mnemonic output
			if password != "" {
				encMnemonic, err := utils.AesEncrypt([]byte(mnemonic), password)
				if err != nil {
					return err
				}
				// Only show truncated version in logs
				truncated := encMnemonic
				if len(encMnemonic) > 20 {
					truncated = encMnemonic[:20] + "..."
				}
				log.Printf("[INFO] Encrypted mnemonic: %s\n", truncated)
				log.Printf("[INFO] Full encrypted mnemonic stored to file: mnemonic.enc\n")
				// Save encrypted mnemonic to file
				if err := os.WriteFile("mnemonic.enc", []byte(encMnemonic), 0600); err != nil {
					log.Printf("[WARN] Failed to save encrypted mnemonic to file: %v\n", err)
				}
			} else {
				// Output to stderr with security warning (not to logs)
				fmt.Fprintf(os.Stderr, "\n╔════════════════════════════════════════════════════════════╗\n")
				fmt.Fprintf(os.Stderr, "║  ⚠️  WARNING: MNEMONIC NOT ENCRYPTED - STORE SECURELY!      ║\n")
				fmt.Fprintf(os.Stderr, "╚════════════════════════════════════════════════════════════╝\n\n")
				fmt.Fprintf(os.Stderr, "Mnemonic: %s\n\n", mnemonic)
				fmt.Fprintf(os.Stderr, "📝 IMPORTANT: Write this down and store it securely.\n")
				fmt.Fprintf(os.Stderr, "   Anyone with access to this phrase can control your wallets.\n")
				fmt.Fprintf(os.Stderr, "   Do NOT share it with anyone or store it in logs.\n\n")
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
