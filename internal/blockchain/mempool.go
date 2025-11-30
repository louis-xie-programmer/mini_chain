package blockchain

// internal/blockchain/mempool.go
// 简单的内存池管理，用于UTXO交易ID（字符串）
// 在生产级节点中，我们会实现优先级、手续费、过期清理等功能

import (
	"sync"
)

var (
	mempoolLock sync.Mutex // 内存池互斥锁，保护并发访问
	mempool     []string   // 内存池，存储待处理的交易ID
)

// AddToMempool 将交易ID添加到内存池（如果不存在）
func AddToMempool(txid string) {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	// 检查交易是否已存在于内存池中
	for _, t := range mempool {
		if t == txid {
			return
		}
	}
	// 添加新交易到内存池
	mempool = append(mempool, txid)
}

// RemoveFromMempool 从内存池中移除已被包含在区块中的交易
func RemoveFromMempool(txids []string) {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	// 创建新的内存池，只保留未被包含在区块中的交易
	newPool := make([]string, 0, len(mempool))
outer:
	// 遍历当前内存池中的所有交易
	for _, e := range mempool {
		// 检查该交易是否在要移除的列表中
		for _, r := range txids {
			if e == r {
				continue outer // 如果在移除列表中，跳过该交易
			}
		}
		// 保留在新内存池中
		newPool = append(newPool, e)
	}
	// 更新内存池
	mempool = newPool
}

// ListMempool 返回当前内存池交易ID的副本
func ListMempool() []string {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	// 创建内存池的副本以避免外部修改
	cp := make([]string, len(mempool))
	copy(cp, mempool)
	return cp
}