package wallet

// internal/wallet/wallet.go
// Wallet utilities using secp256k1 (via go-ethereum crypto library).

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"

	"github.com/ethereum/go-ethereum/crypto"
)

// NewKey generates a new secp256k1 private key and returns (privKey, pubKeyHex).
func NewKey() (*ecdsa.PrivateKey, string, error) {
	priv, err := crypto.GenerateKey()
	if err != nil {
		return nil, "", err
	}
	pubBytes := crypto.FromECDSAPub(&priv.PublicKey)
	pubHex := hex.EncodeToString(pubBytes)
	return priv, pubHex, nil
}

// PubKeyToAddress converts a public key hex string to a canonical address (hex).
// Here we just return pubkey hex (no hash/truncation) — for demo.
// In real chain you may hash and take last 20 bytes like Ethereum.
func PubKeyToAddress(pubHex string) string {
	return pubHex
}

// SignData signs the given data using the private key, returns hex-encoded signature.
// Data is assumed to be hashed appropriately before calling (e.g. SHA256).
func SignData(priv *ecdsa.PrivateKey, data []byte) (string, error) {
	sig, err := crypto.Sign(data, priv) // produces [R || S || V] 65 bytes including recovery id
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// VerifySignature verifies that the signature (hex) on data (hash) is signed by pubHex public key.
// Returns true if valid.
func VerifySignature(pubHex string, sigHex string, data []byte) (bool, error) {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return false, err
	}
	sigBytes, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, err
	}
	// Note: eth crypto expects signature with V at end; use RecoverPubkey to verify equivalence
	recoveredPub, err := crypto.Ecrecover(data, sigBytes)
	if err != nil {
		return false, err
	}
	// compare recoveredPub with pubBytes
	return hex.EncodeToString(recoveredPub) == hex.EncodeToString(pubBytes), nil
}

// Simple wrapper: returns error if not valid
func VerifyRaw(pubHex string, sigHex string, data []byte) error {
	valid, err := VerifySignature(pubHex, sigHex, data)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("signature invalid")
	}
	return nil
}