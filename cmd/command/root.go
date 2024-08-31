package command

import (
	"os"

	"github.com/btcsuite/btcd/rpcclient"
)

func init() {
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
}

func RunMain(args []string) error {
	cmdName := ""
	if len(args) > 1 {
		cmdName = args[1]
	}

	ccmd := NewWalletCommand(cmdName)
	return ccmd.Execute()
}
