package blockchain

// internal/blockchain/transaction.go
// UTXO模型中的交易/原始交易表示
// 注意：签名/验证在外部完成（钱包模块）
// 该文件定义了结构体和辅助函数：交易ID计算和基本检查

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// TxInput 交易输入，通过txid:vout引用前一个UTXO，并携带签名和公钥
type TxInput struct {
	Txid      string `json:"txid"`      // 引用的前一个交易ID
	Vout      int    `json:"vout"`      // 引用的前一个交易的输出索引
	Signature string `json:"signature"` // 十六进制编码的签名
	PubKey    string `json:"pubkey"`    // 十六进制编码的公钥（地址）
}

// TxOutput 交易输出（UTXO候选项）
type TxOutput struct {
	Address string `json:"address"` // 接收地址
	Amount  int    `json:"amount"`  // 金额
}

// UTXOTx UTXO模型中的可序列化原始交易
type UTXOTx struct {
	Inputs  []TxInput  `json:"inputs"`  // 交易输入列表
	Outputs []TxOutput `json:"outputs"` // 交易输出列表
}

// TxID 返回确定性的交易ID：sha256(json(rawtx))
// 通过对交易内容进行哈希来生成唯一标识
func TxID(raw UTXOTx) (string, error) {
	b, err := json.Marshal(raw) // 序列化交易为JSON
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b) // 计算SHA256哈希
	return hex.EncodeToString(sum[:]), nil // 返回十六进制编码的哈希值
}

// ValidateTxStructure 基本健全性检查（结构）
// 验证交易的基本结构是否合法
func ValidateTxStructure(raw UTXOTx) error {
	// 检查交易是否既没有输入也没有输出
	if len(raw.Inputs) == 0 && len(raw.Outputs) == 0 {
		return fmt.Errorf("tx has no inputs and no outputs")
	}
	
	// 检查输出金额是否为负数
	for _, out := range raw.Outputs {
		if out.Amount < 0 {
			return fmt.Errorf("negative amount")
		}
	}
	return nil
}

// validateRawTx 检查交易的所有输入是否存在于UTXO集合中
// 区块链使用此函数验证区块中的交易
func validateRawTx(txid string) error {
	// TODO: 这是一个简化的实现
	// 在实际实现中，我们会：
	// 1. 通过txid从存储中检索交易
	// 2. 验证交易结构
	// 3. 检查所有输入UTXO是否存在
	// 4. 验证签名
	// 
	// 目前，我们只返回nil表示交易有效
	
	// 实际实现应该如下所示：
	/*
	// 1. 通过txid从存储中检索交易
	tx, err := getTransactionById(txid)
	if err != nil {
		return fmt.Errorf("transaction not found: %s", txid)
	}
	
	// 2. 验证交易结构
	if err := ValidateTxStructure(tx); err != nil {
		return fmt.Errorf("invalid transaction structure: %v", err)
	}
	
	// 3. 检查所有输入UTXO是否存在
	for _, input := range tx.Inputs {
		_, err := GetUTXO(input.Txid, input.Vout)
		if err != nil {
			return fmt.Errorf("input UTXO not found: %s:%d", input.Txid, input.Vout)
		}
	}
	
	// 4. 验证签名
	if err := verifyTransactionSignatures(tx); err != nil {
		return fmt.Errorf("invalid signatures: %v", err)
	}
	*/
	
	return nil
}