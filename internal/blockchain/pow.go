package blockchain

// internal/blockchain/pow.go
// 工作量证明（PoW）工具：挖矿和难度目标辅助函数

import (
	"strings"
	"time"
)

// MineBlock 通过递增Nonce直到哈希具有指定数量的前导十六进制'0'字符来执行简单的工作量证明
// 返回挖取的区块（已设置Hash和Nonce）
func MineBlock(prev Block, txids []string, difficulty int) Block {
	// 创建新区块结构
	b := Block{
		Index:        prev.Index + 1,  // 区块索引递增
		Timestamp:    time.Now().Unix(), // 当前时间戳
		Transactions: txids,           // 要打包的交易ID列表
		PrevHash:     prev.Hash,       // 前一区块哈希
		Nonce:        0,               // 初始随机数为0
	}
	
	// 构造目标前缀：difficulty个'0'字符
	prefix := strings.Repeat("0", difficulty)
	
	// 不断尝试不同的Nonce值直到找到满足难度要求的哈希
	for {
		b.Hash = calcHash(&b) // 计算当前区块哈希
		// 检查哈希是否满足难度要求（以指定前缀开头）
		if strings.HasPrefix(b.Hash, prefix) {
			return b // 满足条件，返回挖取的区块
		}
		b.Nonce++ // 增加随机数继续尝试
	}
}

// CheckPoW 验证区块哈希是否具有所需的前导零
func CheckPoW(b *Block, difficulty int) bool {
	// 首先进行基本头部哈希一致性检查
	if !b.ValidateBasic() {
		return false
	}
	
	// 构造目标前缀：difficulty个'0'字符
	prefix := strings.Repeat("0", difficulty)
	
	// 检查区块哈希是否以目标前缀开头
	return strings.HasPrefix(b.Hash, prefix)
}