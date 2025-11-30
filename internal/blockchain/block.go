package blockchain

// internal/blockchain/block.go
// Block data structure and hash helpers.

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Block represents a blockchain block.
// Transactions are referenced by their txid strings to keep blocks compact.
type Block struct {
	Index        int      `json:"index"`
	Timestamp    int64    `json:"timestamp"`
	Transactions []string `json:"transactions"` // txids
	PrevHash     string   `json:"prev_hash"`
	Nonce        int64    `json:"nonce"`
	Hash         string   `json:"hash"`
}

// calcHash computes the SHA256 hex string of block header fields.
func calcHash(b *Block) string {
	var buf bytes.Buffer
	buf.WriteString(strconv.Itoa(b.Index))
	buf.WriteString("|")
	buf.WriteString(strconv.FormatInt(b.Timestamp, 10))
	buf.WriteString("|")
	buf.WriteString(b.PrevHash)
	buf.WriteString("|")
	buf.WriteString(strconv.FormatInt(b.Nonce, 10))
	buf.WriteString("|")
	if len(b.Transactions) > 0 {
		buf.WriteString(strings.Join(b.Transactions, ","))
	}
	sum := sha256.Sum256(buf.Bytes())
	return fmt.Sprintf("%x", sum[:])
}

// NewGenesis returns a genesis block instance (deterministic).
func NewGenesis() Block {
	g := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []string{},
		PrevHash:     "0",
		Nonce:        0,
	}
	g.Hash = calcHash(&g)
	return g
}

// ValidateBasic checks header/hash consistency (no PoW check).
func (b *Block) ValidateBasic() bool {
	return calcHash(b) == b.Hash
}

// Marshal pretty JSON for debugging
func (b *Block) ToJSON() string {
	j, _ := json.MarshalIndent(b, "", "  ")
	return string(j)
}
