package config

import "time"

// Config holds all configuration for the wallet application
type Config struct {
	// Gas configuration
	Gas GasConfig
	// Confirmation configuration
	Confirmation ConfirmationConfig
	// Network configuration
	Network NetworkConfig
	// Concurrency configuration
	Concurrency ConcurrencyConfig
}

// GasConfig contains gas-related settings
type GasConfig struct {
	// Limit for native coin transfer (BNB/ETH)
	NativeTransferLimit uint64
	// Buffer percentage for ERC20 transfers (230 = 230%)
	ERC20BufferPercentage int64
	// Buffer percentage for native transfers (120 = 120%)
	NativeBufferPercentage int64
	// Price cache TTL
	PriceCacheTTL time.Duration
}

// ConfirmationConfig contains transaction confirmation settings
type ConfirmationConfig struct {
	// Default number of confirmations required
	Default uint64
	// Confirmations for token transfers
	TokenTransfer uint64
	// Confirmations for native coin transfers
	NativeTransfer uint64
	// Check interval
	CheckInterval time.Duration
	// Maximum wait time for confirmation
	MaxWaitTime time.Duration
	// Progress output interval (every N checks)
	ProgressInterval int
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	// RPC timeout
	RPCTimeout time.Duration
	// Max retries for RPC calls
	MaxRetries int
	// Base delay for retry backoff
	RetryBaseDelay time.Duration
}

// ConcurrencyConfig contains concurrency settings
type ConcurrencyConfig struct {
	// Maximum number of concurrent workers
	MaxWorkers int
	// Enable concurrent processing
	Enabled bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Gas: GasConfig{
			NativeTransferLimit:   21000,
			ERC20BufferPercentage: 230,
			NativeBufferPercentage: 120,
			PriceCacheTTL:         30 * time.Second,
		},
		Confirmation: ConfirmationConfig{
			Default:        6,
			TokenTransfer:  6,
			NativeTransfer: 1,
			CheckInterval:  2 * time.Second,
			MaxWaitTime:    60 * time.Second,
			ProgressInterval: 5,
		},
		Network: NetworkConfig{
			RPCTimeout:    30 * time.Second,
			MaxRetries:    3,
			RetryBaseDelay: 1 * time.Second,
		},
		Concurrency: ConcurrencyConfig{
			MaxWorkers: 5,
			Enabled:    true,
		},
	}
}

// ProdConfig returns production-optimized configuration
func ProdConfig() *Config {
	cfg := DefaultConfig()
	cfg.Concurrency.MaxWorkers = 10
	cfg.Network.MaxRetries = 5
	return cfg
}

// TestConfig returns test configuration
func TestConfig() *Config {
	cfg := DefaultConfig()
	cfg.Concurrency.MaxWorkers = 2
	cfg.Confirmation.MaxWaitTime = 30 * time.Second
	return cfg
}

// DevConfig returns development configuration for geth -dev mode
// In dev mode, blocks are not mined automatically, so we use minimal confirmations
func DevConfig() *Config {
	return &Config{
		Gas: GasConfig{
			NativeTransferLimit:   21000,
			ERC20BufferPercentage: 230,
			NativeBufferPercentage: 120,
			PriceCacheTTL:         30 * time.Second,
		},
		Confirmation: ConfirmationConfig{
			Default:        0, // 0 confirmations in dev mode (transaction in mempool is enough)
			TokenTransfer:  0,
			NativeTransfer: 0,
			CheckInterval:  1 * time.Second,
			MaxWaitTime:    10 * time.Second, // Short timeout for dev mode
			ProgressInterval: 2,
		},
		Network: NetworkConfig{
			RPCTimeout:    10 * time.Second,
			MaxRetries:    1, // Fewer retries in dev mode
			RetryBaseDelay: 500 * time.Millisecond,
		},
		Concurrency: ConcurrencyConfig{
			MaxWorkers: 1, // Single worker in dev mode to avoid race conditions
			Enabled:    true,
		},
	}
}
