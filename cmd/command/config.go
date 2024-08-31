package command

import "github.com/btcsuite/btcd/rpcclient"

const (
	longName = "Command line BTC Wallet"
	cmdName  = "wallet"
	decimals = 1e8
)

// wallet config
var (
	version = "0.0.1"
	network = "mainnet"
)

// btcd rpc config
var (
	rpc_user     = "rpcusertest"
	rpc_password = "sjVj'rLmng;E>5)"
	rpc_endpoint = "127.0.0.1:8334"
	rpc_cert     = "./config/rpc.cert"
	client       *rpcclient.Client
)
