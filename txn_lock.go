// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
)

const (
	bitmapShift = chunkShift - 6
	bitmapSize  = 1 << bitmapShift
	chunkShift  = 14 // 16K
	chunkSize   = 1 << chunkShift
)

// initialize ensures that the transaction is pre-initialized with the snapshot
// of the owner's fill list.
func (txn *Txn) initialize() {
	if txn.setup {
		return
	}

	txn.owner.lock.RLock()
	txn.index.Grow(uint32(txn.owner.opts.Capacity))
	txn.owner.fill.Clone(&txn.index)
	txn.owner.lock.RUnlock()
	txn.setup = true
}

// --------------------------- Locked Seek ---------------------------

// QueryAt jumps at a particular offset in the collection, sets the cursor to the
// provided position and executes given callback fn.
func (txn *Txn) QueryAt(index uint32, f func(Row) error) (err error) {
	lock := txn.owner.slock
	txn.cursor = index

	chunk := commit.ChunkAt(index)
	lock.RLock(uint(chunk))
	err = f(Row{txn})
	lock.RUnlock(uint(chunk))
	return err
}

// --------------------------- Locked Range ---------------------------

// rangeRead iterates over index, chunk by chunk and ensures that each
// chunk is protected by an appropriate read lock.
func (txn *Txn) rangeRead(f func(chunk commit.Chunk, index bitmap.Bitmap)) {
	limit := commit.Chunk(len(txn.index) >> bitmapShift)
	lock := txn.owner.slock

	for chunk := commit.Chunk(0); chunk <= limit; chunk++ {
		lock.RLock(uint(chunk))
		f(chunk, chunk.OfBitmap(txn.index))
		lock.RUnlock(uint(chunk))
	}
}

// rangeReadPair iterates over the index and another bitmap, chunk by chunk and
// ensures that each chunk is protected by an appropriate read lock.
func (txn *Txn) rangeReadPair(column *column, f func(a, b bitmap.Bitmap)) {
	limit := commit.Chunk(len(txn.index) >> bitmapShift)
	lock := txn.owner.slock

	// Iterate through all of the chunks and acquire appropriate shard locks.
	for chunk := commit.Chunk(0); chunk <= limit; chunk++ {
		lock.RLock(uint(chunk))
		f(chunk.OfBitmap(txn.index), column.Index(chunk))
		lock.RUnlock(uint(chunk))
	}
}

// rangeWrite ranges over the dirty chunks and acquires exclusive latches along
// the way. This is used to commit a transaction.
func (txn *Txn) rangeWrite(fn func(commitID uint64, chunk commit.Chunk, fill bitmap.Bitmap)) {
	lock := txn.owner.slock
	txn.dirty.Range(func(x uint32) {
		chunk := commit.Chunk(x)
		commitID := commit.Next()
		lock.Lock(uint(chunk))

		// Compute the fill and set the last commit ID
		txn.owner.lock.RLock()
		fill := chunk.OfBitmap(txn.owner.fill)
		txn.owner.commits[chunk] = commitID // OK, since we have a shard lock
		txn.owner.lock.RUnlock()

		// Call the delegate
		fn(commitID, chunk, fill)
		lock.Unlock(uint(chunk))
	})
}
