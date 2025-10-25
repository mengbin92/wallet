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

	logger.Printf("BNB 转账已发送: %s -> %s, amount: %s, txHash: %s",
		fromAddr.Hex(), to.Hex(), amount, signedTx.Hash().Hex(),
	)
	return signedTx, nil
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

		// 等待BNB转账确认，最多等待60秒
		logger.Printf("等待Gas费用差额转账确认: %s", tx.Hash().Hex())
		err = WaitForTransactionConfirmationWithBlocks(client, tx.Hash(), REQUIREDCONFIRMATIONS, 60*time.Second, logger)
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