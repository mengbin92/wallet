package blockchain

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client defines the interface for blockchain operations
// This abstraction allows for easier testing and mocking
type Client interface {
	// Balance operations
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error)

	// Transaction operations
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)

	// Transaction sending
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// Transaction receipt
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	// Block operations
	BlockNumber(ctx context.Context) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)

	// Network
	NetworkID(ctx context.Context) (*big.Int, error)

	// Close
	Close()
}

// ETHClient is the standard Ethereum client implementation
type ETHClient struct {
	client *ethclient.Client
}

// NewETHClient creates a new Ethereum client
func NewETHClient(rawURL string) (Client, error) {
	client, err := ethclient.Dial(rawURL)
	if err != nil {
		return nil, err
	}

	return &ETHClient{client: client}, nil
}

// NewETHClientWithRPC creates a new Ethereum client with a custom RPC client
func NewETHClientWithRPC(rpcClient *rpc.Client) Client {
	return &ETHClient{client: ethclient.NewClient(rpcClient)}
}

// BalanceAt implements Client interface
func (e *ETHClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	return e.client.BalanceAt(ctx, account, blockNumber)
}

// PendingBalanceAt implements Client interface
func (e *ETHClient) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	return e.client.PendingBalanceAt(ctx, account)
}

// PendingNonceAt implements Client interface
func (e *ETHClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	return e.client.PendingNonceAt(ctx, account)
}

// SuggestGasPrice implements Client interface
func (e *ETHClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	return e.client.SuggestGasPrice(ctx)
}

// EstimateGas implements Client interface
func (e *ETHClient) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
	return e.client.EstimateGas(ctx, msg)
}

// SendTransaction implements Client interface
func (e *ETHClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	return e.client.SendTransaction(ctx, tx)
}

// TransactionReceipt implements Client interface
func (e *ETHClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return e.client.TransactionReceipt(ctx, txHash)
}

// BlockNumber implements Client interface
func (e *ETHClient) BlockNumber(ctx context.Context) (uint64, error) {
	return e.client.BlockNumber(ctx)
}

// HeaderByNumber implements Client interface
func (e *ETHClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return e.client.HeaderByNumber(ctx, number)
}

// NetworkID implements Client interface
func (e *ETHClient) NetworkID(ctx context.Context) (*big.Int, error) {
	return e.client.NetworkID(ctx)
}

// Close implements Client interface
func (e *ETHClient) Close() {
	e.client.Close()
}

// GetWrappedClient returns the underlying ethclient.Client
// Use sparingly, only when you need access to ethclient-specific methods
func (e *ETHClient) GetWrappedClient() *ethclient.Client {
	return e.client
}
