package wallet

// internal/wallet/wallet.go
// 使用secp256k1的钱包工具（通过go-ethereum加密库）

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewKey 生成新的secp256k1私钥并返回（私钥，公钥十六进制）
func NewKey() (*ecdsa.PrivateKey, string, error) {
	priv, err := crypto.GenerateKey() // 生成私钥
	if err != nil {
		return nil, "", err
	}
	pubBytes := crypto.FromECDSAPub(&priv.PublicKey) // 获取公钥字节
	pubHex := hex.EncodeToString(pubBytes)           // 转换为十六进制字符串
	return priv, pubHex, nil
}

// PubKeyToAddress 将公钥十六进制字符串转换为标准地址（十六进制）
// 这里我们只返回公钥十六进制（无哈希/截断）— 用于演示
// 在实际链中，您可能会像以太坊一样哈希并取最后20字节
func PubKeyToAddress(pubHex string) string {
	return pubHex
}

// SignData 使用私钥对数据进行签名，返回十六进制编码的签名
// 数据在调用前应适当进行哈希处理（例如SHA256）
// priv: 私钥
// data: 待签名的数据
func SignData(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	// 生成包含恢复ID的签名[R || S || V] 65字节
	sig, err := crypto.Sign(data, priv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil // 返回十六进制编码的签名
}

// VerifySignature 验证公钥十六进制字符串上的签名（十六进制）在数据（哈希）上是否由pubHex公钥签名
// pubHex: 公钥十六进制字符串
// sigHex: 签名十六进制字符串
// data: 待验证的数据
// 返回签名是否有效和可能的错误
func VerifySignature(pubHex string, sigHex string, data []byte) (bool, error) {
	pubBytes, err := hex.DecodeString(pubHex) // 解码公钥
	if err != nil {
		return false, err
	}
	sigBytes, err := hex.DecodeString(sigHex) // 解码签名
	if err != nil {
		return false, err
	}
	
	// 注意：eth加密期望签名在末尾带有V；使用RecoverPubkey验证等效性
	recoveredPub, err := crypto.Ecrecover(data, sigBytes)
	if err != nil {
		return false, err
	}
	
	// 比较恢复的公钥与原始公钥
	return hex.EncodeToString(recoveredPub) == hex.EncodeToString(pubBytes), nil
}

// VerifyRaw 验证签名的包装函数，如果无效则返回错误
// pubHex: 公钥十六进制字符串
// sigHex: 签名十六进制字符串
// data: 待验证的数据
func VerifyRaw(pubHex string, sigHex string, data []byte) error {
	valid, err := VerifySignature(pubHex, sigHex, data) // 验证签名
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("signature invalid") // 签名无效
	}
	return nil
}