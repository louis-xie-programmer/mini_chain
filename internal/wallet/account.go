package wallet

// internal/wallet/account.go
// 表示钱包账户：地址 + （可选）内存中的私钥

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
)

// Account 钱包账户结构体
type Account struct {
	Address string            // 公钥地址（十六进制）
	Private *ecdsa.PrivateKey // 私钥；如果为只读账户则可能为nil
}

// NewAccount 创建包含私钥的新账户
func NewAccount() (*Account, error) {
	// 使用wallet.go中的NewKey函数
	priv, pubHex, err := NewKey()
	if err != nil {
		return nil, err
	}
	return &Account{Address: pubHex, Private: priv}, nil
}

// FromPrivate 从现有私钥重建账户
// priv: 私钥
func FromPrivate(priv *ecdsa.PrivateKey) *Account {
	// 直接使用椭圆曲线包序列化公钥
	pubBytes := elliptic.Marshal(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	pubHex := hex.EncodeToString(pubBytes)
	return &Account{Address: pubHex, Private: priv}
}