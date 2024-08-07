package address

import (
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
)

type BTCAddress struct {
	key *btcutil.WIF
}

func NewBTCAddressFromWIF(wif *btcutil.WIF) *BTCAddress {
	return &BTCAddress{
		key: wif,
	}
}

func NewBTCAddress(param *chaincfg.Params) (*BTCAddress, error) {
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate private key")
	}

	wif, err := btcutil.NewWIF(privateKey, param, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate WIF")
	}

	return &BTCAddress{
		key: wif,
	}, nil
}

// GenP2PKAddress Generates the BTC Pay-to-Pubkey address
func (k *BTCAddress) GenP2PKAddress(param *chaincfg.Params) (string, error) {
	address, err := btcutil.NewAddressPubKey(k.key.SerializePubKey(), param)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate P2PK address")
	}
	// AddressPubKey.EncodeAddress 将公钥的字符串编码返回为 pay-to-pubkey-hash
	// so we can get the same address string with GenP2PKHAddress
	return address.EncodeAddress(), nil
}

// GenP2PKHAddress Generates the BTC Pay-to-Pubkey-Hash
func (k *BTCAddress) GenP2PKHAddress(param *chaincfg.Params) (string, error) {
	address, err := btcutil.NewAddressPubKeyHash(btcutil.Hash160(k.key.SerializePubKey()), param)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate P2PKH address")
	}
	return address.EncodeAddress(), nil
}

// GenBech32Address Generates the BTC SegWit address
func (k *BTCAddress) GenBech32Address(param *chaincfg.Params) (string, error) {
	address, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(k.key.SerializePubKey()), param)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate P2PKH address")
	}
	return address.EncodeAddress(), nil
}

func (k *BTCAddress) ExportPrivateKey(pwd string) (string, error) {
	encryptData, err := utils.AesEncrypt([]byte(k.key.String()), pwd)
	if err != nil {
		return "", errors.Wrap(err, "failed to encrypt private key")
	}
	return encryptData, nil
}

func (k *BTCAddress) LoadPrivateKey(encryptStr, pwd string) error {
	decryptData, err := utils.AesDecrypt(encryptStr, pwd)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt private key")
	}
	wif, err := btcutil.DecodeWIF(string(decryptData))
	if err != nil {
		return errors.Wrap(err, "failed to decode WIF")
	}
	k.key = wif
	return nil
}
