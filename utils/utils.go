package utils

import (
	"github.com/btcsuite/btcd/chaincfg"
)

var (
	networks = map[string]*chaincfg.Params{
		"mainnet": &chaincfg.MainNetParams,
		"testnet": &chaincfg.TestNet3Params,
	}
)

func GetNetwork(network string) *chaincfg.Params {
	return networks[network]
}
