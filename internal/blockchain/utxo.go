package blockchain

// internal/blockchain/utxo.go
// In-memory UTXO set implementation for PoC (persistence layer plugs in).
// This module exposes an interface that higher-level storage can implement.
// For simplicity here we use an in-memory map protected by mutex.
// Later when integrating storage, this will be backed by Badger/LevelDB.

import (
	"fmt"
	"sync"
)

type UTXOKey struct {
	Txid string
	Vout int
}

type UTXOEntry struct {
	Address string
	Amount  int
}

var (
	utxoLock sync.RWMutex
	utxos    = make(map[UTXOKey]UTXOEntry) // in-memory set
)

// UTXO management APIs

// GetUTXO returns a UTXO entry or error.
func GetUTXO(txid string, vout int) (UTXOEntry, error) {
	utxoLock.RLock()
	defer utxoLock.RUnlock()
	k := UTXOKey{Txid: txid, Vout: vout}
	e, ok := utxos[k]
	if !ok {
		return UTXOEntry{}, fmt.Errorf("utxo not found %s:%d", txid, vout)
	}
	return e, nil
}

// PutUTXO inserts a new UTXO (called when a block is applied).
func PutUTXO(txid string, vout int, entry UTXOEntry) {
	utxoLock.Lock()
	defer utxoLock.Unlock()
	k := UTXOKey{Txid: txid, Vout: vout}
	utxos[k] = entry
}

// DeleteUTXO removes a UTXO (consumed by a tx input).
func DeleteUTXO(txid string, vout int) {
	utxoLock.Lock()
	defer utxoLock.Unlock()
	k := UTXOKey{Txid: txid, Vout: vout}
	delete(utxos, k)
}

// FindUTXOsForAddress returns all UTXOs owned by address.
func FindUTXOsForAddress(address string) []struct {
	Txid string
	Vout int
	UTXOEntry
} {
	utxoLock.RLock()
	defer utxoLock.RUnlock()
	res := []struct {
		Txid string
		Vout int
		UTXOEntry
	}{}
	for k, v := range utxos {
		if v.Address == address {
			res = append(res, struct {
				Txid string
				Vout int
				UTXOEntry
			}{Txid: k.Txid, Vout: k.Vout, UTXOEntry: v})
		}
	}
	return res
}

// applyTxsInBlock applies the UTXO changes for all transactions in a block.
// For each transaction:
// 1. Remove consumed UTXOs (from inputs)
// 2. Add new UTXOs (from outputs)
func applyTxsInBlock(txids []string) error {
	// TODO: This is a simplified implementation.
	// In a real implementation, we would:
	// 1. For each transaction:
	//    a. Remove all input UTXOs from the set
	//    b. Add all output UTXOs to the set
	// 
	// For now, we'll just return nil to indicate success
	return nil
}