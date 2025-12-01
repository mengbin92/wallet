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
	REQUIREDCONFIRMATIONS = 6
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

// WaitForTransactionConfirmation 等待交易确认，支持超时和进度反馈
func WaitForTransactionConfirmation(client *ethclient.Client, txHash common.Hash, maxWaitTime time.Duration, logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second) // 每2秒检查一次
	defer ticker.Stop()

	startTime := time.Now()
	attempts := 0
	var receipt *types.Receipt
	var receiptErr error

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("transaction confirmation timeout after %v", maxWaitTime)
		case <-ticker.C:
			attempts++
			elapsed := time.Since(startTime)

			// 获取交易收据
			receipt, receiptErr = client.TransactionReceipt(ctx, txHash)
			if receiptErr == nil && receipt != nil {
				// 交易已被打包，检查状态
				if receipt.Status == types.ReceiptStatusSuccessful {
					// 获取当前区块号，计算确认数
					currentBlock, err := client.BlockNumber(ctx)
					if err != nil {
						logger.Printf("无法获取当前区块号，但交易已成功: %s", txHash.Hex())
						return nil
					}

					confirmations := currentBlock - receipt.BlockNumber.Uint64()
					logger.Printf("交易确认成功: %s (耗时: %v, 确认数: %d)",
						txHash.Hex(), elapsed, confirmations)

					// 对于BNB转账，1个确认通常就足够了
					// 对于大额转账，可以考虑等待更多确认
					return nil
				} else {
					return errors.Errorf("transaction failed with status: %d", receipt.Status)
				}
			} else if receiptErr != nil {
				// 交易还未被确认，继续等待
				// 提供进度反馈
				if attempts%5 == 0 { // 每10秒提供一次进度反馈
					logger.Printf("等待交易确认中... (已等待: %v, 尝试次数: %d) - 交易尚未被打包", elapsed, attempts)
				}
			}
		}
	}
}

// WaitForTransactionConfirmationWithBlocks 等待交易确认，支持自定义确认数
func WaitForTransactionConfirmationWithBlocks(client *ethclient.Client, txHash common.Hash, requiredConfirmations uint64, maxWaitTime time.Duration, logger *log.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxWaitTime)
	defer cancel()

	ticker := time.NewTicker(3 * time.Second) // 每3秒检查一次
	defer ticker.Stop()

	startTime := time.Now()
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("transaction confirmation timeout after %v", maxWaitTime)
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
					logger.Printf("交易确认进度: %s (确认数: %d/%d, 耗时: %v)",
						txHash.Hex(), confirmations, requiredConfirmations, elapsed)

					if confirmations >= requiredConfirmations {
						logger.Printf("交易确认完成: %s (确认数: %d)", txHash.Hex(), confirmations)
						return nil
					}
				} else {
					return errors.Errorf("transaction failed with status: %d", receipt.Status)
				}
			} else if err != nil {
				// 交易还未被确认，继续等待
				if attempts%5 == 0 { // 每15秒提供一次进度反馈
					logger.Printf("等待交易确认中... (已等待: %v, 尝试次数: %d) - 交易尚未被打包", elapsed, attempts)
				}
			}
		}
	}
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

	logger.Printf("BNB 转账已发送: %s -> %s, amount: %s BNB, txHash: %s",
		fromAddr.Hex(), to.Hex(), WeiToBNB(amount), signedTx.Hash().Hex(),
	)
	return signedTx, nil
}

// EnsureGasFee 确保 fromAddr 有足够的主币支付一次 ERC20 转账 Gas
func EnsureGasFee(client *ethclient.Client, tokenAddr string, fromAddr, gasPayer common.Address, payerPriv *ecdsa.PrivateKey, logger *log.Logger) error {
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

	logger.Printf("账户 %s 余额不足，需要补充 %s BNB 作为 Gas 费用", fromAddr.Hex(), WeiToBNB(needWei))

	tx, err := TransferBNB(client, payerPriv, fromAddr, needWei, logger)
	if err != nil {
		return errors.Wrap(err, "failed to transfer gas fee")
	}

	// 等待BNB转账确认，最多等待60秒
	logger.Printf("等待BNB转账确认: %s", tx.Hash().Hex())
	err = WaitForTransactionConfirmation(client, tx.Hash(), 60*time.Second, logger)
	if err != nil {
		return errors.Wrap(err, "BNB转账确认超时或失败")
	}

	return nil
}

