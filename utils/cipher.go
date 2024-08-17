package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcutil/base58"
	"github.com/pkg/errors"
	"golang.org/x/crypto/scrypt"
)

// AesEncrypt encrypts the given data using the provided passphrase
func AesEncrypt(data []byte, passphrase string) (string, error) {
	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return hex.EncodeToString(ciphertext), nil
}

// AesDecrypt decrypts the given data using the provided passphrase
func AesDecrypt(encryptedData string, passphrase string) ([]byte, error) {
	data, err := hex.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(createHash(passphrase))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// createHash creates a hash of the passphrase
func createHash(passphrase string) []byte {
	hash := sha256.New()
	hash.Write([]byte(passphrase))
	return hash.Sum(nil)
}

// 双SHA256哈希
func doubleSha256(b []byte) []byte {
	hash := sha256.Sum256(b)
	hash = sha256.Sum256(hash[:])
	return hash[:]
}

// RIPEMD-160哈希
// 已弃用：RIPEMD-160 是旧版哈希，不应用于新应用程序。此外，这个包现在和将来都不会提供优化的实现。
// 所以 使用SHA-256（crypto/sha256）替代
func ripemd160Hash(b []byte) []byte {
	// hasher := ripemd160.New()
	// hasher.Write(b)
	// return hasher.Sum(nil)
	hash := sha256.Sum256(b)
	return hash[:]
}

// BIP38加密
func BIP38Encrypt(wifStr, passphrase string) (string, error) {
	// 尝试解码WIF格式的私钥
	wif, err := btcutil.DecodeWIF(wifStr)
	if err != nil {
		return "", errors.Wrap(err, "无法解码WIF格式私钥")
	}

	// 生成盐值 (来自于公钥的RIPEMD-160前4字节)
	salt := ripemd160Hash(wif.PrivKey.PubKey().SerializeCompressed())[:4]

	// 使用scrypt生成密钥
	scryptKey, err := scrypt.Key([]byte(passphrase), salt, 16384, 8, 8, 64)
	if err != nil {
		return "", fmt.Errorf("scrypt密钥生成失败: %v", err)
	}

	derivedHalf1 := scryptKey[:32]
	derivedHalf2 := scryptKey[32:]

	block, err := aes.NewCipher(derivedHalf2)
	if err != nil {
		return "", fmt.Errorf("AES密码生成失败: %v", err)
	}

	// 私钥的前16字节和后16字节加密
	xorBytes := func(a, b []byte) []byte {
		n := len(a)
		xored := make([]byte, n)
		for i := 0; i < n; i++ {
			xored[i] = a[i] ^ b[i]
		}
		return xored
	}

	privKeyBytes := wif.PrivKey.Serialize()
	encryptedHalf1 := xorBytes(privKeyBytes[:16], derivedHalf1[:16])
	encryptedHalf2 := xorBytes(privKeyBytes[16:], derivedHalf1[16:])

	encryptedBytes := make([]byte, 32)
	block.Encrypt(encryptedBytes[:16], encryptedHalf1)
	block.Encrypt(encryptedBytes[16:], encryptedHalf2)

	// 构建BIP38格式
	bip38Key := append([]byte{0x01, 0x42, 0xC0}, salt...)
	bip38Key = append(bip38Key, encryptedBytes...)

	// 加入校验和
	checksum := doubleSha256(bip38Key)[:4]
	bip38Key = append(bip38Key, checksum...)

	// Base58编码
	return base58.Encode(bip38Key), nil
}

func BIP38Decrypt(encryptedKey, passphrase, network string) (string, error) {
	// Base58解码
	decoded := base58.Decode(encryptedKey)

	// 检查校验和
	checksum := decoded[len(decoded)-4:]
	hash := doubleSha256(decoded[:len(decoded)-4])
	if !reflect.DeepEqual(hash[:4], checksum) {
		return "", errors.New("校验和不匹配")
	}

	// 从加密字节中提取盐值
	salt := decoded[3:7]
	encryptedHalf1 := decoded[7:23]
	encryptedHalf2 := decoded[23:39]

	// 使用scrypt生成密钥
	scryptKey, err := scrypt.Key([]byte(passphrase), salt, 16384, 8, 8, 64)
	if err != nil {
		return "", errors.Wrap(err, "scrypt密钥生成失败")
	}

	derivedHalf1 := scryptKey[:32]
	derivedHalf2 := scryptKey[32:]

	block, err := aes.NewCipher(derivedHalf2)
	if err != nil {
		return "", errors.Wrap(err, "AES密码生成失败")
	}

	decryptedHalf1 := make([]byte, 16)
	block.Decrypt(decryptedHalf1, encryptedHalf1)
	decryptedHalf2 := make([]byte, 16)
	block.Decrypt(decryptedHalf2, encryptedHalf2)

	privKeyBytes := append(decryptedHalf1, decryptedHalf2...)
	for i := 0; i < 32; i++ {
		privKeyBytes[i] ^= derivedHalf1[i]
	}

	// 将解密后的私钥字节切片转换为 *btcec.PrivateKey 类型
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

	// 使用解密的私钥生成WIF格式
	wif, err := btcutil.NewWIF(privKey, GetNetwork(network), true)
	if err != nil {
		return "", errors.Wrap(err, "生成WIF失败")
	}

	return wif.String(), nil
}
