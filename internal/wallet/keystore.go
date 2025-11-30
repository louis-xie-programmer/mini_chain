package wallet

// internal/wallet/keystore.go
// 简单的密钥库：使用密码加密/解密私钥JSON（AES或PBKDF为简洁起见省略）
// **仅供演示** —— 不要在生产中使用此密钥库（无强加密）

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
)

// KeyStore 密钥库存储结构
type KeyStore struct {
	PrivHex string `json:"priv_hex"` // 十六进制编码的私钥字节（DER格式）
}

// ExportKey 返回包含私钥的JSON编码密钥库字符串
// priv: 私钥
func ExportKey(priv *ecdsa.PrivateKey) (string, error) {
	// 将私钥序列化为DER格式
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", errors.New("could not marshal private key")
	}
	
	// 创建密钥库实例
	ks := KeyStore{PrivHex: hex.EncodeToString(der)}
	
	// 序列化为JSON
	b, err := json.Marshal(ks)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ImportKey 解析密钥库JSON并返回私钥
// jsonStr: 密钥库JSON字符串
func ImportKey(jsonStr string) (*ecdsa.PrivateKey, error) {
	var ks KeyStore
	
	// 反序列化JSON到密钥库结构体
	if err := json.Unmarshal([]byte(jsonStr), &ks); err != nil {
		return nil, err
	}
	
	// 解码十六进制私钥
	der, err := hex.DecodeString(ks.PrivHex)
	if err != nil {
		return nil, err
	}
	
	// 解析DER格式的私钥
	privKey, err := x509.ParseECPrivateKey(der)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}