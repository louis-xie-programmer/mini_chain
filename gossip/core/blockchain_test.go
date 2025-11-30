package core

import (
	"testing"
)

// TestNewKeyPair 测试密钥对生成
func TestNewKeyPair(t *testing.T) {
	priv, pub := NewKeyPair()
	
	if priv == nil {
		t.Error("Private key should not be nil")
	}
	
	if pub == "" {
		t.Error("Public key should not be empty")
	}
}

// TestHashTransaction 测试交易哈希计算
func TestHashTransaction(t *testing.T) {
	tx := Transaction{
		From:   "sender",
		To:     "receiver",
		Amount: 100,
	}
	
	hash := HashTransaction(tx)
	
	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}
}

// TestSignAndVerifyTransaction 测试交易签名和验证
func TestSignAndVerifyTransaction(t *testing.T) {
	// 生成密钥对
	priv, pub := NewKeyPair()
	
	// 创建交易
	tx := Transaction{
		From:   pub,
		To:     "receiver",
		Amount: 100,
	}
	
	// 签名交易
	signature, err := SignTransaction(priv, tx)
	if err != nil {
		t.Errorf("Failed to sign transaction: %v", err)
	}
	
	// 设置签名
	tx.Signature = signature
	
	// 验证交易
	if !VerifyTransaction(tx) {
		t.Error("Transaction verification should pass")
	}
	
	// 测试无效签名
	tx.Signature = "invalid_signature"
	if VerifyTransaction(tx) {
		t.Error("Transaction verification should fail with invalid signature")
	}
}

// TestCalculateHash 测试区块哈希计算
func TestCalculateHash(t *testing.T) {
	block := Block{
		Index:     1,
		Timestamp: 1234567890,
		Transactions: []Transaction{
			{
				From:      "sender",
				To:        "receiver",
				Amount:    100,
				Signature: "signature",
			},
		},
		PrevHash: "prev_hash",
		Nonce:    12345,
	}
	
	hash := CalculateHash(block)
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	if len(hash) != 64 {
		t.Errorf("Hash should be 64 characters long, got %d", len(hash))
	}
}

// TestMineBlock 测试区块挖掘
func TestMineBlock(t *testing.T) {
	// 创建前一个区块
	prevBlock := Block{
		Index:     0,
		Timestamp: 1234567890,
		Transactions: []Transaction{},
		PrevHash:  "0",
		Nonce:     0,
	}
	prevBlock.Hash = CalculateHash(prevBlock)
	
	// 创建交易
	transactions := []Transaction{
		{
			From:      "sender",
			To:        "receiver",
			Amount:    100,
			Signature: "signature",
		},
	}
	
	// 挖掘新区块
	newBlock := MineBlock(transactions, prevBlock)
	
	// 验证区块属性
	if newBlock.Index != 1 {
		t.Errorf("Expected index 1, got %d", newBlock.Index)
	}
	
	if newBlock.PrevHash != prevBlock.Hash {
		t.Errorf("Previous hash mismatch")
	}
	
	if len(newBlock.Transactions) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(newBlock.Transactions))
	}
	
	// 验证哈希满足难度要求（以3个0开头）
	if newBlock.Hash[:Difficulty] != "000" {
		t.Errorf("Block hash should start with %d zeros, got %s", Difficulty, newBlock.Hash[:Difficulty])
	}
	
	// 验证哈希计算正确
	calculatedHash := CalculateHash(newBlock)
	if calculatedHash != newBlock.Hash {
		t.Error("Block hash mismatch")
	}
}

// TestBlockchainCreation 测试区块链创建
func TestBlockchainCreation(t *testing.T) {
	bc := NewBlockchain()
	
	if bc == nil {
		t.Error("Blockchain should not be nil")
	}
	
	blocks := bc.GetBlocks()
	if len(blocks) != 1 {
		t.Errorf("Expected 1 block (genesis), got %d", len(blocks))
	}
	
	// 验证创世区块
	genesis := blocks[0]
	if genesis.Index != 0 {
		t.Errorf("Genesis block index should be 0, got %d", genesis.Index)
	}
	
	if genesis.PrevHash != "0" {
		t.Errorf("Genesis block prev hash should be '0', got %s", genesis.PrevHash)
	}
}

