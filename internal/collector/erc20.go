package collector

import (
	"context"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	ERC20 "github.com/mengbin92/wallet/pkg/contracts/erc20"
)

// GetTokenBalance 查询 ERC20 余额
func GetTokenBalance(client *ethclient.Client, tokenAddr, owner string) (*big.Int, error) {
	contract, err := ERC20.NewERC20(common.HexToAddress(tokenAddr), client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to bind contract (tokenAddr: %s)", tokenAddr)
	}
	balance, err := contract.BalanceOf(&bind.CallOpts{}, common.HexToAddress(owner))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get balance (owner: %s)", owner)
	}
	return balance, nil
}

// TransferToken 构造并发送 ERC20 转账交易
func TransferToken(client *ethclient.Client, ks *keystore.Key, password, tokenAddr, to string, amount *big.Int, logger *log.Logger) (*types.Transaction, error) {
	// 1. 解密私钥
	privKey := ks.PrivateKey
	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)

	// 2. 查询 ETH/BNB 余额
	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get balance (owner: %s)", fromAddr.String())
	}
	logger.Printf("Main currency balance (for gas): %s\n", balance.String())

	// 3. 代币余额检查
	tokenAddress := common.HexToAddress(tokenAddr)
	instance, err := ERC20.NewERC20(tokenAddress, client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to bind contract (tokenAddr: %s)", tokenAddr)
	}

	tokenBalance, err := instance.BalanceOf(&bind.CallOpts{}, fromAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token balance")
	}
	logger.Printf("token balance: %s\n", tokenBalance.String())

	if tokenBalance.Cmp(amount) < 0 {
		return nil, errors.Wrapf(err, "Insufficient balance: Needs %s, but only %s available", amount.String(), tokenBalance.String())
	}

	// 4. 构造交易
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chainID")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transactor")
	}

	// 配置 gasPrice
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get gas price")
	}
	auth.GasPrice = gasPrice

	toAddr := common.HexToAddress(to)

	// 5. 发送交易，节点自动估算gas
	tx, err := instance.Transfer(auth, toAddr, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send tx")
	}

	logger.Printf("tx had been sent with hash: %s\n", tx.Hash().Hex())
	return tx, nil
}
