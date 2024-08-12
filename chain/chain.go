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
	networks = map[string]*chaincfg.Params{
		"mainnet": &chaincfg.MainNetParams,
		"testnet": &chaincfg.TestNet3Params,
	}
	decimals = 10e8
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

func (c *ChainPower) Source() string {
	return "btcd"
}

func (c *ChainPower) GetBalanceByAddress(addr string) (int64, error) {
	// 解析比特币地址
	address, err := btcutil.DecodeAddress(addr, networks[c.network])
	if err != nil {
		return 0, errors.Wrap(err, "解析比特币地址失败")
	}

	// 使用SearchRawTransactionsVerbose获取与地址相关的所有交易
	transactions, err := c.client.SearchRawTransactionsVerbose(address, 0, 100, true, false, nil)
	if err != nil {
		return 0, errors.Wrap(err, "获取与地址相关的所有交易失败")
	}

	amount := int64(0)
	// 遍历所有交易
	for _, tx := range transactions {
		// 将交易ID字符串转换为链哈希对象
		txid, err := chainhash.NewHashFromStr(tx.Txid)
		if err != nil {
			return 0, errors.Wrap(err, "将交易ID字符串转换为链哈希对象失败")
		}

		// 遍历交易的输出
		for _, vout := range tx.Vout {
			// 检查输出地址是否是我们关心的地址
			if vout.ScriptPubKey.Address != addr {
				continue
			}

			// 使用GetTxOut方法获取交易输出，确认该输出是否未花费
			utxo, err := c.client.GetTxOut(txid, vout.N, true)
			if err != nil {
				return 0, errors.Wrap(err, "获取交易输出失败")
			}

			// 如果交易输出未花费，则累加金额
			if utxo != nil {
				amount += int64(utxo.Value * decimals)
			}
		}
	}
	return amount, nil
}
