package wallet

import (
	"fmt"
	"log"
	"testing"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/mengbin92/wallet/address"
	"github.com/tyler-smith/go-bip39"
)

func TestMnemonic(t *testing.T) {
	// 1. 生成随机熵（Entropy）
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		log.Fatalf("Failed to generate entropy: %v", err)
	}

	// 2. 生成助记词
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		log.Fatalf("Failed to generate mnemonic: %v", err)
	}

	fmt.Printf("Mnemonic: %s\n", mnemonic)

	// 3. 从助记词生成种子（Seed）
	seed := bip39.NewSeed(mnemonic, "your_passphrase") // 第二个参数是一个可选的密码短语
	fmt.Printf("Seed: %x\n", seed)

	// 4. 从种子生成主密钥
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		log.Fatalf("Failed to create master key: %v", err)
	}

	fmt.Printf("Master Key: %v\n", masterKey)

	// 5. 派生子密钥并生成比特币地址
	for i := 0; i < 5; i++ {
		childKey, err := masterKey.Derive(uint32(i))
		if err != nil {
			log.Fatalf("Failed to derive child key: %v", err)
		}
		childKey.Neuter()

		privKey, err := childKey.ECPrivKey()
		if err != nil {
			log.Fatalf("Failed to get private key: %v", err)
		}

		wif, err := btcutil.NewWIF(privKey, &chaincfg.MainNetParams, true)
		if err != nil {
			log.Fatalf("Failed to get WIF: %v", err)
		}

		btcAddr := address.NewBTCAddressFromWIF(wif)

		// 生成比特币地址
		// BTC Pay-to-Pubkey-Hash
		p2pkh,_ := btcAddr.GenP2PKHAddress(&chaincfg.MainNetParams)
		fmt.Printf("P2PKH address %d: %s\n", i, p2pkh)

		// BTC SegWit address
		bech32,_ := btcAddr.GenBech32Address(&chaincfg.MainNetParams)
		fmt.Printf("Bech32 address %d: %s\n", i, bech32)

	}
}
