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

const (
	// DefaultRequiredConfirmations is the default number of confirmations for token transfers
	DefaultRequiredConfirmations = 6
)

var (
	DECIMALS = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(18), nil)
)

// WeiToBNB 将 wei 转换为 BNB 字符串（保留6位小数）
func WeiToBNB(wei *big.Int) string {
	if wei == nil || wei.Sign() == 0 {
		return "0"
	}
	// 转换为浮点数并除以 10^18
	bnb := new(big.Float).SetInt(wei)
	bnb.Quo(bnb, new(big.Float).SetInt(DECIMALS))
	return bnb.Text('f', 6)
}

// WaitForConfirmationConfig holds configuration for transaction confirmation
type WaitForConfirmationConfig struct {
	RequiredConfirmations uint64
	CheckInterval         time.Duration
	MaxWaitTime           time.Duration
	ProgressInterval      int
}

// WaitForTransactionConfirmation waits for transaction confirmation with configurable options
func WaitForTransactionConfirmation(client *ethclient.Client, txHash common.Hash, cfg *WaitForConfirmationConfig, logger *log.Logger) error {
	if cfg == nil {
		// Default: 1 confirmation, 2s interval, 60s timeout
		cfg = &WaitForConfirmationConfig{
			RequiredConfirmations: 1,
			CheckInterval:         2 * time.Second,
			MaxWaitTime:           60 * time.Second,
			ProgressInterval:      5,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.MaxWaitTime)
	defer cancel()

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	startTime := time.Now()
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("transaction confirmation timeout after %v", cfg.MaxWaitTime)
		case <-ticker.C:
			attempts++
			elapsed := time.Since(startTime)

			// 获取交易收据
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err == nil && receipt != nil {
				// 交易已被打包，检查状态
				if receipt.Status == types.ReceiptStatusSuccessful {
					// 获取当前区块号，计算确认数
					currentBlock, err := client.BlockNumber(ctx)
					if err != nil {
						logger.Printf("无法获取当前区块号，但交易已成功: %s", txHash.Hex())
						return nil
					}

					confirmations := currentBlock - receipt.BlockNumber.Uint64()

					// 如果需要多个确认，显示进度
					if cfg.RequiredConfirmations > 1 {
						logger.Printf("交易确认进度: %s (确认数: %d/%d, 耗时: %v)",
							txHash.Hex(), confirmations, cfg.RequiredConfirmations, elapsed)

						if confirmations >= cfg.RequiredConfirmations {
							logger.Printf("交易确认完成: %s (确认数: %d)", txHash.Hex(), confirmations)
							return nil
						}
						// 继续等待更多确认
						continue
					} else {
						// 只需要1个确认，直接返回
						logger.Printf("交易确认成功: %s (耗时: %v, 确认数: %d)",
							txHash.Hex(), elapsed, confirmations)
						return nil
					}
				} else {
					return errors.Errorf("transaction failed with status: %d", receipt.Status)
				}
			} else if err != nil {
				// 交易还未被确认，继续等待
				if attempts%cfg.ProgressInterval == 0 {
					logger.Printf("等待交易确认中... (已等待: %v, 尝试次数: %d) - 交易尚未被打包", elapsed, attempts)
				}
			}
		}
	}
}

// Deprecated: Use WaitForTransactionConfirmation with config instead
// WaitForTransactionConfirmationSimple waits for transaction with default 1 confirmation
func WaitForTransactionConfirmationSimple(client *ethclient.Client, txHash common.Hash, maxWaitTime time.Duration, logger *log.Logger) error {
	return WaitForTransactionConfirmation(client, txHash, &WaitForConfirmationConfig{
		RequiredConfirmations: 1,
		CheckInterval:         2 * time.Second,
		MaxWaitTime:           maxWaitTime,
		ProgressInterval:      5,
	}, logger)
}

