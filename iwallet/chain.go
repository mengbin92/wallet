package iwallet

import "github.com/btcsuite/btcd/btcutil"

type IChainPower interface {
	// get block from blockchain
	GetBlock(blkHash string) (*btcutil.Block, error)
	// get block count from blockchain
	GetBlockCount() (int64, error)
	// get block hash from blockchain
	GetBlockHash(int64) (string, error)
	// get ChainPower name
	Source() string
}
