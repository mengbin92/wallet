package collector

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/mengbin92/wallet/internal/config"
	"github.com/mengbin92/wallet/internal/retry"
	"github.com/mengbin92/wallet/internal/wallet"
	"github.com/pkg/errors"
)

// CollectionResult holds the result of a single collection attempt
type CollectionResult struct {
	Address string
	Success bool
	Error   error
	TxHash  string
	Amount  *big.Int
}

// CollectTokensConcurrent collects tokens from multiple accounts concurrently
func CollectTokensConcurrent(rpcURL, mainKeystoreDir, keystoreDir, password, tokenAddr, targetAddr string, cfg *config.Config, logger *log.Logger) error {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// Connect to RPC
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return errors.Wrapf(err, "failed to connect rpc (rpcURL: %s)", rpcURL)
	}
	defer client.Close()

	// Load keys
	mainKey, err := wallet.LoadAllKeys(mainKeystoreDir, password, logger)
	if err != nil {
		return errors.Wrapf(err, "failed to load main key (mainKey: %s)", mainKeystoreDir)
	}

	keys, err := wallet.LoadAllKeys(keystoreDir, password, logger)
	if err != nil {
		return errors.Wrapf(err, "failed to load keys (keystoreDir: %s)", keystoreDir)
	}

	logger.Printf("开始从 %d 个账户归集代币 (并发数: %d)", len(keys), cfg.Concurrency.MaxWorkers)

	// Create worker pool
	results := collectWithWorkerPool(client, keys, mainKey[0].PrivateKey, password, tokenAddr, targetAddr, cfg, logger)

	// Print summary
	printCollectionSummary(results, logger)

	return nil
}

// collectWithWorkerPool processes collections using a worker pool pattern
func collectWithWorkerPool(client *ethclient.Client, keys []*keystore.Key, gasPayerKey *ecdsa.PrivateKey, password, tokenAddr, targetAddr string, cfg *config.Config, logger *log.Logger) []*CollectionResult {
	// Prepare tasks
	type collectionTask struct {
		key     *keystore.Key
		balance *big.Int
	}

	// First, get all balances concurrently
	tasks := make([]collectionTask, 0, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrent balance checks
	balanceSem := make(chan struct{}, cfg.Concurrency.MaxWorkers)

	for _, key := range keys {
		wg.Add(1)
		go func(k *keystore.Key) {
			defer wg.Done()
			balanceSem <- struct{}{}        // Acquire
			defer func() { <-balanceSem }() // Release

			addr := k.Address.Hex()
			err := retry.WithExponentialBackoff(func() error {
				bal, err := GetTokenBalance(client, tokenAddr, addr)
				if err != nil {
					return err
				}

				mu.Lock()
				tasks = append(tasks, collectionTask{key: k, balance: bal})
				mu.Unlock()
				return nil
			}, &retry.Config{MaxRetries: cfg.Network.MaxRetries, BaseDelay: cfg.Network.RetryBaseDelay})

			if err != nil {
				logger.Printf("获取 %s 余额失败: %v", addr, err)
			}
		}(key)
	}
	wg.Wait()

	// Filter tasks with balance
	validTasks := make([]collectionTask, 0, len(tasks))
	for _, t := range tasks {
		if t.balance.Cmp(big.NewInt(0)) == 0 {
			logger.Printf("跳过 %s, 余额=0", t.key.Address.Hex())
		} else {
			logger.Printf("发现余额 %s: %s tokens", t.key.Address.Hex(), new(big.Int).Div(t.balance, DECIMALS).String())
			validTasks = append(validTasks, t)
		}
	}

	if len(validTasks) == 0 {
		return []*CollectionResult{}
	}

	logger.Printf("找到 %d 个有余额的账户", len(validTasks))

	// Process transfers with worker pool
	results := make([]*CollectionResult, 0, len(validTasks))
	resultChan := make(chan *CollectionResult, len(validTasks))

	// Create worker pool
	sem := make(chan struct{}, cfg.Concurrency.MaxWorkers)
	var workerWg sync.WaitGroup

	for _, task := range validTasks {
		workerWg.Add(1)
		go func(t collectionTask) {
			defer workerWg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			result := &CollectionResult{
				Address: t.key.Address.Hex(),
				Amount:  t.balance,
			}

			logger.Printf("开始归集 %s, 余额=%s", t.key.Address.Hex(), new(big.Int).Div(t.balance, DECIMALS).String())

			// Execute transfer with retry
			err := retry.WithExponentialBackoff(func() error {
				// Prepare confirmation config
				confirmCfg := &WaitForConfirmationConfig{
					RequiredConfirmations: cfg.Confirmation.TokenTransfer,
					CheckInterval:         cfg.Confirmation.CheckInterval,
					MaxWaitTime:           cfg.Confirmation.MaxWaitTime,
					ProgressInterval:      cfg.Confirmation.ProgressInterval,
				}

				tx, err := TransferToken(client, t.key, password, tokenAddr, targetAddr, gasPayerKey, logger, confirmCfg)
				if err != nil {
					return err
				}
				result.TxHash = tx.Hash().Hex()

				// Wait for confirmation
				return WaitForTransactionConfirmation(client, tx.Hash(), confirmCfg, logger)
			}, &retry.Config{MaxRetries: cfg.Network.MaxRetries, BaseDelay: cfg.Network.RetryBaseDelay})

			if err != nil {
				result.Success = false
				result.Error = err
				logger.Printf("归集 %s 失败: %v", t.key.Address.Hex(), err)
			} else {
				result.Success = true
				logger.Printf("归集 %s 成功: tx=%s", t.key.Address.Hex(), result.TxHash)
			}

			resultChan <- result
		}(task)
	}

	// Close result channel when all workers complete
	go func() {
		workerWg.Wait()
		close(resultChan)
	}()

	// Collect results
	for result := range resultChan {
		results = append(results, result)
	}

	return results
}

// printCollectionSummary prints a summary of the collection process
func printCollectionSummary(results []*CollectionResult, logger *log.Logger) {
	if len(results) == 0 {
		logger.Printf("没有需要归集的账户")
		return
	}

	successCount := 0
	failureCount := 0
	totalAmount := big.NewInt(0)

	for _, r := range results {
		if r.Success {
			successCount++
			totalAmount.Add(totalAmount, r.Amount)
		} else {
			failureCount++
		}
	}

	logger.Printf("\n=== 归集统计 ===")
	logger.Printf("总账户数: %d", len(results))
	logger.Printf("成功: %d", successCount)
	logger.Printf("失败: %d", failureCount)
	logger.Printf("总归集金额: %s tokens", new(big.Int).Div(totalAmount, DECIMALS).String())
	logger.Printf("成功率为: %.2f%%", float64(successCount)/float64(len(results))*100)
}

// GetRPCClient creates a new RPC client
func GetRPCClient(rpcURL string) (*rpc.Client, error) {
	return rpc.Dial(rpcURL)
}
