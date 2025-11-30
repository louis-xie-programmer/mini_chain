package blockchain

// internal/blockchain/blockchain.go
// High level Blockchain object that uses the above pieces.
// For the PoC this object manages in-memory latest height and basic
// storage operations which will be later hooked into persistent DB.

import (
	"errors"
	"sync"
)

type Blockchain struct {
	lock       sync.RWMutex
	difficulty int
	// Note: block storage is expected to be handled by storage module.
	// Here we keep a small cache of the latest block in memory for fast mining.
	latest Block
}

// NewBlockchain creates a blockchain instance and initializes with genesis.
// difficulty: PoW difficulty (leading hex zeros)
func NewBlockchain(difficulty int) *Blockchain {
	gen := NewGenesis()
	bc := &Blockchain{
		difficulty: difficulty,
		latest:     gen,
	}
	// Note: storage persistence to be done by storage module (callers)
	return bc
}

// GetLatest returns the latest cached block.
func (bc *Blockchain) GetLatest() Block {
	bc.lock.RLock()
	defer bc.lock.RUnlock()
	return bc.latest
}

// SetLatest updates latest (call after successfully persisting a new block).
func (bc *Blockchain) SetLatest(b Block) {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	bc.latest = b
}

// ValidateAndApplyBlock performs validation (PoW + prev-hash) and
// applies transactions to UTXO set (if valid). It expects the caller
// to persist the block (DB) before or after calling depending on design.
func (bc *Blockchain) ValidateAndApplyBlock(b Block) error {
	// 1. basic header hash check
	if !b.ValidateBasic() {
		return errors.New("block header invalid")
	}
	// 2. PoW
	if !CheckPoW(&b, bc.difficulty) {
		return errors.New("block PoW invalid")
	}
	// 3. prev linkage
	latest := bc.GetLatest()
	if b.PrevHash != latest.Hash {
		return errors.New("block does not extend latest")
	}
	// 4. validate contained txs (validateRawTx ensures inputs exist)
	for _, txid := range b.Transactions {
		if err := validateRawTx(txid); err != nil {
			return err
		}
	}
	// 5. apply UTXO changes
	if err := applyTxsInBlock(b.Transactions); err != nil {
		return err
	}
	// 6. update latest
	bc.SetLatest(b)
	// 7. remove from mempool
	RemoveFromMempool(b.Transactions)
	return nil
}

// Simple helper to mine a block given mempool txids:
// - collects mempool
// - runs PoW
// - returns mined block (caller should store and call ValidateAndApplyBlock)
func (bc *Blockchain) MinePending() (Block, error) {
	txids := ListMempool()
	if len(txids) == 0 {
		return Block{}, errors.New("no txs to mine")
	}
	prev := bc.GetLatest()
	b := MineBlock(prev, txids, bc.difficulty)
	// Caller: persist b and then call ValidateAndApplyBlock to commit UTXO changes.
	return b, nil
}
