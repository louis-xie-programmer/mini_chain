package blockchain

// internal/blockchain/blockchain.go
// 高层区块链对象，使用上述组件构建
// 在这个概念验证实现中，该对象管理内存中的最新区块高度和基本存储操作，
// 后续会接入持久化数据库

import (
	"errors"
	"sync"
)

// Blockchain 区块链结构体，管理区块的验证、存储和同步
type Blockchain struct {
	lock       sync.RWMutex // 读写锁，保护区块链数据的并发访问
	difficulty int          // 工作量证明难度（前导十六进制0的个数）
	// 注意：区块存储预计由存储模块处理
	// 这里我们在内存中缓存最新区块，以便快速挖矿
	latest Block // 最新区块缓存
}

// NewBlockchain 创建区块链实例并用创世区块初始化
// difficulty: PoW难度（前导十六进制0的个数）
func NewBlockchain(difficulty int) *Blockchain {
	gen := NewGenesis() // 创建创世区块
	bc := &Blockchain{
		difficulty: difficulty,
		latest:     gen, // 初始化最新区块为创世区块
	}
	// 注意：存储持久化由存储模块处理（调用者负责）
	return bc
}

// GetLatest 返回缓存的最新区块
// 使用读锁确保并发安全
func (bc *Blockchain) GetLatest() Block {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.latest
}

// SetLatest 更新最新区块（在成功持久化新区块后调用）
func (bc *Blockchain) SetLatest(b Block) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.latest = b
}

// ValidateAndApplyBlock 执行区块验证（PoW + 前一区块哈希链接）并应用交易到UTXO集合
// 该函数期望调用者在调用前后根据设计持久化区块
func (bc *Blockchain) ValidateAndApplyBlock(b Block) error {
	// 1. 基本头部哈希检查
	if !b.ValidateBasic() {
		return errors.New("block header invalid")
	}
	// 2. 工作量证明验证
	if !CheckPoW(&b, bc.difficulty) {
		return errors.New("block PoW invalid")
	}
	// 3. 前一区块链接验证
	latest := bc.GetLatest()
	if b.PrevHash != latest.Hash {
		return errors.New("block does not extend latest")
	}
	// 4. 验证包含的交易（validateRawTx确保输入存在）
	for _, txid := range b.Transactions {
		if err := validateRawTx(txid); err != nil {
			return err
		}
	}
	// 5. 应用UTXO变更
	if err := applyTxsInBlock(b.Transactions); err != nil {
		return err
	}
	// 6. 更新最新区块
	bc.SetLatest(b)
	// 7. 从内存池中移除已打包的交易
	RemoveFromMempool(b.Transactions)
	return nil
}

// MinePending 挖取包含内存池交易的新区块的辅助函数:
// - 收集内存池中的交易
// - 运行工作量证明算法
// - 返回挖取的区块（调用者应存储并调用ValidateAndApplyBlock提交UTXO变更）
func (bc *Blockchain) MinePending(minerAddress string, reward int) (Block, error) {
	txids := ListMempool() // 获取当前内存池中的交易ID列表

	// 创建coinbase交易作为矿工奖励
	coinbaseTx := CoinbaseTx("Mining Reward", minerAddress, reward)
	coinbaseTxId, err := TxID(coinbaseTx)
	if err != nil {
		return Block{}, errors.New("failed to generate coinbase transaction")
	}

	// 将coinbase交易ID添加到交易列表开头
	allTxIds := append([]string{coinbaseTxId}, txids...)

	if len(allTxIds) <= 1 { // 只有coinbase交易
		return Block{}, errors.New("no txs to mine")
	}

	prev := bc.GetLatest()                        // 获取前一个区块
	b := MineBlock(prev, allTxIds, bc.difficulty) // 挖取新区块
	// 调用者：持久化b然后调用ValidateAndApplyBlock提交UTXO变更
	return b, nil
}
