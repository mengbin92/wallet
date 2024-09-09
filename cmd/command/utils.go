package command

import (
	"encoding/hex"

	"github.com/AlecAivazis/survey/v2"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/mengbin92/wallet/storage"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
)

// getUTXOs 获取与地址相关的所有UTXO
// 该函数使用SearchRawTransactionsVerbose方法获取与地址相关的所有交易，然后遍历每个交易的输出，
func getUTXOs(addr, network string) ([]*btcjson.ListUnspentResult, error) {
	// 解析比特币地址
	address, err := btcutil.DecodeAddress(addr, utils.GetNetwork(network))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse bitcoin address")
	}

	// 使用SearchRawTransactionsVerbose获取与地址相关的所有交易
	transactions, err := client.SearchRawTransactionsVerbose(address, 0, 100, true, false, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get transactions with address: %s", addr)
	}

	// 用于存储UTXO的切片
	utxos := []*btcjson.ListUnspentResult{}

	// 遍历所有交易
	for _, tx := range transactions {
		// 将交易ID字符串转换为链哈希对象
		txid, err := chainhash.NewHashFromStr(tx.Txid)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse transaction hash")
		}

		// 遍历交易的输出
		for _, vout := range tx.Vout {
			// 检查输出地址是否是我们关心的地址
			if vout.ScriptPubKey.Address != addr {
				continue
			}

			// 使用GetTxOut方法获取交易输出，确认该输出是否未花费
			utxo, err := client.GetTxOut(txid, vout.N, true)
			if err != nil {
				return nil, errors.Wrap(err, "failed to get transaction output")
			}

			// 如果交易输出未花费，则将其添加到UTXO切片中
			if utxo != nil {
				utxo := &btcjson.ListUnspentResult{
					TxID:          tx.Txid,
					Vout:          uint32(vout.N),
					Address:       addr,
					ScriptPubKey:  vout.ScriptPubKey.Hex,
					Amount:        vout.Value,
					Confirmations: int64(tx.Confirmations),
					Spendable:     true,
				}
				utxos = append(utxos, utxo)
			}
		}
	}

	// 返回UTXO集合
	return utxos, nil
}

// buildTxOut 构建一个比特币交易输出（TxOut）
func buildTxOut(addr, network string, amount int64) (*wire.TxOut, []byte, error) {
	// 解析比特币地址
	destinationAddress, err := btcutil.DecodeAddress(addr, utils.GetNetwork(network))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse bitcoin address")
	}

	// 生成支付到地址的脚本
	pkScript, err := txscript.PayToAddrScript(destinationAddress)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate script")
	}

	// 创建一个新的交易输出，金额单位为 satoshis
	return wire.NewTxOut(amount, pkScript), pkScript, nil
}

// buildTxIn 构建一个比特币交易输入（TxIn）
func buildTxIn(wif *btcutil.WIF, amount int64, txOut *wire.TxOut, network string) (*wire.MsgTx, error) {
	// 解析比特币地址
	fromAddr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), utils.GetNetwork(network))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse bitcoin address")
	}

	// 获取UTXOs
	utxos, err := getUTXOs(fromAddr.EncodeAddress(), network)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get utxos")
	}

	msgTx := wire.NewMsgTx(wire.TxVersion)
	// 创建一个新的交易输入，金额单位为 satoshis
	totalInput := int64(0)
	for _, utxo := range utxos {
		// totalInput 大于 amount，用于计算交易费
		if totalInput > amount {
			break
		}
		txHash, err := chainhash.NewHashFromStr(utxo.TxID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse transaction hash")
		}

		txIn := wire.NewTxIn(&wire.OutPoint{Hash: *txHash, Index: uint32(utxo.Vout)}, nil, nil)
		msgTx.AddTxIn(txIn)
		totalInput += int64(utxo.Amount * decimals)
	}
	msgTx.AddTxOut(txOut)

	// 交易费
	fee := int64(msgTx.SerializeSize())

	// 找零
	change := totalInput - amount
	if change > fee {
		changePkScript, err := txscript.PayToAddrScript(fromAddr)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate change script")
		}
		txOut := wire.NewTxOut(change-fee, changePkScript)
		msgTx.AddTxOut(txOut)
	}

	// 签署交易，适用P2WPKH地址
	for i, txIn := range msgTx.TxIn {
		prevOutputScript, err := hex.DecodeString(utxos[i].ScriptPubKey)
		if err != nil {
			panic(err)
		}
		txHash, err := chainhash.NewHashFromStr(utxos[i].TxID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse transaction hash")
		}
		outPoint := wire.OutPoint{Hash: *txHash, Index: uint32(utxos[i].Vout)}
		prevOutputFetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
			outPoint: {Value: int64(utxos[i].Amount * 1e8), PkScript: prevOutputScript}, // 假设前一个输出的金额是100000 satoshis
		})
		sigHashes := txscript.NewTxSigHashes(msgTx, prevOutputFetcher)
		sigScript, err := txscript.WitnessSignature(msgTx, sigHashes, int(utxos[i].Vout), int64(utxos[i].Amount*1e8), prevOutputScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, errors.Wrap(err, "failed to sign transaction")
		}
		txIn.Witness = sigScript
	}
	return msgTx, nil
}

func listKeys(keyFile string) ([]string, error) {
	return storage.NewLocalStorage(keyFile).ListKeys()
}

func saveKey(keyFile, encryptKey string) error {
	return storage.NewLocalStorage(keyFile).SaveKey(encryptKey)
}

func saveMnemonic(keyFile, mnemonic string) error {
	return storage.NewLocalStorage(keyFile).Save(mnemonic)
}

func loadMnemonic(keyFile string) (string, error) {
	return storage.NewLocalStorage(keyFile).Load()
}

func askOneString(msg string) (string, error) {
	var intput string
	prompt := &survey.Input{
		Message: msg,
	}
	err := survey.AskOne(prompt, &intput)
	if err != nil {
		return "", errors.Wrap(err, "failed to get input")
	}
	return intput, nil
}

func askOneNumber(msg string) (uint64, error) {
	var number uint64
	prompt := &survey.Input{
		Message: msg,
	}
	err := survey.AskOne(prompt, &number)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get number")
	}
	return number, nil
}

func askNetwork() (string, error) {
	var network string
	prompt := &survey.Select{
		Message: "请选择网络类型：",
		Options: []string{"mainnet", "testnet"},
	}
	err := survey.AskOne(prompt, &network)
	if err != nil {
		return "", errors.Wrap(err, "获取网络类型失败")
	}
	return network, nil
}