// TestAddValidBlock 测试添加有效区块
func TestAddValidBlock(t *testing.T) {
	bc := NewBlockchain()
	
	// 获取创世区块
	blocks := bc.GetBlocks()
	// genesis := blocks[0]  // 注释掉这行因为我们不再使用它
	
	// 创建新区块 (我们使用挖矿来创建有效的区块)
	blocks = bc.GetBlocks()
	genesis := blocks[0]
	
	// 创建交易
	transactions := []Transaction{
		{
			From:      "sender",
			To:        "receiver",
			Amount:    100,
			Signature: "signature",
		},
	}
	
	// 使用挖矿创建新区块
	newBlock := MineBlock(transactions, genesis)
	
	// 添加区块
	result := bc.AddBlock(newBlock)
	
	if !result {
		t.Error("Adding valid block should succeed")
	}
	
	// 验证区块链长度
	blocks = bc.GetBlocks()
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}
}

// TestAddInvalidBlock 测试添加无效区块
func TestAddInvalidBlock(t *testing.T) {
	bc := NewBlockchain()
	
	// 创建无效区块（错误的前哈希）
	invalidBlock := Block{
		Index:     1,
		Timestamp: 1234567890,
		Transactions: []Transaction{
			{
				From:      "sender",
				To:        "receiver",
				Amount:    100,
				Signature: "signature",
			},
		},
		PrevHash: "invalid_prev_hash",
		Nonce:    12345,
	}
	invalidBlock.Hash = CalculateHash(invalidBlock)
	
	// 尝试添加无效区块
	result := bc.AddBlock(invalidBlock)
	
	if result {
		t.Error("Adding invalid block should fail")
	}
	
	// 验证区块链长度未变
	blocks := bc.GetBlocks()
	if len(blocks) != 1 {
		t.Errorf("Expected 1 block, got %d", len(blocks))
	}
}

// TestAddTransaction 测试添加交易
func TestAddTransaction(t *testing.T) {
	bc := NewBlockchain()
	
	// 生成密钥对
	priv, pub := NewKeyPair()
	
	// 创建并签名交易
	tx := Transaction{
		From:   pub,
		To:     "receiver",
		Amount: 100,
	}
	
	signature, err := SignTransaction(priv, tx)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	tx.Signature = signature
	
	// 添加交易
	result := bc.AddTransaction(tx)
	
	if !result {
		t.Error("Adding valid transaction should succeed")
	}
	
	// 验证交易池
	txs := bc.GetTransactions()
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction in pool, got %d", len(txs))
	}
	
	if txs[0].Signature != tx.Signature {
		t.Error("Transaction signature mismatch")
	}
}

// TestAddInvalidTransaction 测试添加无效交易
func TestAddInvalidTransaction(t *testing.T) {
	bc := NewBlockchain()
	
	// 创建无效交易（无有效签名）
	invalidTx := Transaction{
		From:      "sender",
		To:        "receiver",
		Amount:    100,
		Signature: "invalid_signature",
	}
	
	// 添加无效交易
	result := bc.AddTransaction(invalidTx)
	
	if result {
		t.Error("Adding invalid transaction should fail")
	}
	
	// 验证交易池为空
	txs := bc.GetTransactions()
	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions in pool, got %d", len(txs))
	}
}

// TestClearTransactions 测试清除交易
func TestClearTransactions(t *testing.T) {
	bc := NewBlockchain()
	
	// 生成密钥对
	priv, pub := NewKeyPair()
	
	// 创建并添加多个交易
	var transactions []Transaction
	for i := 0; i < 3; i++ {
		tx := Transaction{
			From:   pub,
			To:     "receiver",
			Amount: 100 + i,
		}
		
		signature, err := SignTransaction(priv, tx)
		if err != nil {
			t.Fatalf("Failed to sign transaction: %v", err)
		}
		tx.Signature = signature
		
		bc.AddTransaction(tx)
		transactions = append(transactions, tx)
	}
	
	// 验证初始状态
	txs := bc.GetTransactions()
	if len(txs) != 3 {
		t.Errorf("Expected 3 transactions in pool, got %d", len(txs))
	}
	
	// 清除部分交易（只清除前两个）
	bc.ClearTransactions(transactions[:2])
	
	// 验证剩余交易
	txs = bc.GetTransactions()
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction remaining, got %d", len(txs))
	}
	
	// 验证保留的是正确的交易
	if txs[0].Amount != 102 {
		t.Errorf("Expected amount 102 in remaining transaction, got %d", txs[0].Amount)
	}
}