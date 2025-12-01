package blockchain

import (
 	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// ProofOfWork 结构体，封装区块和目标值
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork 创建新的工作量证明实例
// difficulty: 难度系数（十六进制前导零数量）
func NewProofOfWork(block *Block, difficulty int) *ProofOfWork {
	// 将十六进制前导零数量转换为比特位数（1 hex char = 4 bits）
	bits := difficulty * 4
	target := big.NewInt(1)
	target.Lsh(target, uint(256-bits))
	return &ProofOfWork{block, target}
}

// prepareData 准备用于哈希计算的数据
func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	transactions := strings.Join(pow.block.Transactions, "")
	data := fmt.Sprintf("%d%d%s%s%d",
		pow.block.Index,
		pow.block.Timestamp,
		pow.block.PrevHash,
		transactions,
		nonce)
	return []byte(data)
}

// Run 执行挖矿过程，寻找满足条件的nonce
func (pow *ProofOfWork) Run() (int64, []byte) {
	var hashInt big.Int
	nonce := int64(0)

	for {
		data := pow.prepareData(nonce)
		hash := sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			return nonce, hash[:]
		}
		nonce++
	}
}

// Validate 验证区块是否满足工作量证明
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int
	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(pow.target) == -1
}

// MineBlock 使用改进的PoW算法挖取新区块
func MineBlock(prev Block, txids []string, difficulty int) Block {
	b := Block{
		Index:        prev.Index + 1,
		Timestamp:    time.Now().Unix(),
		Transactions: txids,
		PrevHash:     prev.Hash,
		Nonce:        0,
	}

	pow := NewProofOfWork(&b, difficulty)
	nonce, hash := pow.Run()
	b.Nonce = nonce
	b.Hash = hex.EncodeToString(hash)
	return b
}

// CheckPoW 验证区块是否满足PoW要求
func CheckPoW(b *Block, difficulty int) bool {
	pow := NewProofOfWork(b, difficulty)
	return pow.Validate()
}