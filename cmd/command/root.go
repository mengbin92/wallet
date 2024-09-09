package command

import (
	"os"

	"github.com/btcsuite/btcd/rpcclient"
)

func init() {
	var err error 

	if os.Getenv("RPC_CERT") != "" {
		rpc_cert = os.Getenv("RPC_CERT")
	}
	if os.Getenv("RPC_ENDPOINT") != "" {
		rpc_endpoint = os.Getenv("RPC_ENDPOINT")
	}
	if os.Getenv("RPC_USER") != "" {
		rpc_user = os.Getenv("RPC_USER")
	}
	if os.Getenv("RPC_PASSWORD") != "" {
		rpc_password = os.Getenv("RPC_PASSWORD")
	}

	cert, err := os.ReadFile(rpc_cert)
	if err != nil {
		panic(err)
	}
	connCfg := &rpcclient.ConnConfig{
		Host:         rpc_endpoint,
		User:         rpc_user,
		Pass:         rpc_password,
		HTTPPostMode: true,
		Certificates: cert,
	}

	client, err = rpcclient.New(connCfg, nil)
	if err != nil {
		panic(err)
	}

	// TODO: add ping check
	// err = client.Ping()
	// if err != nil {
	// 	panic(err)
	// }
}

func RunMain(args []string) error {
	cmdName := ""
	if len(args) > 1 {
		cmdName = args[1]
	}

	ccmd := NewWalletCommand(cmdName)
	return ccmd.Execute()
}
