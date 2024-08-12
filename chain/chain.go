package chain

import (
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	pb "github.com/mengbin92/wallet/config"
	"github.com/pkg/errors"
)

var (
	networks = map[string]chaincfg.Params{
		"mainnet": chaincfg.MainNetParams,
		"testnet": chaincfg.TestNet3Params,
	}
)

type ChainPower struct {
	client  *rpcclient.Client
	network string
}

// NewChainPower create a new ChainPower instance
func NewChinPower(chainConf *pb.Chain) (*ChainPower, error) {
	certBytes, err := os.ReadFile(chainConf.RpcCert)
	if err != nil {
		return nil, errors.Wrap(err, "read rpc cert file error")
	}

	connCfg := &rpcclient.ConnConfig{
		Host:         chainConf.RpcEndpoint,
		User:         chainConf.RpcUser,
		Pass:         chainConf.RpcPassword,
		HTTPPostMode: true,
		Certificates: certBytes,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new rpc client error")
	}

	return &ChainPower{
		client:  client,
		network: chainConf.Network,
	}, nil
}

// shutdown rpc client
func (c *ChainPower) Shutdown() {
	c.client.Shutdown()
}

func (c *ChainPower) GetBlockCount() (int64, error) {
	return c.client.GetBlockCount()
}

func (c *ChainPower) GetBlockHash(height int64) (string, error) {
	blkHash, err := c.client.GetBlockHash(height)
	if err != nil {
		return "", errors.Wrap(err, "get block hash error")
	}
	return blkHash.String(), nil
}

func (c *ChainPower) GetBlock(blkHashStr string) (*btcutil.Block, error) {
	blkHash, err := chainhash.NewHashFromStr(blkHashStr)
	if err != nil {
		return nil, errors.Wrap(err, "new hash from str error")
	}
	msgBlk, err := c.client.GetBlock(blkHash)
	if err != nil {
		return nil, errors.Wrap(err, "get block error")
	}

	return btcutil.NewBlock(msgBlk), nil
}

func (c *ChainPower) Source() string{
	return "btcd"
}