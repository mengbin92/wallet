package collector

import (
	"context"
	"crypto/ecdsa"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"

	ERC20 "github.com/mengbin92/wallet/pkg/contracts/erc20"
)

var (
	instance *ERC20.ERC20
)

// GetTokenBalance 查询 ERC20 余额
func GetTokenBalance(client *ethclient.Client, tokenAddr, owner string) (*big.Int, error) {
	contract, err := ERC20.NewERC20(common.HexToAddress(tokenAddr), client)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to bind contract (tokenAddr: %s)", tokenAddr)
	}
	return contract.BalanceOf(&bind.CallOpts{}, common.HexToAddress(owner))
}

// TransferBNB 主币转账 (BNB/ETH)
func TransferBNB(client *ethclient.Client, priv *ecdsa.PrivateKey, to common.Address, amount *big.Int, logger *log.Logger) (*types.Transaction, error) {
	fromAddr := crypto.PubkeyToAddress(priv.PublicKey)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nonce")
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to suggest gas price")
	}

	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get balance")
	}
	if balance.Cmp(amount) < 0 {
		return nil, errors.Errorf("insufficient balance: need %s, have %s", amount, balance)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chainID")
	}

	tx := types.NewTransaction(nonce, to, amount, 21000, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), priv)
	if err != nil {
		return nil, errors.Wrap(err, "failed to sign tx")
	}

	if err := client.SendTransaction(context.Background(), signedTx); err != nil {
		return nil, errors.Wrap(err, "failed to send tx")
	}

	logger.Printf("BNB 转账已发送: %s -> %s, amount: %s, txHash: %s",
		fromAddr.Hex(), to.Hex(), amount, signedTx.Hash().Hex(),
	)
	return signedTx, nil
}

// EnsureGasFee 确保 fromAddr 有足够的主币支付一次 ERC20 转账 Gas
func EnsureGasFee(client *ethclient.Client, tokenAddr string, fromAddr, gasPayer common.Address, payerPriv *ecdsa.PrivateKey, logger *log.Logger) error {
	// 粗略估算：ERC20 转账一般 gasLimit ~ 50,000
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to suggest gas price")
	}
	needWei := new(big.Int).Mul(big.NewInt(50000), gasPrice)

	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get balance")
	}

	if balance.Cmp(needWei) >= 0 {
		return nil // 足够，不需要补充
	}

	_, err = TransferBNB(client, payerPriv, fromAddr, needWei, logger)
	if err != nil {
		return errors.Wrap(err, "failed to transfer gas fee")
	}
	// 等待交易确认
	time.Sleep(5 * time.Second)
	return nil
}

// TransferToken 构造并发送 ERC20 转账交易
func TransferToken(client *ethclient.Client, ks *keystore.Key, password, tokenAddr, to string, amount *big.Int, gasPayerPriv *ecdsa.PrivateKey, logger *log.Logger) (*types.Transaction, error) {
	privKey := ks.PrivateKey
	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)
	toAddr := common.HexToAddress(to)

	// 检查代币余额
	instance, err := ERC20.NewERC20(common.HexToAddress(tokenAddr), client)
	if err != nil {
		logger.Println("NewERC20 failed: ", err.Error())
		return nil, errors.Wrapf(err, "failed to bind contract (tokenAddr: %s)", tokenAddr)
	}
	tokenBalance, err := instance.BalanceOf(&bind.CallOpts{}, fromAddr)
	if err != nil {
		logger.Println("BalanceOf failed: ", err.Error())
		return nil, errors.Wrap(err, "failed to get token balance")
	}

	if tokenBalance.Cmp(amount) < 0 {
		logger.Printf("insufficient token balance: need %d, have %d\n", amount, tokenBalance)
		return nil, errors.Errorf("insufficient token balance: need %s, have %s", amount, tokenBalance)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chainID")
	}

	// 预估gas
	parsedABI, err := abi.JSON(strings.NewReader(ERC20.ERC20ABI))
	if err != nil{
		logger.Println("abi.JSON failed: ", err.Error())
		return nil, errors.Wrap(err, "failed to parse ABI")
	}
	data,err := parsedABI.Pack("transfer", toAddr, amount)
	if err != nil{
		logger.Println("abi.Pack failed: ", err.Error())
		return nil, errors.Wrap(err, "failed to pack ABI")
	}
	msg := ethereum.CallMsg{
		From: fromAddr,
		To:   &toAddr,
		Data: data, // ABI 编码的数据
	}
	estimatedGas, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to estimate gas for ERC20 transfer")
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to suggest gas price")
	}
	needWei := new(big.Int).Mul(big.NewInt(int64(estimatedGas)), gasPrice)
	logger.Printf("ERC20 transfer estimated gas cost: %s\n", needWei.String())

	_, err = TransferBNB(client, gasPayerPriv, toAddr, needWei, logger)
	if err != nil {
		return nil, errors.Wrap(err, "failed to transfer gas fee")
	}
	time.Sleep(5 * time.Second)

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transactor")
	}
	auth.GasPrice = gasPrice
	auth.GasLimit = estimatedGas

	// 发起转账
	tx, err := instance.Transfer(auth, toAddr, amount)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send token transfer tx")
	}

	logger.Printf("Token 转账已发送: %s -> %s, amount: %s, txHash: %s",
		fromAddr.Hex(), to, amount, tx.Hash().Hex(),
	)
	return tx, nil
}