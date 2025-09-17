package wallet

import (
	"github.com/mengbin92/wallet/internal/models"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	bip32 "github.com/tyler-smith/go-bip32"
	bip39 "github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic 生成 24 字助记词（256bits entropy）
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate entropy (bits: 256)")
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate mnemonic (entropy length: 256 bits)")
	}
	return mnemonic, nil
}

// DeriveByMnemonic 使用助记词派生 ETH 私钥/地址（BIP44）
func DeriveByMnemonic(chain, mnemonic, pwd string, index int) (*models.Address, error) {
	seed := bip39.NewSeed(mnemonic, pwd)
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive master key")
	}

	// m/44'/60'/0'/0/index
	pathParts := []uint32{
		44 + bip32.FirstHardenedChild,
		60 + bip32.FirstHardenedChild,
		0 + bip32.FirstHardenedChild,
		0,
		uint32(index),
	}
	key := masterKey
	for _, p := range pathParts {
		key, err = key.NewChildKey(p)
		if err != nil {
			return nil, errors.Wrap(err, "failed to derive child key")
		}
	}

	// key.Key is 32 bytes
	priv, err := crypto.ToECDSA(key.Key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert to ecdsa private key")
	}
	return &models.Address{
		Chain:      chain,
		Address:    crypto.PubkeyToAddress(priv.PublicKey).Hex(),
		PublicKey:  hex.EncodeToString(crypto.FromECDSAPub(&priv.PublicKey)),
		PrivateKey: hex.EncodeToString(crypto.FromECDSA(priv)),
		Path:       fmt.Sprintf("m/44'/60'/0'/0/%d", index),
	}, nil
}

func DeriveBatchEVM(chain, mnemonic, pwd string, count int) ([]*models.Address, error) {
	addrs := make([]*models.Address, 0, count)
	for i := 0; i < count; i++ {
		addr, err := DeriveByMnemonic(chain, mnemonic, pwd, i)
		if err != nil {
			return nil, errors.Wrap(err, "failed to derive address")
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// ImportPrivateKeyHex 将私钥 hex 转为 *ecdsa.PrivateKey（辅助）
func ImportPrivateKeyHex(privHex string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(privHex)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode private key hex")
	}
	return crypto.ToECDSA(b)
}
