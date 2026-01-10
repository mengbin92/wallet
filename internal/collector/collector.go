package collector

import (
	"log"
	"math/big"

	"github.com/mengbin92/wallet/internal/wallet"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

// CollectTokens 遍历 keystore 归集所有代币
func CollectTokens(rpcURL, mainKeystoreDir, keystoreDir, password, tokenAddr, targetAddr string, logger *log.Logger) error {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return errors.Wrapf(err, "failed to connect rpc (rpcURL: %s)", rpcURL)
	}
	defer client.Close()

	mainKey, err := wallet.LoadAllKeys(mainKeystoreDir, password, logger)
	if err != nil {
		return errors.Wrapf(err, "failed to load main key (mainKey: %s)", mainKeystoreDir)
	}

	keys, err := wallet.LoadAllKeys(keystoreDir, password, logger)
	if err != nil {
		return errors.Wrapf(err, "failed to load keys (keystoreDir: %s)", keystoreDir)
	}

	for _, key := range keys {
		addr := key.Address.Hex()
		balance, err := GetTokenBalance(client, tokenAddr, addr)
		if err != nil {
			logger.Printf("balanceOf %s failed: %v\n", addr, err)
			continue
		}

		if balance.Cmp(big.NewInt(0)) == 0 {
			logger.Printf("skip %s, balance=0\n", addr)
			continue
		}

		logger.Printf("collecting from %s, balance=%s", addr, new(big.Int).Div(balance, DECIMALS).String())

		tx, err := TransferToken(client, key, password, tokenAddr, targetAddr, mainKey[0].PrivateKey, logger)
		if err != nil {
			logger.Printf("transfer from %s failed: %v", addr, err)
			continue
		}

		logger.Printf("tx sent: %s\n", tx.Hash().Hex())
	}

	return nil
}