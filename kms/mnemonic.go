package kms

import (
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
	"github.com/tyler-smith/go-bip39"
)

// GenMnemonic 生成助记词
func GenMnemonic() (string, error) {
	// 生成随机熵（Entropy）
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate entropy")
	}

	return bip39.NewMnemonic(entropy)
}

// ExportMnemonic 加密导出助记词
func ExportMnemonic(mnemonic, password string) (string, error) {
	encryptData, err := utils.AesEncrypt([]byte(mnemonic), password)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt private key")
	}
	return encryptData, nil
}

// ImportMnemonic 解密导入助记词
func ImportMnemonic(encryptStr, password string) (string, error) {
	decryptData, err := utils.AesDecrypt(encryptStr, password)
	if err != nil {
		return "", errors.Wrap(err, "failed to decrypt private key")
	}
	return string(decryptData), nil
}

// GenMasterKey 生成钱包主秘钥
func GenMasterKey(mnemonic, password, network string) (*hdkeychain.ExtendedKey, error) {
	seed := bip39.NewSeed(mnemonic, password)

	return hdkeychain.NewMaster(seed, utils.GetNetwork(network))
}

// DeriveChildKey 依据BIP-44路径生成子秘钥
// m / purpose' / coin' / account' / change / address_index
func DeriveChildKey(masterKey *hdkeychain.ExtendedKey, coin, account, address_index uint32) (*hdkeychain.ExtendedKey, error) {
	purpose, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive purpose key")
	}

	coinType, err := purpose.Derive(hdkeychain.HardenedKeyStart + coin)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive coinType key")
	}

	accountKey, err := coinType.Derive(hdkeychain.HardenedKeyStart + account)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive account key")
	}

	change, err := accountKey.Derive(0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive change key")
	}

	address, err := change.Derive(address_index)
	if err != nil {
		return nil, errors.Wrap(err, "failed to derive address key")
	}
	return address, nil
}

// GetWIFFromExtendedKey 从扩展秘钥获取Wallet Import Format (WIF)
func GetWIFFromExtendedKey(extendedKey *hdkeychain.ExtendedKey, network string) (*btcutil.WIF, error) {
	privateKey, err := extendedKey.ECPrivKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get private key")
	}

	return btcutil.NewWIF(privateKey, utils.GetNetwork("mainnet"), true)
}
