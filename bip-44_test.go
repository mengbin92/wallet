package wallet

import (
	"fmt"
	"strings"
	"testing"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/tyler-smith/go-bip39"
)

func generateMasterKey(seed []byte) (*hdkeychain.ExtendedKey, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}
	return masterKey, nil
}

func deriveChildKey(masterKey *hdkeychain.ExtendedKey, path string) (*hdkeychain.ExtendedKey, error) {
	segments := strings.Split(path, "/")
	var key *hdkeychain.ExtendedKey
	for _, segment := range segments {
		if segment == "m" {
			continue
		}

		var index uint32
		if strings.HasSuffix(segment, "'") {
			index = hdkeychain.HardenedKeyStart
			segment = strings.TrimSuffix(segment, "'")
		}

		i, err := parseUint32(segment)
		if err != nil {
			return nil, err
		}
		index += i

		key, err = masterKey.Derive(index)
		if err != nil {
			return nil, err
		}
	}
	return key, nil
}

// parseUint32 将字符串解析为 uint32
func parseUint32(s string) (uint32, error) {
	var n uint32
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func bip44DerivationPath(coinType uint32, accountIndex uint32) string {
	return fmt.Sprintf("m/44'/%d'/%d'", coinType, accountIndex)
}

func generateAddressAndPrivateKey(childKey *hdkeychain.ExtendedKey) (string, []byte, error) {
	privKey, err := childKey.ECPrivKey()
	if err != nil {
		return "", nil, err
	}

	pubKey, err := childKey.Neuter()
	if err != nil {
		return "", nil, err
	}
	address, err := pubKey.Address(&chaincfg.MainNetParams)
	if err != nil {
		return "", nil, err
	}

	privateKeyBytes := privKey.Serialize()
	return address.String(), privateKeyBytes, nil
}

func generateSeed() ([]byte, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, err
	}

	seed := bip39.NewSeed(mnemonic, "your_passphrase")
	return seed, nil
}

func TestBIP44(t *testing.T) {
	// 1. 生成种子
	seed, _ := generateSeed()

	// 2. 生成主私钥
	masterKey, _ := generateMasterKey(seed)

	// 3. 派生BIP-44路径
	coinType := uint32(0)     // Bitcoin
	accountIndex := uint32(0) // Account 0
	path := bip44DerivationPath(coinType, accountIndex)

	// 4. 派生子私钥
	childKey, _ := deriveChildKey(masterKey, path)

	// 5. 生成地址和私钥
	address, privateKeyBytes, _ := generateAddressAndPrivateKey(childKey)

	fmt.Printf("BIP-44 Address: %s\n", address)
	fmt.Printf("Private Key (hex): %x\n", privateKeyBytes)
}