// Deprecated: Use WaitForTransactionConfirmation with config instead
// WaitForTransactionConfirmationWithBlocks waits for specified number of confirmations
func WaitForTransactionConfirmationWithBlocks(client *ethclient.Client, txHash common.Hash, requiredConfirmations uint64, maxWaitTime time.Duration, logger *log.Logger) error {
	return WaitForTransactionConfirmation(client, txHash, &WaitForConfirmationConfig{
		RequiredConfirmations: requiredConfirmations,
		CheckInterval:         3 * time.Second,
		MaxWaitTime:           maxWaitTime,
		ProgressInterval:      5,
	}, logger)
}

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
func EnsureGasFee(client *ethclient.Client, tokenAddr string, fromAddr, gasPayer common.Address, payerPriv *ecdsa.PrivateKey, logger *log.Logger, cfg *WaitForConfirmationConfig) error {
	// 保守估算：ERC20 转账一般 gasLimit ~ 65,000 (包含20%缓冲)
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to suggest gas price")
	}
	needWei := new(big.Int).Mul(big.NewInt(65000), gasPrice)

	balance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return errors.Wrap(err, "failed to get balance")
	}

	if balance.Cmp(needWei) >= 0 {
		return nil // 足够，不需要补充
	}

	logger.Printf("账户 %s 余额不足，需要补充 %s wei 作为 Gas 费用", fromAddr.Hex(), needWei.String())

	tx, err := TransferBNB(client, payerPriv, fromAddr, needWei, logger)
	if err != nil {
		return errors.Wrap(err, "failed to transfer gas fee")
	}

	// 使用传入的配置等待BNB转账确认
	logger.Printf("等待BNB转账确认: %s", tx.Hash().Hex())
	err = WaitForTransactionConfirmation(client, tx.Hash(), cfg, logger)
	if err != nil {
		return errors.Wrap(err, "BNB转账确认超时或失败")
	}

	return nil
}

// TransferToken 构造并发送 ERC20 转账交易
func TransferToken(client *ethclient.Client, ks *keystore.Key, password, tokenAddr, to string, gasPayerPriv *ecdsa.PrivateKey, logger *log.Logger, confirmCfg *WaitForConfirmationConfig) (*types.Transaction, error) {
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

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get chainID")
	}

	// 预估gas
	parsedABI, err := abi.JSON(strings.NewReader(ERC20.ERC20ABI))
	if err != nil {
		logger.Println("abi.JSON failed: ", err.Error())
		return nil, errors.Wrap(err, "failed to parse ABI")
	}
	data, err := parsedABI.Pack("transfer", toAddr, tokenBalance)
	if err != nil {
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
	// 计算带缓冲的gas费用
	gasWithBuffer := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(estimatedGas)), big.NewInt(230)), big.NewInt(100))
	needWei := new(big.Int).Mul(gasWithBuffer, gasPrice)
	logger.Printf("ERC20 transfer estimated gas cost: %s (带230%%缓冲)\n", needWei.String())

	// 检查fromAddr的BNB余额是否足够
	bnbBalance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get BNB balance")
	}

	if bnbBalance.Cmp(needWei) < 0 {
		// 计算差额：需要多少 - 现有多少 = 差额
		deficit := new(big.Int).Sub(needWei, bnbBalance)
		logger.Printf("账户 %s BNB余额不足，现有: %s wei，需要: %s wei，差额: %s wei",
			fromAddr.Hex(), bnbBalance.String(), needWei.String(), deficit.String())

		tx, err := TransferBNB(client, gasPayerPriv, fromAddr, deficit, logger)
		if err != nil {
			return nil, errors.Wrap(err, "failed to transfer gas fee")
		}

		// 使用传入的配置等待BNB转账确认
		logger.Printf("等待Gas费用差额转账确认: %s", tx.Hash().Hex())
		if confirmCfg == nil {
			// 默认配置：6个确认，60秒超时
			confirmCfg = &WaitForConfirmationConfig{
				RequiredConfirmations: DefaultRequiredConfirmations,
				CheckInterval:         3 * time.Second,
				MaxWaitTime:           60 * time.Second,
				ProgressInterval:      5,
			}
		}
		err = WaitForTransactionConfirmation(client, tx.Hash(), confirmCfg, logger)
		if err != nil {
			return nil, errors.Wrap(err, "Gas费用差额转账确认超时或失败")
		}
	} else {
		logger.Printf("账户 %s BNB余额充足: %s wei，无需补充", fromAddr.Hex(), bnbBalance.String())
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transactor")
	}
	auth.GasPrice = gasPrice
	// 使用已经计算好的带缓冲的gas值
	auth.GasLimit = gasWithBuffer.Uint64()

	// 发起转账
	tokenTx, err := instance.Transfer(auth, toAddr, tokenBalance)
	if err != nil {
		return nil, errors.Wrap(err, "failed to send token transfer tx")
	}

	logger.Printf("Token 转账已发送: %s -> %s, amount: %s, txHash: %s",
		fromAddr.Hex(), to, new(big.Int).Div(tokenBalance, DECIMALS).String(), tokenTx.Hash().Hex(),
	)
	return tokenTx, nil
}