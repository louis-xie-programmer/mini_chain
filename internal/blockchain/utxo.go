package blockchain

// internal/blockchain/utxo.go
// 概念验证的内存UTXO集合实现（持久化层可插拔）
// 该模块暴露了高层存储可以实现的接口
// 为了简化，这里使用受互斥锁保护的内存映射
// 后续集成存储时，将由Badger/LevelDB支持

import (
	"fmt"
	"sync"
)

// UTXOKey UTXO键值结构，用于唯一标识一个UTXO
type UTXOKey struct {
	Txid string // 交易ID
	Vout int    // 输出索引
}

// UTXOEntry UTXO条目结构，包含UTXO的具体信息
type UTXOEntry struct {
	Address string // 地址
	Amount  int    // 金额
}

var (
	utxoLock sync.RWMutex        // UTXO读写锁，保护并发访问
	utxos    = make(map[UTXOKey]UTXOEntry) // 内存中的UTXO集合
)

// UTXO管理接口

// GetUTXO 返回指定的UTXO条目或错误
func GetUTXO(txid string, vout int) (UTXOEntry, error) {
	utxoLock.RLock() // 获取读锁
	defer utxoLock.RUnlock()
	
	// 构造UTXO键值
	k := UTXOKey{Txid: txid, Vout: vout}
	
	// 查找UTXO条目
	e, ok := utxos[k]
	if !ok {
		// 未找到对应UTXO
		return UTXOEntry{}, fmt.Errorf("utxo not found %s:%d", txid, vout)
	}
	return e, nil
}

// PutUTXO 插入新的UTXO（在区块被应用时调用）
func PutUTXO(txid string, vout int, entry UTXOEntry) {
	utxoLock.Lock() // 获取写锁
	defer utxoLock.Unlock()
	
	// 构造UTXO键值并插入
	k := UTXOKey{Txid: txid, Vout: vout}
	utxos[k] = entry
}

// DeleteUTXO 删除UTXO（被交易输入消费）
func DeleteUTXO(txid string, vout int) {
	utxoLock.Lock() // 获取写锁
	defer utxoLock.Unlock()
	
	// 构造UTXO键值并删除
	k := UTXOKey{Txid: txid, Vout: vout}
	delete(utxos, k)
}

// FindUTXOsForAddress 返回指定地址拥有的所有UTXO
func FindUTXOsForAddress(address string) []struct {
	Txid string
	Vout int
	UTXOEntry
} {
	utxoLock.RLock() // 获取读锁
	defer utxoLock.RUnlock()
	
	// 初始化结果列表
	res := []struct {
		Txid string
		Vout int
		UTXOEntry
	}{}
	
	// 遍历所有UTXO，筛选指定地址的UTXO
	for k, v := range utxos {
		if v.Address == address {
			// 找到匹配地址的UTXO，添加到结果列表
			res = append(res, struct {
				Txid string
				Vout int
				UTXOEntry
			}{Txid: k.Txid, Vout: k.Vout, UTXOEntry: v})
		}
	}
	return res
}

// applyTxsInBlock 应用区块中所有交易的UTXO变更
// 对于每笔交易：
// 1. 删除被消费的UTXO（来自输入）
// 2. 添加新的UTXO（来自输出）
func applyTxsInBlock(txids []string) error {
	// TODO: 这是一个简化的实现
	// 在实际实现中，我们会：
	// 1. 对于每笔交易：
	//    a. 从集合中删除所有输入UTXO
	//    b. 将所有输出UTXO添加到集合中
	// 
	// 目前，我们只返回nil表示成功
	
	// 实际实现应该如下所示：
	/*
	for _, txid := range txids {
		// 获取交易详情
		tx, err := getTransactionById(txid)
		if err != nil {
			return fmt.Errorf("failed to get transaction %s: %v", txid, err)
		}
		
		// 删除被消费的UTXO（来自输入）
		for _, input := range tx.Inputs {
			DeleteUTXO(input.Txid, input.Vout)
		}
		
		// 添加新的UTXO（来自输出）
		for i, output := range tx.Outputs {
			entry := UTXOEntry{
				Address: output.Address,
				Amount:  output.Amount,
			}
			PutUTXO(txid, i, entry)
		}
	}
	*/
	
	return nil
}