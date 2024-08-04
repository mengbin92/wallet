package address

import (
	"testing"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/stretchr/testify/assert"
)

var (
	pwd = "pwdtest0pwdtest0"
)

func TestNewBTCAddress(t *testing.T) {
	tests := []struct {
		name  string
		param *chaincfg.Params
		main  bool
	}{
		{
			name:  "MainNetParams",
			param: &chaincfg.MainNetParams,
			main:  true,
		},
		{
			name:  "TestNet3Params",
			param: &chaincfg.TestNet3Params,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBTCAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, addr)
		})
	}
}

func TestGenP2PKAddress(t *testing.T) {
	// TODO
	tests := []struct {
		name  string
		param *chaincfg.Params
	}{
		{
			name:  "MainNetParams-GenP2PKAddress",
			param: &chaincfg.MainNetParams,
		},
		{
			name:  "TestNet3Params-GenP2PKAddress",
			param: &chaincfg.TestNet3Params,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBTCAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, addr)

			p2pkAddr, err := addr.GenP2PKAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, p2pkAddr)
		})
	}
}

func TestGenP2PKHAddress(t *testing.T) {
	// TODO
	tests := []struct {
		name  string
		param *chaincfg.Params
		main  bool
	}{
		{
			name:  "MainNetParams-GenP2PKHAddress",
			param: &chaincfg.MainNetParams,
			main:  true,
		},
		{
			name:  "TestNet3Params-GenP2PKHAddress",
			param: &chaincfg.TestNet3Params,
			main:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBTCAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, addr)
			if tt.main {
				p2pkhAddr, err := addr.GenP2PKHAddress(tt.param)
				assert.Empty(t, err)
				assert.NotEmpty(t, p2pkhAddr)
				assert.Equal(t, p2pkhAddr[0], byte('1'))
			} else {
				p2pkhAddr, err := addr.GenP2PKHAddress(tt.param)
				assert.Empty(t, err)
				assert.NotEmpty(t, p2pkhAddr)
				assert.Equal(t, p2pkhAddr[0], byte('m'))
			}

		})
	}
}

func TestGenBech32Address(t *testing.T) {
	tests := []struct {
		name  string
		param *chaincfg.Params
		main  bool
	}{
		{
			name:  "MainNetParams-GenBech32Address",
			param: &chaincfg.MainNetParams,
			main:  true,
		},
		{
			name:  "TestNet3Params-GenP2PKHAddress",
			param: &chaincfg.TestNet3Params,
			main:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBTCAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, addr)
			if tt.main {
				p2pkhAddr, err := addr.GenBech32Address(tt.param)
				assert.Empty(t, err)
				assert.NotEmpty(t, p2pkhAddr)
				assert.Equal(t, p2pkhAddr[0:3], "bc1")
			} else {
				p2pkhAddr, err := addr.GenBech32Address(tt.param)
				assert.Empty(t, err)
				assert.NotEmpty(t, p2pkhAddr)
				assert.Equal(t, p2pkhAddr[0:3], "tb1")
			}

		})
	}
}

func TestExportPrivateKey(t *testing.T) {
	tests := []struct {
		name  string
		param *chaincfg.Params
	}{
		{
			name:  "MainNetParams-GenBech32Address",
			param: &chaincfg.MainNetParams,
		},
		{
			name:  "TestNet3Params-GenP2PKHAddress",
			param: &chaincfg.TestNet3Params,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := NewBTCAddress(tt.param)
			assert.Empty(t, err)
			assert.NotEmpty(t, addr)
			encryptStr, err := addr.ExportPrivateKey(pwd)
			assert.Empty(t, err)
			assert.NotEmpty(t, encryptStr)

			newAddress := &BTCAddress{}
			err = newAddress.LoadPrivateKey(encryptStr, pwd)
			assert.Empty(t, err)
			
			p2pkhAddr, _ := addr.GenBech32Address(tt.param)
			newp2pkhAddr,_ := newAddress.GenBech32Address(tt.param)
			assert.Equal(t, p2pkhAddr, newp2pkhAddr)
		})
	}
}
