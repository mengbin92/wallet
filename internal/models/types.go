package models

import "math/big"

type Address struct {
	Chain      string `json:"chain"`       // 链类型 (eth, btc, tron...)
	Address    string `json:"address"`     // 地址字符串
	PublicKey  string `json:"public_key"`  // 公钥（hex/base58）
	PrivateKey string `json:"private_key"` // 私钥（加密存储，导出时可为空）
	Path       string `json:"path"`        // HD 钱包派生路径 (m/44'/60'/0'/0/0)
}

type Transaction struct {
	Chain      string   `json:"chain"`       // 链类型
	From       string   `json:"from"`        // 发起地址
	To         string   `json:"to"`          // 接收地址
	Amount     *big.Int `json:"amount"`      // 转账金额 (统一用最小单位的浮点/字符串)
	Fee        *big.Int `json:"fee"`         // 手续费 (如果链有特殊机制，可以置 0)
	Nonce      uint64   `json:"nonce"`       // ETH/账户模型需要
	UTXOs      []*UTXO  `json:"utxos"`       // BTC/UTXO 模型需要
	RawPayload string   `json:"raw_payload"` // 原始构造的未签名交易 (hex/json)
	TxID       string   `json:"txid"`        // 交易哈希 (签名+广播后填充)
}

type UTXO struct {
	TxID   string  `json:"txid"`
	Vout   uint32  `json:"vout"`
	Amount float64 `json:"amount"`
	Script string  `json:"script"` // redeem script/pubkey script
}

type SignedTransaction struct {
	Chain string `json:"chain"`
	Raw   string `json:"raw"`  // 已签名的原始交易 (hex/base64)
	TxID  string `json:"txid"` // 交易哈希
}
