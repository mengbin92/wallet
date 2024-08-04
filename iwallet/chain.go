package iwallet

import "github.com/btcsuite/btcd/btcutil"

type IChainPower interface {
	// get block from blockchain
	GetBlock(network, blkHash string) (*btcutil.Block, error)
	// get block count from blockchain
	GetBlockCount(string) (uint64, error)
	// get block hash from blockchain
	GetBlockHash(string, uint64) (string, error)
	// get ChainPower name
	Source() string
}
