package blockchain

// internal/blockchain/block.go
// 区块数据结构定义和哈希计算相关的辅助函数

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// Block 区块结构体，代表区块链中的一个区块
type Block struct {
	Index        int      `json:"index"`        // 区块索引（高度）
	Timestamp    int64    `json:"timestamp"`    // 区块生成时间戳
	Transactions []string `json:"transactions"` // 包含的交易ID列表
	PrevHash     string   `json:"prev_hash"`    // 前一个区块的哈希值
	Nonce        int64    `json:"nonce"`        // 工作量证明的随机数
	Hash         string   `json:"hash"`         // 当前区块的哈希值
}

// calcHash 计算区块头部字段的SHA256哈希值
// 该函数用于生成区块的唯一标识，包含区块索引、时间戳、前一区块哈希、随机数和交易ID等信息
func calcHash(b *Block) string {
    var buf bytes.Buffer

    // 1. 固定顺序写入基本字段（不含 Hash 自身）
    buf.WriteString(strconv.Itoa(b.Index))
    buf.WriteString("|")
    buf.WriteString(strconv.FormatInt(b.Timestamp, 10))
    buf.WriteString("|")
    buf.WriteString(b.PrevHash)
    buf.WriteString("|")
    buf.WriteString(strconv.FormatInt(b.Nonce, 10))
    buf.WriteString("|")

    // 2. 为防止因交易顺序不同导致分叉，先排序
    if len(b.Transactions) > 0 {
        txCopy := make([]string, len(b.Transactions))
        copy(txCopy, b.Transactions)
        sort.Strings(txCopy) // 固定序

        // 使用不可分割的分隔符，避免 "|" 与 "," 模糊边界
        for _, tx := range txCopy {
            buf.WriteString(tx)
            buf.WriteString(";")  // 用分号做交易间隔
        }
    }

    // 3. 计算 SHA-256
    sum := sha256.Sum256(buf.Bytes())
    return fmt.Sprintf("%x", sum[:])
}

// NewGenesis 创建一个创世区块实例（确定性的）
// 创世区块是区块链的第一个区块，具有固定的参数值
func NewGenesis() Block {
	g := Block{
		Index:        0,                     // 创世区块索引为0
		Timestamp:    time.Now().Unix(),     // 当前时间戳
		Transactions: []string{},            // 初始无交易
		PrevHash:     "0",                   // 前一区块哈希为"0"
		Nonce:        0,                     // 随机数初始为0
	}
	g.Hash = calcHash(&g) // 计算并设置创世区块的哈希值
	return g
}

// ValidateBasic 检查区块头部和哈希的一致性（不包括工作量证明检查）
// 用于验证区块的基本完整性
func (b *Block) ValidateBasic() bool {
	return calcHash(b) == b.Hash
}

// ToJSON 将区块转换为美化格式的JSON字符串，用于调试
func (b *Block) ToJSON() string {
	j, _ := json.MarshalIndent(b, "", "  ")
	return string(j)
}