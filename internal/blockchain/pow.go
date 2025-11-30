package blockchain

// internal/blockchain/pow.go
// Proof-of-Work (PoW) utilities: mining and difficulty target helpers.

import (
	"strings"
	"time"
)

// MineBlock performs simple PoW by incrementing Nonce until hash has
// `difficulty` leading hex '0' characters.
// Returns the mined block (with Hash and Nonce set).
func MineBlock(prev Block, txids []string, difficulty int) Block {
	b := Block{
		Index:        prev.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: txids,
		PrevHash:     prev.Hash,
		Nonce:        0,
	}
	prefix := strings.Repeat("0", difficulty)
	for {
		b.Hash = calcHash(&b)
		if strings.HasPrefix(b.Hash, prefix) {
			return b
		}
		b.Nonce++
	}
}

// CheckPoW verifies that block hash has required leading zeros.
func CheckPoW(b *Block, difficulty int) bool {
	if !b.ValidateBasic() {
		return false
	}
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(b.Hash, prefix)
}
