package core

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Difficulty 挖矿难度，表示哈希值需要以多少个0开头
const Difficulty = 3

// Transaction 交易结构体，表示一笔转账交易
type Transaction struct {
	From      string `json:"from"`      // 发送方地址
	To        string `json:"to"`        // 接收方地址
	Amount    int    `json:"amount"`    // 转账金额
	Signature string `json:"signature"` // 交易签名，用于验证交易有效性
}

// Block 区块结构体，包含区块的所有信息
type Block struct {
	Index        int           `json:"index"`         // 区块索引（高度）
	Timestamp    int64         `json:"timestamp"`     // 时间戳
	Transactions []Transaction `json:"transactions"`  // 包含的交易列表
	PrevHash     string        `json:"prev_hash"`     // 前一个区块的哈希值
	Nonce        int64         `json:"nonce"`         // 工作量证明的随机数
	Hash         string        `json:"hash"`          // 当前区块的哈希值
}

// Blockchain 区块链结构体
type Blockchain struct {
	chain       []Block
	transaction []Transaction
	mutex       sync.Mutex
}

// NewKeyPair 生成新的椭圆曲线密钥对，用于创建钱包地址
func NewKeyPair() (*ecdsa.PrivateKey, string) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubBytes := elliptic.Marshal(priv.PublicKey.Curve, priv.PublicKey.X, priv.PublicKey.Y)
	return priv, hex.EncodeToString(pubBytes)
}

// HashTransaction 计算交易的哈希值，用于签名和验证
func HashTransaction(tx Transaction) []byte {
	data := tx.From + "|" + tx.To + "|" + strconv.Itoa(tx.Amount)
	h := sha256.Sum256([]byte(data))
	return h[:]
}

// SignTransaction 使用私钥对交易进行签名
func SignTransaction(priv *ecdsa.PrivateKey, tx Transaction) (string, error) {
	h := HashTransaction(tx)
	sig, err := ecdsa.SignASN1(rand.Reader, priv, h)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sig), nil
}

// VerifyTransaction 验证交易签名的有效性
func VerifyTransaction(tx Transaction) bool {
	pubBytes, err := hex.DecodeString(tx.From)
	if err != nil {
		return false
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil {
		return false
	}
	pub := ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
	sigBytes, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}
	h := HashTransaction(tx)
	return ecdsa.VerifyASN1(&pub, h, sigBytes)
}

// CalculateHash 计算区块的哈希值
func CalculateHash(b Block) string {
	txBytes, _ := json.Marshal(b.Transactions)
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) + string(txBytes) + b.PrevHash + strconv.FormatInt(b.Nonce, 10)
	h := sha256.Sum256([]byte(record))
	return fmt.Sprintf("%x", h)
}

// MineBlock 挖掘新区块，通过工作量证明找到满足难度要求的哈希值
func MineBlock(transactions []Transaction, prev Block) Block {
	newBlock := Block{
		Index:        prev.Index + 1,        // 新区块索引为前一区块索引+1
		Timestamp:    time.Now().Unix(),     // 设置当前时间戳
		Transactions: transactions,          // 设置要打包的交易
		PrevHash:     prev.Hash,             // 设置前一区块哈希
		Nonce:        0,                     // 初始化随机数为0
		Hash:         "",                    // 初始化哈希为空
	}
	// 不断尝试不同的Nonce值直到找到满足难度要求的哈希
	for {
		newBlock.Hash = CalculateHash(newBlock)
		// 检查哈希是否满足难度要求（以指定数量的0开头）
		if strings.HasPrefix(newBlock.Hash, strings.Repeat("0", Difficulty)) {
			break
		}
		newBlock.Nonce++ // 增加Nonce值继续尝试
	}
	return newBlock
}

// NewBlockchain 创建新的区块链实例
func NewBlockchain() *Blockchain {
	bc := &Blockchain{
		chain:       []Block{},
		transaction: []Transaction{},
	}
	bc.initGenesis()
	return bc
}

// initGenesis 初始化创世区块
func (bc *Blockchain) initGenesis() {
	genesis := Block{
		Index:        0,                     // 创世区块索引为0
		Timestamp:    time.Now().Unix(),     // 设置当前时间戳
		Transactions: []Transaction{},       // 创世区块不包含交易
		PrevHash:     "0",                   // 前一区块哈希为"0"
		Nonce:        0,                     // 随机数初始化为0
	}
	genesis.Hash = CalculateHash(genesis)   // 计算创世区块哈希
	bc.chain = []Block{genesis}             // 将创世区块加入区块链
}

// AddBlock 向区块链添加新区块
func (bc *Blockchain) AddBlock(b Block) bool {
	bc.mutex.Lock()                         // 加锁保护区块链数据
	defer bc.mutex.Unlock()                 // 函数结束时解锁
	last := bc.chain[len(bc.chain)-1]       // 获取最后一个区块

	// 验证区块有效性：
	// 1. 前一区块哈希必须匹配
	// 2. 区块哈希必须正确
	// 3. 区块哈希必须满足难度要求
	if b.PrevHash != last.Hash || CalculateHash(b) != b.Hash || !strings.HasPrefix(b.Hash, strings.Repeat("0", Difficulty)) {
		return false
	}

	bc.chain = append(bc.chain, b)          // 将新区块添加到区块链末尾
	return true
}

// GetBlocks 获取区块链副本
func (bc *Blockchain) GetBlocks() []Block {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	chain := make([]Block, len(bc.chain))
	copy(chain, bc.chain)
	return chain
}

// AddTransaction 添加交易到交易池
func (bc *Blockchain) AddTransaction(tx Transaction) bool {
	// 首先验证交易签名的有效性
	if !VerifyTransaction(tx) {
		return false
	}

	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// 检查交易是否已经在交易池中
	for _, t := range bc.transaction {
		if t.Signature == tx.Signature {
			return false
		}
	}

	// 将交易添加到交易池
	bc.transaction = append(bc.transaction, tx)
	return true
}

// GetTransactions 获取交易池副本
func (bc *Blockchain) GetTransactions() []Transaction {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()
	txs := make([]Transaction, len(bc.transaction))
	copy(txs, bc.transaction)
	return txs
}

// ClearTransactions 从交易池中移除已打包的交易
func (bc *Blockchain) ClearTransactions(txs []Transaction) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	newPool := []Transaction{}              // 创建新的交易池

	// 遍历当前交易池中的所有交易
	for _, p := range bc.transaction {
		found := false
		// 检查该交易是否在要移除的交易列表中
		for _, t := range txs {
			if p.Signature == t.Signature {
				found = true
				break
			}
		}
		// 如果不在要移除的列表中，则保留在新交易池中
		if !found {
			newPool = append(newPool, p)
		}
	}
	bc.transaction = newPool                 // 更新交易池
}