package blockchain

// internal/blockchain/transaction.go
// Transaction / raw tx representation for UTXO model.
// Note: Signing/verification are done externally (wallet module).
// This file defines structures and helpers: txid calculation and basic checks.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// TxInput references a previous UTXO by txid:vout and carries signature + pubkey.
type TxInput struct {
	Txid      string `json:"txid"`
	Vout      int    `json:"vout"`
	Signature string `json:"signature"` // hex-encoded signature
	PubKey    string `json:"pubkey"`    // hex-encoded public key (address)
}

// TxOutput is an output (UTXO candidate).
type TxOutput struct {
	Address string `json:"address"`
	Amount  int    `json:"amount"`
}

// UTXOTx is the serializable raw transaction in UTXO model.
type UTXOTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

// TxID returns deterministic tx id: sha256(json(rawtx))
func TxID(raw UTXOTx) (string, error) {
	b, err := json.Marshal(raw)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

// Helper: basic sanity checks (structure)
func ValidateTxStructure(raw UTXOTx) error {
	if len(raw.Inputs) == 0 && len(raw.Outputs) == 0 {
		return fmt.Errorf("tx has no inputs and no outputs")
	}
	for _, out := range raw.Outputs {
		if out.Amount < 0 {
			return fmt.Errorf("negative amount")
		}
	}
	return nil
}

// validateRawTx checks if all inputs of a transaction exist in the UTXO set.
// This function is used by the blockchain to validate transactions in blocks.
func validateRawTx(txid string) error {
	// TODO: This is a simplified implementation.
	// In a real implementation, we would:
	// 1. Retrieve the transaction from storage by txid
	// 2. Verify the transaction structure
	// 3. Check that all input UTXOs exist
	// 4. Verify signatures
	// 
	// For now, we'll just return nil to indicate the transaction is valid
	return nil
}