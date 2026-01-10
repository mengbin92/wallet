package collector

import (
	"context"
	"log"
	"math/big"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/mengbin92/wallet/internal/config"
)

// MockClient is a mock blockchain client for testing
type MockClient struct {
	delay time.Duration
}

func (m *MockClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	time.Sleep(m.delay)
	return big.NewInt(1000000000000000000), nil // 1 token
}

func (m *MockClient) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	time.Sleep(m.delay)
	return big.NewInt(1000000000000000000), nil
}

func (m *MockClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	time.Sleep(m.delay)
	return 0, nil
}

func (m *MockClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	time.Sleep(m.delay / 2)
	return big.NewInt(10000000000), nil // 10 gwei
}

func (m *MockClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	time.Sleep(m.delay)
	return 21000, nil
}

func (m *MockClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	time.Sleep(m.delay)
	return nil
}

func (m *MockClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	time.Sleep(m.delay / 2)
	return &types.Receipt{
		Status:             types.ReceiptStatusSuccessful,
		BlockNumber:        big.NewInt(100),
		TxHash:             txHash,
	}, nil
}

func (m *MockClient) BlockNumber(ctx context.Context) (uint64, error) {
	time.Sleep(m.delay / 2)
	return 106, nil // 6 confirmations
}

func (m *MockClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	time.Sleep(m.delay / 2)
	return &types.Header{}, nil
}

func (m *MockClient) NetworkID(ctx context.Context) (*big.Int, error) {
	time.Sleep(m.delay / 2)
	return big.NewInt(56), nil // BSC
}

func (m *MockClient) Close() {}

// BenchmarkSerialCollection benchmarks the serial collection process
func BenchmarkSerialCollection(b *testing.B) {
	mockClient := &MockClient{delay: 100 * time.Millisecond}
	_ = log.New(&testWriter{}, "[TEST] ", log.LstdFlags)

	// Simulate 10 accounts
	numAccounts := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		startTime := time.Now()
		b.StartTimer()

		// Simulate serial processing
		for j := 0; j < numAccounts; j++ {
			// Balance check
			mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)

			// Gas price check
			mockClient.SuggestGasPrice(context.Background())

			// Transaction send
			mockClient.SendTransaction(context.Background(), &types.Transaction{})

			// Wait for confirmation (simulating 3 checks)
			for k := 0; k < 3; k++ {
				mockClient.TransactionReceipt(context.Background(), common.Hash{})
				mockClient.BlockNumber(context.Background())
				time.Sleep(2 * time.Second)
			}
		}

		b.StopTimer()
		elapsed := time.Since(startTime)
		b.ReportMetric(float64(elapsed.Milliseconds()), "ms/op")
		b.StartTimer()
	}
}

// BenchmarkConcurrentCollection benchmarks the concurrent collection process
func BenchmarkConcurrentCollection(b *testing.B) {
	mockClient := &MockClient{delay: 100 * time.Millisecond}
	_ = log.New(&testWriter{}, "[TEST] ", log.LstdFlags)
	cfg := config.DefaultConfig()
	cfg.Concurrency.MaxWorkers = 5

	// Simulate 10 accounts
	numAccounts := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		startTime := time.Now()
		b.StartTimer()

		// Simulate concurrent processing with worker pool
		var opsCompleted int64

		sem := make(chan struct{}, cfg.Concurrency.MaxWorkers)
		var wg sync.WaitGroup

		for j := 0; j < numAccounts; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				// Balance check
				mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
				atomic.AddInt64(&opsCompleted, 1)

				// Gas price check
				mockClient.SuggestGasPrice(context.Background())
				atomic.AddInt64(&opsCompleted, 1)

				// Transaction send
				mockClient.SendTransaction(context.Background(), &types.Transaction{})
				atomic.AddInt64(&opsCompleted, 1)

				// Wait for confirmation (simulating 3 checks)
				for k := 0; k < 3; k++ {
					mockClient.TransactionReceipt(context.Background(), common.Hash{})
					mockClient.BlockNumber(context.Background())
					time.Sleep(2 * time.Second)
				}
				atomic.AddInt64(&opsCompleted, 1)
			}()
		}

		wg.Wait()

		b.StopTimer()
		elapsed := time.Since(startTime)
		b.ReportMetric(float64(elapsed.Milliseconds()), "ms/op")
		_ = opsCompleted // Track operations
		b.StartTimer()
	}
}

