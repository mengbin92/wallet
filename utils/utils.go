package utils

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/pkg/errors"
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

func CreatePassphrase(n int)(string,error){
	buf := make([]byte, n)
	if _,err := io.ReadFull(rand.Reader, buf); err!= nil {
		return "",errors.Wrap(err,"failed to read random bytes")
	}
	return hex.EncodeToString(buf),nil
}