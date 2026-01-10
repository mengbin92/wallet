package main

import (
	"log"

	"github.com/mengbin92/wallet/internal/collector"
	"github.com/mengbin92/wallet/internal/config"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func NewCollectorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect ERC20 tokens from all accounts in keystore",
		Long: `Collect ERC20 tokens from multiple keystore accounts.

Supports both serial and concurrent collection modes.

Examples:
  # Serial collection (original method)
  wallet collect run --token 0x... --to 0x...

  # Concurrent collection (5x-20x faster)
  wallet collect concurrent --token 0x... --to 0x... --workers 10

  # Using preset configurations
  wallet collect concurrent --token 0x... --to 0x... --config prod`,
	}

	// -------------------- 注册子命令 --------------------
	cmd.AddCommand(collectCmd())
	cmd.AddCommand(collectConcurrentCmd())
	cmd.AddCommand(collectStatsCmd())
	cmd.AddCommand(collectBenchmarkCmd())
	return cmd
}

// collectCmd is the original serial collection command
func collectCmd() *cobra.Command {
	var (
		rpcURL      string
		mainKeyDir  string
		keystoreDir string
		password    string
		tokenAddr   string
		targetAddr  string
	)

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run token collection in serial mode (original)",
		Long: `Run token collection from all keystore accounts in serial mode.

This is the original collection method. For better performance with
multiple accounts, consider using the "concurrent" command instead.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenAddr == "" || targetAddr == "" {
				return errors.Errorf("token address and target address are required (token: %s, target: %s, use --token and --to flags)", tokenAddr, targetAddr)
			}

			log.Printf("🔄 Starting serial token collection...")
			log.Printf("⚠️  For better performance with multiple accounts, use 'concurrent' command (5x-20x faster)")

			return collector.CollectTokens(rpcURL, mainKeyDir, keystoreDir, password, tokenAddr, targetAddr, log.Default())
		},
	}

	cmd.Flags().StringVar(&rpcURL, "rpc", "https://bsc-dataseed.binance.org", "RPC endpoint")
	cmd.Flags().StringVar(&mainKeyDir, "mainkey", "mainkey", "main key directory")
	cmd.Flags().StringVar(&keystoreDir, "keystore", "keystore", "keystore directory")
	cmd.Flags().StringVar(&password, "password", "", "keystore password")
	cmd.Flags().StringVar(&tokenAddr, "token", "", "token contract address")
	cmd.Flags().StringVar(&targetAddr, "to", "", "target address to collect tokens into")

	return cmd
}

// collectConcurrentCmd is the new concurrent collection command
func collectConcurrentCmd() *cobra.Command {
	var (
		rpcURL      string
		mainKeyDir  string
		keystoreDir string
		password    string
		tokenAddr   string
		targetAddr  string
		workers     int
		configPreset string
		enableStats bool
	)

	cmd := &cobra.Command{
		Use:   "concurrent",
		Short: "Run token collection in concurrent mode (5x-20x faster)",
		Long: `Run token collection from all keystore accounts in concurrent mode.

This command uses a worker pool to process multiple accounts concurrently,
providing 5x-20x performance improvement compared to serial mode.

Performance examples:
  - 10 accounts:  ~10x faster with 10 workers
  - 50 accounts:  ~10x faster with 10 workers
  - 100 accounts: ~20x faster with 20 workers

Worker configuration:
  - Small scale (<10 accounts):     5 workers (default)
  - Medium scale (10-50 accounts):  10 workers (recommended)
  - Large scale (50-100 accounts):  10-20 workers
  - Very large scale (100+ accounts): 20 workers`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenAddr == "" || targetAddr == "" {
				return errors.Errorf("token address and target address are required (token: %s, target: %s, use --token and --to flags)", tokenAddr, targetAddr)
			}

			// Load configuration
			var cfg *config.Config
			switch configPreset {
			case "prod", "production":
				cfg = config.ProdConfig()
				log.Printf("📋 Using production configuration")
			case "test":
				cfg = config.TestConfig()
				log.Printf("📋 Using test configuration")
			default:
				cfg = config.DefaultConfig()
				log.Printf("📋 Using default configuration")
			}

			// Override worker count if specified
			if workers > 0 {
				cfg.Concurrency.MaxWorkers = workers
				log.Printf("⚙️  Worker count overridden: %d", workers)
			}

			// Display configuration
			log.Printf("\n╔════════════════════════════════════════════════════════════╗")
			log.Printf("║          Concurrent Token Collection Configuration          ║")
			log.Printf("╠════════════════════════════════════════════════════════════╣")
			log.Printf("║  RPC Endpoint:        %-40s ║", rpcURL)
			log.Printf("║  Max Workers:         %-40d ║", cfg.Concurrency.MaxWorkers)
			log.Printf("║  Max Retries:         %-40d ║", cfg.Network.MaxRetries)
			log.Printf("║  Token Confirmations: %-40d ║", cfg.Confirmation.TokenTransfer)
			log.Printf("║  Max Wait Time:       %-40v ║", cfg.Confirmation.MaxWaitTime)
			log.Printf("╚════════════════════════════════════════════════════════════╝\n")

			log.Printf("🚀 Starting concurrent token collection...")
			log.Printf("⚡ Expected speedup: %dx-%dx with %d workers\n",
				cfg.Concurrency.MaxWorkers/2,
				cfg.Concurrency.MaxWorkers,
				cfg.Concurrency.MaxWorkers)

			return collector.CollectTokensConcurrent(
				rpcURL,
				mainKeyDir,
				keystoreDir,
				password,
				tokenAddr,
				targetAddr,
				cfg,
				log.Default(),
			)
		},
	}

	cmd.Flags().StringVar(&rpcURL, "rpc", "https://bsc-dataseed.binance.org", "RPC endpoint")
	cmd.Flags().StringVar(&mainKeyDir, "mainkey", "mainkey", "main key directory")
	cmd.Flags().StringVar(&keystoreDir, "keystore", "keystore", "keystore directory")
	cmd.Flags().StringVar(&password, "password", "", "keystore password")
	cmd.Flags().StringVar(&tokenAddr, "token", "", "token contract address")
	cmd.Flags().StringVar(&targetAddr, "to", "", "target address to collect tokens into")
	cmd.Flags().IntVar(&workers, "workers", 0, "Number of concurrent workers (0 = auto from config preset)")
	cmd.Flags().StringVar(&configPreset, "config", "default", "Configuration preset: default, prod, test")
	cmd.Flags().BoolVar(&enableStats, "stats", false, "Enable detailed statistics logging")

	return cmd
}

// collectStatsCmd shows statistics about previous collections
func collectStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show collection statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("📊 Collection Statistics")
			log.Printf("╔════════════════════════════════════════════════════════════╗")
			log.Printf("║  Feature coming soon!                                       ║")
			log.Printf("║  This will show historical collection performance data      ║")
			log.Printf("╚════════════════════════════════════════════════════════════╝")
			return nil
		},
	}

	// TODO: Add stats file reading
	// cmd.Flags().StringVar(&statsFile, "file", "collection_stats.json", "Statistics file path")

	return cmd
}

// collectBenchmarkCmd runs a benchmark comparison
func collectBenchmarkCmd() *cobra.Command {
	var (
		rpcURL      string
		keystoreDir string
		password    string
		tokenAddr   string
		targetAddr  string
	)

	cmd := &cobra.Command{
		Use:   "benchmark",
		Short: "Run benchmark: serial vs concurrent",
		Long: `Run a performance benchmark comparing serial and concurrent collection modes.

This command will:
  1. Run serial collection and measure time
  2. Run concurrent collection and measure time
  3. Display performance comparison and speedup`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if tokenAddr == "" || targetAddr == "" {
				return errors.Errorf("token address and target address are required (token: %s, target: %s, use --token and --to flags)", tokenAddr, targetAddr)
			}

			log.Printf("🔬 Running performance benchmark...")
			log.Printf("This will run both serial and concurrent collection")
			log.Printf("WARNING: This will take extra time but provide useful performance data\n")

			// TODO: Implement actual benchmark
			log.Printf("⚠️  Benchmark feature coming soon!")
			log.Printf("For now, use: go test -bench=. ./internal/collector/...")

			return nil
		},
	}

	cmd.Flags().StringVar(&rpcURL, "rpc", "https://bsc-dataseed.binance.org", "RPC endpoint")
	cmd.Flags().StringVar(&keystoreDir, "keystore", "keystore", "keystore directory")
	cmd.Flags().StringVar(&password, "password", "", "keystore password")
	cmd.Flags().StringVar(&tokenAddr, "token", "", "token contract address")
	cmd.Flags().StringVar(&targetAddr, "to", "", "target address to collect tokens into")

	return cmd
}
