package wallet

// internal/wallet/keystore.go
// Simple keystore: encrypt/decrypt private key JSON with password (AES or PBKDF omitted for brevity).
// **For demo only** — do NOT use this keystore in production (no strong encryption).

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
)

type KeyStore struct {
	PrivHex string `json:"priv_hex"` // hex-encoded private key bytes (DER)
}

// ExportKey returns a JSON-encoded keystore string containing the private key.
func ExportKey(priv *ecdsa.PrivateKey) (string, error) {
	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", errors.New("could not marshal private key")
	}
	ks := KeyStore{PrivHex: hex.EncodeToString(der)}
	b, err := json.Marshal(ks)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ImportKey parses a keystore JSON and returns a private key.
func ImportKey(jsonStr string) (*ecdsa.PrivateKey, error) {
	var ks KeyStore
	if err := json.Unmarshal([]byte(jsonStr), &ks); err != nil {
		return nil, err
	}
	der, err := hex.DecodeString(ks.PrivHex)
	if err != nil {
		return nil, err
	}
	privKey, err := x509.ParseECPrivateKey(der)
	if err != nil {
		return nil, err
	}
	return privKey, nil
}