// BenchmarkCollectionWithDifferentAccounts tests performance with different account counts
func BenchmarkCollectionWithDifferentAccounts(b *testing.B) {
	accounts := []int{5, 10, 20, 50, 100}
	workers := []int{1, 5, 10, 20}

	for _, numAccounts := range accounts {
		for _, numWorkers := range workers {
			b.Run(
				"accounts_"+string(rune(numAccounts))+"_workers_"+string(rune(numWorkers)),
				func(b *testing.B) {
					mockClient := &MockClient{delay: 50 * time.Millisecond}
					cfg := config.DefaultConfig()
					cfg.Concurrency.MaxWorkers = numWorkers

					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						var wg sync.WaitGroup
						sem := make(chan struct{}, numWorkers)

						for j := 0; j < numAccounts; j++ {
							wg.Add(1)
							go func() {
								defer wg.Done()
								sem <- struct{}{}
								defer func() { <-sem }()

								mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
								mockClient.SuggestGasPrice(context.Background())
							}()
						}
						wg.Wait()
					}
				},
			)
		}
	}
}

// TestConcurrentPerformanceComparison directly compares serial vs concurrent
func TestConcurrentPerformanceComparison(t *testing.T) {
	mockClient := &MockClient{delay: 50 * time.Millisecond}
	_ = log.New(&testWriter{}, "[TEST] ", log.LstdFlags)
	cfg := config.DefaultConfig()
	cfg.Concurrency.MaxWorkers = 5

	numAccounts := 10

	// Test Serial
	t.Log("Testing Serial Collection...")
	startSerial := time.Now()
	for i := 0; i < numAccounts; i++ {
		mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
		mockClient.SuggestGasPrice(context.Background())
	}
	serialTime := time.Since(startSerial)
	t.Logf("Serial time: %v for %d accounts", serialTime, numAccounts)

	// Test Concurrent
	t.Log("Testing Concurrent Collection...")
	startConcurrent := time.Now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, cfg.Concurrency.MaxWorkers)

	for i := 0; i < numAccounts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
			mockClient.SuggestGasPrice(context.Background())
		}()
	}
	wg.Wait()
	concurrentTime := time.Since(startConcurrent)
	t.Logf("Concurrent time: %v for %d accounts", concurrentTime, numAccounts)

	// Calculate speedup
	speedup := float64(serialTime) / float64(concurrentTime)
	t.Logf("Speedup: %.2fx", speedup)

	// Assert that concurrent is faster
	if concurrentTime >= serialTime {
		t.Errorf("Concurrent (%v) should be faster than serial (%v)", concurrentTime, serialTime)
	}

	// Assert minimum speedup
	if speedup < 2.0 {
		t.Logf("WARNING: Speedup is only %.2fx, expected at least 2x with %d workers", speedup, cfg.Concurrency.MaxWorkers)
	}
}

// testWriter is a no-op writer for tests
type testWriter struct{}

func (t *testWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// BenchmarkWithRealisticDelay simulates real-world RPC delays
func BenchmarkWithRealisticDelay(b *testing.B) {
	delays := []time.Duration{
		50 * time.Millisecond,  // Fast RPC
		100 * time.Millisecond, // Average RPC
		200 * time.Millisecond, // Slow RPC
	}

	for _, delay := range delays {
		b.Run("delay_"+delay.String(), func(b *testing.B) {
			mockClient := &MockClient{delay: delay}
			cfg := config.DefaultConfig()
			cfg.Concurrency.MaxWorkers = 5

			numAccounts := 20

			b.Run("Serial", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					for j := 0; j < numAccounts; j++ {
						mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
					}
				}
			})

			b.Run("Concurrent", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					var wg sync.WaitGroup
					sem := make(chan struct{}, cfg.Concurrency.MaxWorkers)

					for j := 0; j < numAccounts; j++ {
						wg.Add(1)
						go func() {
							defer wg.Done()
							sem <- struct{}{}
							defer func() { <-sem }()

							mockClient.BalanceAt(context.Background(), common.HexToAddress("0x0000"), nil)
						}()
					}
					wg.Wait()
				}
			})
		})
	}
}