// TransferToken 构造并发送 ERC20 转账交易
func TransferToken(client *ethclient.Client, ks *keystore.Key, password, tokenAddr, to string, gasPayerPriv *ecdsa.PrivateKey, logger *log.Logger) (*types.Transaction, error) {
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
	// 计算带缓冲的gas费用 - 使用更高的缓冲比例（300%）以应对gas price波动
	gasWithBuffer := new(big.Int).Div(new(big.Int).Mul(big.NewInt(int64(estimatedGas)), big.NewInt(300)), big.NewInt(100))

	// 确保账户有足够的gas费用，可能需要多次补充（因为gas price可能变化）
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		// 每次重新获取最新的gas price
		currentGasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to suggest gas price")
		}

		// 使用更高的gas price估算（取当前建议价格的120%），以应对价格上涨
		adjustedGasPrice := new(big.Int).Div(new(big.Int).Mul(currentGasPrice, big.NewInt(120)), big.NewInt(100))
		needWei := new(big.Int).Mul(gasWithBuffer, adjustedGasPrice)

		if retry == 0 {
			logger.Printf("BEP20 transfer estimated gas cost: %s BNB (带300%% gas limit缓冲 + 120%% gas price缓冲)\n", WeiToBNB(needWei))
		} else {
			logger.Printf("第 %d 次重试：重新计算gas费用: %s BNB (gas price可能已变化)\n", retry+1, WeiToBNB(needWei))
		}

		// 检查fromAddr的BNB余额是否足够
		bnbBalance, err := client.BalanceAt(context.Background(), fromAddr, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get BNB balance")
		}

		if bnbBalance.Cmp(needWei) < 0 {
			// 计算差额：需要多少 - 现有多少 = 差额
			deficit := new(big.Int).Sub(needWei, bnbBalance)
			logger.Printf("账户 %s BNB余额不足，现有: %s BNB，需要: %s BNB，差额: %s BNB",
				fromAddr.Hex(), WeiToBNB(bnbBalance), WeiToBNB(needWei), WeiToBNB(deficit))

			tx, err := TransferBNB(client, gasPayerPriv, fromAddr, deficit, logger)
			if err != nil {
				return nil, errors.Wrap(err, "failed to transfer gas fee")
			}

			// 等待BNB转账确认，最多等待60秒
			logger.Printf("等待Gas费用差额转账确认: %s", tx.Hash().Hex())
			err = WaitForTransactionConfirmationWithBlocks(client, tx.Hash(), REQUIREDCONFIRMATIONS, 60*time.Second, logger)
			if err != nil {
				return nil, errors.Wrap(err, "Gas费用差额转账确认超时或失败")
			}

			// 补充完成后，继续循环检查（因为gas price可能又变化了）
			continue
		} else {
			logger.Printf("账户 %s BNB余额充足: %s BNB，满足要求: %s BNB",
				fromAddr.Hex(), WeiToBNB(bnbBalance), WeiToBNB(needWei))
			// 余额足够，跳出循环
			break
		}
	}

	// 发送交易前最后一次检查余额和gas price
	finalGasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to get final gas price")
	}
	finalAdjustedGasPrice := new(big.Int).Div(new(big.Int).Mul(finalGasPrice, big.NewInt(120)), big.NewInt(100))
	finalNeedWei := new(big.Int).Mul(gasWithBuffer, finalAdjustedGasPrice)

	finalBalance, err := client.BalanceAt(context.Background(), fromAddr, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get final balance")
	}

	if finalBalance.Cmp(finalNeedWei) < 0 {
		// 最后一次检查发现余额不足，快速补充
		deficit := new(big.Int).Sub(finalNeedWei, finalBalance)
		logger.Printf("发送前最后检查：余额不足，快速补充差额: %s BNB", WeiToBNB(deficit))

		tx, err := TransferBNB(client, gasPayerPriv, fromAddr, deficit, logger)
		if err != nil {
			return nil, errors.Wrap(err, "failed to transfer final gas fee")
		}

		// 等待确认（可以减少确认数，因为只是补充gas费用）
		logger.Printf("等待最终Gas费用补充确认: %s", tx.Hash().Hex())
		err = WaitForTransactionConfirmationWithBlocks(client, tx.Hash(), 1, 30*time.Second, logger)
		if err != nil {
			return nil, errors.Wrap(err, "最终Gas费用补充确认超时或失败")
		}

		// 再次获取最新的gas price
		finalGasPrice, err = client.SuggestGasPrice(context.Background())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get gas price after final top-up")
		}
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privKey, chainID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transactor")
	}
	auth.GasPrice = finalGasPrice
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
