package wallet

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/mengbin92/wallet/internal/models"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
)

// SaveToKeystore 将地址导出为 V3 keystore JSON 文件（每个地址一个文件）
func SaveToKeystore(addresses []*models.Address, password, outDir string) error {
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, 0700); err != nil {
			return errors.Wrap(err, "failed to create keystore output directory")
		}
	}

	for _, a := range addresses {
		privBytes, err := crypto.HexToECDSA(a.PrivateKey)
		if err != nil {
			return errors.Wrapf(err, "failed to parse private key for %s", a.PrivateKey)
		}

		keyJSON, err := keystore.EncryptKey(&keystore.Key{
			Address:    crypto.PubkeyToAddress(privBytes.PublicKey),
			PrivateKey: privBytes,
		}, password, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			return errors.Wrapf(err, "failed to encrypt key for %s", a.Address)
		}

		filename := filepath.Join(outDir, fmt.Sprintf("%s.json", a.Address))
		if err := os.WriteFile(filename, keyJSON, 0600); err != nil {
			return errors.Wrapf(err, "failed to write keystore for %s", a.Address)
		}
	}
	return nil
}

// LoadAllKeys 从 keystore 目录加载所有地址
func LoadAllKeys(keystoreDir, password string, logger *log.Logger) ([]*keystore.Key, error) {
	entries, err := os.ReadDir(keystoreDir)
	if err != nil {
		return nil, errors.Wrap(err, "read keystore dir")
	}

	var keys []*keystore.Key

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		keyPath := filepath.Join(keystoreDir, entry.Name())

		logger.Printf("loading key: %s", keyPath)
		key, err := LoadKey(keyPath, password,logger)
		if err != nil {
			logger.Printf("skip %s, decrypt error: %v\n", keyPath, err)
			continue
		}

		keys = append(keys, key)
	}

	if len(keys) == 0 {
		return nil, errors.New(fmt.Sprintf("no valid keys found in %s", keystoreDir))
	}

	return keys, nil
}

func LoadKey(keyPath, password string, logger *log.Logger) (*keystore.Key, error) {
	logger.Printf("loading key: %s", keyPath)
	keyJSON, err := os.ReadFile(keyPath)
	if err != nil {
		logger.Printf("file %s, read error: %s\n", keyPath, err)
		return nil, err
	}

	key, err := keystore.DecryptKey(keyJSON, password)
	if err != nil {
		logger.Printf("file %s, decrypt error: %s\n", keyPath, err.Error())
		return nil, err
	}
	return key, nil
}

// ExportPrivateKeyHex 返回私钥 hex 字符串
func ExportPrivateKeyHex(key *keystore.Key) string {
	return fmt.Sprintf("%x", crypto.FromECDSA(key.PrivateKey))
}

// SaveAddressesToExcel 只导出一列 address
func SaveAddressesToExcel(addresses []*models.Address, outFile string) error {
	f := excelize.NewFile()

	// 创建默认工作表
	sheet := "Sheet1"
	index := f.NewSheet(sheet)
	f.SetActiveSheet(index)

	// 设置表头
	f.SetCellValue(sheet, "A1", "address")

	// 写入地址
	for i, addr := range addresses {
		cell := fmt.Sprintf("A%d", i+2) // 从第二行开始
		f.SetCellValue(sheet, cell, addr.Address)
	}

	// 保存文件
	if err := f.SaveAs(outFile); err != nil {
		return errors.Wrap(err, "failed to save excel file")
	}
	return nil
}
