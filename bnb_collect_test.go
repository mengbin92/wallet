package main

import (
	"context"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mengbin92/wallet/internal/collector"
	"github.com/mengbin92/wallet/internal/wallet"
)

// CollectBNB 归集所有keystore中的BNB到指定地址
func CollectBNB(rpcURL, keystoreDir, password, targetAddr string, logger *log.Logger) error {
	// 连接RPC
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return err
	}
	defer client.Close()

	// 加载所有keystore
	keys, err := wallet.LoadAllKeys(keystoreDir, password, logger)
	if err != nil {
		return err
	}

	logger.Printf("📁 找到 %d 个keystore文件", len(keys))

	targetAddress := common.HexToAddress(targetAddr)
	totalTransferred := big.NewInt(0)
	successCount := 0
	failureCount := 0

	// 遍历所有账户
	for i, key := range keys {
		fromAddr := key.Address.Hex()
		logger.Printf("\n--- 处理账户 %d/%d: %s ---", i+1, len(keys), fromAddr)

		// 获取BNB余额
		balance, err := client.BalanceAt(context.Background(), key.Address, nil)
		if err != nil {
			logger.Printf("❌ 获取余额失败: %v", err)
			failureCount++
			continue
		}

		logger.Printf("💰 当前余额: %s wei", balance.String())

		// 如果余额为0，跳过
		if balance.Cmp(big.NewInt(0)) == 0 {
			logger.Printf("💸 余额为0，跳过")
			continue
		}

		// 计算转账金额（保留少量BNB作为gas费用）
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			logger.Printf("❌ 获取gas价格失败: %v", err)
			failureCount++
			continue
		}

		// 估算gas费用（BNB转账需要21000 gas）
		gasCost := new(big.Int).Mul(big.NewInt(21000), gasPrice)
		// 添加20%缓冲
		gasBuffer := new(big.Int).Div(new(big.Int).Mul(gasCost, big.NewInt(120)), big.NewInt(100))

		// 计算实际转账金额
		transferAmount := new(big.Int).Sub(balance, gasBuffer)

		// 如果转账金额小于等于0，跳过
		if transferAmount.Cmp(big.NewInt(0)) <= 0 {
			logger.Printf("⚠️  余额不足以支付gas费用，跳过")
			continue
		}

		logger.Printf("⛽ 预估gas费用: %s wei", gasBuffer.String())
		logger.Printf("💸 转账金额: %s wei", transferAmount.String())

		// 执行BNB转账
		tx, err := collector.TransferBNB(client, key.PrivateKey, targetAddress, transferAmount, logger)
		if err != nil {
			logger.Printf("❌ BNB转账失败: %v", err)
			failureCount++
			continue
		}

		// 等待交易确认
		logger.Printf("⏳ 等待交易确认: %s", tx.Hash().Hex())
		err = collector.WaitForTransactionConfirmationSimple(client, tx.Hash(), 60*time.Second, logger)
		if err != nil {
			logger.Printf("❌ 交易确认失败: %v", err)
			failureCount++
			continue
		}

		successCount++
		totalTransferred.Add(totalTransferred, transferAmount)
		logger.Printf("✅ BNB转账成功: %s", transferAmount.String())
	}

	// 输出统计结果
	logger.Printf("\n📊 归集统计:")
	logger.Printf("总账户数: %d", len(keys))
	logger.Printf("成功转账: %d", successCount)
	logger.Printf("失败转账: %d", failureCount)
	logger.Printf("总转账金额: %s wei", totalTransferred.String())

	// 转换为ETH显示
	tenTo18 := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	ethAmount := new(big.Int).Div(totalTransferred, tenTo18)
	logger.Printf("总转账金额: %s ETH", ethAmount.String())

	return nil
}

// TestCollectBNB 测试BNB归集功能
func TestCollectBNB(t *testing.T) {
	logger := log.New(log.Writer(), "[BNBTest] ", log.LstdFlags)

	// 测试配置
	config := struct {
		rpcURL      string
		keystoreDir string
		password    string
		targetAddr  string
	}{
		rpcURL:      "https://public-bsc-mainnet.fastnode.io",
		keystoreDir: "/Users/mengbin/Desktop/test/main/keystore",
		password:    "",
		targetAddr:  "0x6ee6Bb78166451E369CD6914190453C45A76ca6a", // 目标地址
	}

	logger.Printf("🚀 开始BNB归集测试")
	logger.Printf("RPC: %s", config.rpcURL)
	logger.Printf("Keystore目录: %s", config.keystoreDir)
	logger.Printf("目标地址: %s", config.targetAddr)

	startTime := time.Now()
	err := CollectBNB(config.rpcURL, config.keystoreDir, config.password, config.targetAddr, logger)
	duration := time.Since(startTime)

	if err != nil {
		logger.Printf("❌ BNB归集失败: %v", err)
	} else {
		logger.Printf("✅ BNB归集完成 (耗时: %v)", duration)
	}
}
