package wallet

// internal/wallet/account.go
// Represents a wallet account: address + (optional) private key in memory.

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
)

type Account struct {
	Address string            // public address (pubkey hex)
	Private *ecdsa.PrivateKey // private key; may be nil for read-only
}

// NewAccount creates a new account with private key.
func NewAccount() (*Account, error) {
	// Use the NewKey function from wallet.go
	priv, pubHex, err := NewKey()
	if err != nil {
		return nil, err
	}
	return &Account{Address: pubHex, Private: priv}, nil
}

// FromPrivate reconstruct account from existing private key.
func FromPrivate(priv *ecdsa.PrivateKey) *Account {
	// Marshal the public key directly using elliptic package
	pubBytes := elliptic.Marshal(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	pubHex := hex.EncodeToString(pubBytes)
	return &Account{Address: pubHex, Private: priv}
}
