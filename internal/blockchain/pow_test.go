package blockchain

import (
	"math/big"
	"testing"
)

func TestProofOfWork_Basic(t *testing.T) {
	// 测试低难度挖矿（difficulty=1）
	block := &Block{
		Index:        1,
		Timestamp:    123456,
		Transactions: []string{"tx1"},
		PrevHash:     "prev",
		Nonce:        0,
	}
	
	pow := NewProofOfWork(block, 1)
	nonce, hash := pow.Run()

	// 验证哈希值是否满足目标难度
	var hashInt big.Int
	hashInt.SetBytes(hash)
	if hashInt.Cmp(pow.target) >= 0 {
		t.Errorf("哈希值 %x 未达到目标难度 %x", hash, pow.target.Bytes())
	}

	// 验证有效区块
	block.Nonce = nonce
	if !pow.Validate() {
		t.Error("有效区块验证失败")
	}

	// 验证无效区块
	block.Nonce = nonce + 1
	if pow.Validate() {
		t.Error("无效区块通过验证")
	}
}

func TestProofOfWork_DifficultyConversion(t *testing.T) {
	// 验证难度转换逻辑 (difficulty * 4 bits)
	block := &Block{
		Index:        1,
		Timestamp:    123456,
		Transactions: []string{"tx1"},
		PrevHash:     "prev",
		Nonce:        0,
	}
	pow := NewProofOfWork(block, 2)

	// 难度2应对应8 bits (2*4)
	expectedTarget := big.NewInt(1)
	expectedTarget.Lsh(expectedTarget, uint(256-8))

	if pow.target.Cmp(expectedTarget) != 0 {
		t.Errorf("难度转换错误: 期望 %x, 实际 %x", expectedTarget.Bytes(), pow.target.Bytes())
	}
}

func TestProofOfWork_HighDifficulty(t *testing.T) {
	// 测试较高难度（difficulty=4）
	block := &Block{
		Index:        1,
		Timestamp:    123456,
		Transactions: []string{"tx1"},
		PrevHash:     "prev",
		Nonce:        0,
	}
	pow := NewProofOfWork(block, 4)

	// 验证目标值是否正确设置为 2^(256 - bits)
	expectedTarget := big.NewInt(1)
	expectedTarget.Lsh(expectedTarget, uint(240))
	if pow.target.Cmp(expectedTarget) != 0 {
		t.Errorf("目标值错误: 期望 %x, 实际 %x", expectedTarget.Bytes(), pow.target.Bytes())
	}
}
