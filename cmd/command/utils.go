package command

import (
	"encoding/hex"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/mengbin92/wallet/utils"
	"github.com/pkg/errors"
)

// getUTXOs 获取与地址相关的所有UTXO
// 该函数使用SearchRawTransactionsVerbose方法获取与地址相关的所有交易，然后遍历每个交易的输出，
func getUTXOs(addr, network string) ([]*btcjson.ListUnspentResult, error) {
	// 解析比特币地址
	address, err := btcutil.DecodeAddress(addr, utils.GetNetwork(network))
	if err != nil {
		return nil, errors.Wrap(err, "解析比特币地址失败")
	}

	// 使用SearchRawTransactionsVerbose获取与地址相关的所有交易
	transactions, err := client.SearchRawTransactionsVerbose(address, 0, 100, true, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "获取与地址相关的所有交易失败")
	}

	// 用于存储UTXO的切片
	utxos := []*btcjson.ListUnspentResult{}

	// 遍历所有交易
	for _, tx := range transactions {
		// 将交易ID字符串转换为链哈希对象
		txid, err := chainhash.NewHashFromStr(tx.Txid)
		if err != nil {
			return nil, errors.Wrap(err, "将交易ID字符串转换为链哈希对象失败")
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
				return nil, errors.Wrap(err, "获取交易输出失败")
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
		return nil, nil, errors.Wrap(err, "解析比特币地址失败")
	}

	// 生成支付到地址的脚本
	pkScript, err := txscript.PayToAddrScript(destinationAddress)
	if err != nil {
		return nil, nil, errors.Wrap(err, "生成支付到地址的脚本失败")
	}

	// 创建一个新的交易输出，金额单位为 satoshis
	return wire.NewTxOut(amount, pkScript), pkScript, nil
}

// buildTxIn 构建一个比特币交易输入（TxIn）
func buildTxIn(wif *btcutil.WIF, amount int64, txOut *wire.TxOut, network string) (*wire.MsgTx, error) {
	// 解析比特币地址
	fromAddr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(wif.SerializePubKey()), utils.GetNetwork(network))
	if err != nil {
		return nil, errors.Wrap(err, "解析比特币地址失败")
	}

	// 获取UTXOs
	utxos, err := getUTXOs(fromAddr.EncodeAddress(), network)
	if err != nil {
		return nil, errors.Wrap(err, "获取UTXOs失败")
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
			return nil, errors.Wrap(err, "解析交易哈希失败")
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
			return nil, errors.Wrap(err, "生成找零地址的脚本失败")
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
			return nil, errors.Wrap(err, "解析交易哈希失败")
		}
		outPoint := wire.OutPoint{Hash: *txHash, Index: uint32(utxos[i].Vout)}
		prevOutputFetcher := txscript.NewMultiPrevOutFetcher(map[wire.OutPoint]*wire.TxOut{
			outPoint: {Value: int64(utxos[i].Amount * 1e8), PkScript: prevOutputScript}, // 假设前一个输出的金额是100000 satoshis
		})
		sigHashes := txscript.NewTxSigHashes(msgTx, prevOutputFetcher)
		sigScript, err := txscript.WitnessSignature(msgTx, sigHashes, int(utxos[i].Vout), int64(utxos[i].Amount*1e8), prevOutputScript, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return nil, errors.Wrap(err, "签名交易失败")
		}
		txIn.Witness = sigScript
	}
	return msgTx, nil
}

// buildRawTx 构建一个未签名的比特币交易
func buildRawTx(wif *btcutil.WIF, addr, network string, amount int64) (*wire.MsgTx, error) {

	// 构建交易输出
	txOut, _, err := buildTxOut(addr, network, amount)
	if err != nil {
		return nil, errors.Wrap(err, "构建交易输出失败")
	}

	// 构建交易输入
	msgTx, err := buildTxIn(wif, amount, txOut, network)
	if err != nil {
		return nil, errors.Wrap(err, "构建交易输入失败")
	}

	return msgTx, nil
}
