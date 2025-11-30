package core

import (
	"fmt"
)

// Example demonstrates how to use the core blockchain package
func Example() {
	// Create a new blockchain
	bc := NewBlockchain()
	fmt.Printf("Created blockchain with %d blocks\n", len(bc.GetBlocks()))

	// Generate a wallet
	priv, pub := NewKeyPair()
	fmt.Printf("Generated wallet with public key: %s...\n", pub[:2])

	// Create and sign a transaction
	tx := Transaction{
		From:   pub,
		To:     "recipient_address",
		Amount: 50,
	}

	signature, err := SignTransaction(priv, tx)
	if err != nil {
		fmt.Printf("Error signing transaction: %v\n", err)
		return
	}
	tx.Signature = signature

	// Add transaction to the pool
	if bc.AddTransaction(tx) {
		fmt.Printf("Added transaction to pool. Pool size: %d\n", len(bc.GetTransactions()))
	}

	// Mine a block with the transactions
	blocks := bc.GetBlocks()
	lastBlock := blocks[len(blocks)-1]
	
	// Get transactions to mine
	txs := bc.GetTransactions()
	fmt.Printf("Transactions to mine: %d\n", len(txs))
	
	// Mine new block
	newBlock := MineBlock(txs, lastBlock)
	fmt.Printf("Mined new block with %d transactions\n", len(newBlock.Transactions))

	// Add the mined block to the chain
	if bc.AddBlock(newBlock) {
		fmt.Printf("Added block to blockchain. New chain length: %d\n", len(bc.GetBlocks()))
	}

	// Clear transactions from pool
	bc.ClearTransactions(newBlock.Transactions)
	fmt.Printf("Cleared transactions. Pool size: %d\n", len(bc.GetTransactions()))
	// Output:
	// Created blockchain with 1 blocks
	// Generated wallet with public key: 04...
	// Added transaction to pool. Pool size: 1
	// Transactions to mine: 1
	// Mined new block with 1 transactions
	// Added block to blockchain. New chain length: 2
	// Cleared transactions. Pool size: 0
}