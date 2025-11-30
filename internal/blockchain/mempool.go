package blockchain

// internal/blockchain/mempool.go
// Simple mempool management for UTXO txids (strings).
// In a production node we'd have prioritization, fees, eviction, etc.

import (
	"sync"
)

var (
	mempoolLock sync.Mutex
	mempool     []string // txids
)

// AddToMempool appends txid if not present.
func AddToMempool(txid string) {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	for _, t := range mempool {
		if t == txid {
			return
		}
	}
	mempool = append(mempool, txid)
}

// RemoveFromMempool removes txids that were included in a block.
func RemoveFromMempool(txids []string) {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	newPool := make([]string, 0, len(mempool))
outer:
	for _, e := range mempool {
		for _, r := range txids {
			if e == r {
				continue outer
			}
		}
		newPool = append(newPool, e)
	}
	mempool = newPool
}

// ListMempool returns a copy of current mempool txids.
func ListMempool() []string {
	mempoolLock.Lock()
	defer mempoolLock.Unlock()
	cp := make([]string, len(mempool))
	copy(cp, mempool)
	return cp
}
