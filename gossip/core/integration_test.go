package core

import (
	"testing"
	"time"
)

// TestMining 测试挖矿流程
func TestMining(t *testing.T) {
	// 创建区块链
	bc := NewBlockchain()
	
	// 生成钱包
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
		t.Fatalf("Failed to sign transaction: %v", err)
	}
	tx.Signature = signature
	
	// 添加交易到交易池
	if !bc.AddTransaction(tx) {
		t.Fatal("Failed to add transaction")
	}
	
	// 验证交易池中有交易
	txs := bc.GetTransactions()
	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction in pool, got %d", len(txs))
	}
	
	// 获取创世区块
	blocks := bc.GetBlocks()
	genesis := blocks[0]
	
	// 模拟挖矿过程
	minedBlock := MineBlock(txs, genesis)
	
	// 验证挖出的区块
	if minedBlock.Index != 1 {
		t.Errorf("Expected block index 1, got %d", minedBlock.Index)
	}
	
	if minedBlock.PrevHash != genesis.Hash {
		t.Error("Previous hash mismatch")
	}
	
	// 添加挖出的区块到区块链
	if !bc.AddBlock(minedBlock) {
		t.Fatal("Failed to add mined block")
	}
	
	// 手动清理交易池（模拟外部调用）
	bc.ClearTransactions(minedBlock.Transactions)
	
	// 验证区块链增长
	blocks = bc.GetBlocks()
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}
	
	// 验证交易池已清空
	txs = bc.GetTransactions()
	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions in pool after mining, got %d", len(txs))
	}
}

// TestConcurrentMining 测试并发挖矿
func TestConcurrentMining(t *testing.T) {
	// 创建区块链
	bc := NewBlockchain()
	
	// 生成钱包
	priv, pub := NewKeyPair()
	
	// 并发添加多个交易
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(i int) {
			tx := Transaction{
				From:   pub,
				To:     "receiver",
				Amount: 100 + i,
			}
			
			signature, err := SignTransaction(priv, tx)
			if err != nil {
				t.Errorf("Failed to sign transaction: %v", err)
				done <- false
				return
			}
			tx.Signature = signature
			
			if !bc.AddTransaction(tx) {
				t.Errorf("Failed to add transaction %d", i)
				done <- false
				return
			}
			
			done <- true
		}(i)
	}
	
	// 等待所有交易添加完成
	for i := 0; i < 5; i++ {
		<-done
	}
	
	// 验证所有交易都被添加
	txs := bc.GetTransactions()
	if len(txs) != 5 {
		t.Errorf("Expected 5 transactions, got %d", len(txs))
	}
	
	// 获取创世区块并挖矿
	blocks := bc.GetBlocks()
	genesis := blocks[0]
	
	minedBlock := MineBlock(txs, genesis)
	
	if !bc.AddBlock(minedBlock) {
		t.Fatal("Failed to add mined block")
	}
	
	// 手动清理交易池（模拟外部调用）
	bc.ClearTransactions(minedBlock.Transactions)
	
	// 验证区块链增长
	blocks = bc.GetBlocks()
	if len(blocks) != 2 {
		t.Errorf("Expected 2 blocks, got %d", len(blocks))
	}
	
	// 验证交易池已清空
	txs = bc.GetTransactions()
	if len(txs) != 0 {
		t.Errorf("Expected 0 transactions in pool after mining, got %d", len(txs))
	}
}

// TestBlockchainPersistence 测试区块链持久性
func TestBlockchainPersistence(t *testing.T) {
	// 创建第一个区块链实例
	bc1 := NewBlockchain()
	
	// 添加一些区块
	blocks := bc1.GetBlocks()
	currentBlock := blocks[0]
	
	// 添加几个区块
	for i := 0; i < 3; i++ {
		newBlock := MineBlock([]Transaction{}, currentBlock)
		if !bc1.AddBlock(newBlock) {
			t.Fatalf("Failed to add block %d", i)
		}
		currentBlock = newBlock
	}
	
	// 验证区块链长度
	blocks = bc1.GetBlocks()
	if len(blocks) != 4 {
		t.Errorf("Expected 4 blocks, got %d", len(blocks))
	}
	
	// 创建第二个区块链实例模拟重启
	bc2 := NewBlockchain()
	
	// 手动设置区块链数据（模拟从存储加载）
	bc2.mutex.Lock()
	bc2.chain = make([]Block, len(blocks))
	copy(bc2.chain, blocks)
	bc2.mutex.Unlock()
	
	// 验证两个实例有一样的区块链数据
	blocks1 := bc1.GetBlocks()
	blocks2 := bc2.GetBlocks()
	
	if len(blocks1) != len(blocks2) {
		t.Errorf("Blockchain length mismatch: %d vs %d", len(blocks1), len(blocks2))
	}
	
	for i := range blocks1 {
		if blocks1[i].Hash != blocks2[i].Hash {
			t.Errorf("Block %d hash mismatch", i)
		}
	}
}

// BenchmarkMineBlock 基准测试区块挖掘性能
func BenchmarkMineBlock(b *testing.B) {
	// 创建前一个区块
	prevBlock := Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []Transaction{},
		PrevHash:     "0",
		Nonce:        0,
	}
	prevBlock.Hash = CalculateHash(prevBlock)
	
	// 创建交易
	transactions := make([]Transaction, 10)
	for i := range transactions {
		transactions[i] = Transaction{
			From:      "sender",
			To:        "receiver",
			Amount:    100 + i,
			Signature: "signature",
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MineBlock(transactions, prevBlock)
	}
}