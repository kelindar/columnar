// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"unsafe"

	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
	"github.com/kelindar/iostream"
	"github.com/klauspost/compress/s2"
)

var (
	errUnexpectedEOF = errors.New("column: unable to restore, unexpected EOF")
)

// --------------------------- Commit Replay ---------------------------

// Replay replays a commit on a collection, applying the changes.
func (c *Collection) Replay(change commit.Commit) error {
	return c.Query(func(txn *Txn) error {
		txn.dirty.Set(uint32(change.Chunk))
		for i := range change.Updates {
			if !change.Updates[i].IsEmpty() {
				txn.updates = append(txn.updates, change.Updates[i])
			}
		}
		return nil
	})
}

// --------------------------- Snapshotting ---------------------------

// Restore restores the collection from the underlying snapshot reader. This operation
// should be called before any of transactions, right after initialization.
func (c *Collection) Restore(snapshot io.Reader) error {
	commits, err := c.readState(s2.NewReader(snapshot))
	if err != nil {
		return err
	}

	// Reconcile the pending commit log
	return commit.Open(snapshot).Range(func(commit commit.Commit) error {
		lastCommit := commits[commit.Chunk]
		if commit.ID > lastCommit {
			return c.Replay(commit)
		}
		return nil
	})
}

// Snapshot writes a collection snapshot into the underlying writer.
func (c *Collection) Snapshot(dst io.Writer) error {
	recorder, err := c.recorderOpen()
	if err != nil {
		return err
	}

	// Take a snapshot of the current state
	defer os.Remove(recorder.Name())
	if _, err := c.writeState(s2.NewWriter(dst)); err != nil {
		return err
	}

	// Close the recorder
	c.recorderClose()
	return recorder.Copy(dst)
}

// recorderOpen opens a recorder for commits while the snapshot is in progress
func (c *Collection) recorderOpen() (log *commit.Log, err error) {
	if log, err = commit.OpenTemp(); err == nil {
		dst := (*unsafe.Pointer)(unsafe.Pointer(&c.record))
		ptr := unsafe.Pointer(log)
		if !atomic.CompareAndSwapPointer(dst, nil, ptr) {
			return nil, fmt.Errorf("column: unable to snapshot, another one might be in progress")
		}
	}
	return
}

// recorderClose closes the pending commit recorder and deletes the file
func (c *Collection) recorderClose() {
	if _, ok := c.isSnapshotting(); ok {
		dst := (*unsafe.Pointer)(unsafe.Pointer(&c.record))
		atomic.StorePointer(dst, nil)
	}
}

// isSnapshotting loads a currently used commit log for a pending snapshot
func (c *Collection) isSnapshotting() (*commit.Log, bool) {
	dst := (*unsafe.Pointer)(unsafe.Pointer(&c.record))
	ptr := atomic.LoadPointer(dst)
	if ptr == nil {
		return nil, false
	}

	return (*commit.Log)(ptr), true
}

// --------------------------- Collection Encoding ---------------------------

// writeState writes collection state into the specified writer.
func (c *Collection) writeState(dst io.Writer) (int64, error) {
	writer := iostream.NewWriter(dst)
	buffer := c.txns.acquirePage(rowColumn)
	defer c.txns.releasePage(buffer)

	// Write the schema version
	if err := writer.WriteUvarint(0x1); err != nil {
		return writer.Offset(), err
	}

	// Load the number of columns and the max index
	chunks := c.chunks()
	columns := uint64(c.cols.Count()) + 1 // extra 'insert' column

	// Write the number of columns
	if err := writer.WriteUvarint(columns); err != nil {
		return writer.Offset(), err
	}

	// Write each chunk
	if err := writer.WriteRange(chunks, func(i int, w *iostream.Writer) error {
		return c.readChunk(commit.Chunk(i), func(lastCommit uint64, chunk commit.Chunk, fill bitmap.Bitmap) error {
			offset := chunk.Min()

			// Write the last written commit for this chunk
			if err := writer.WriteUvarint(lastCommit); err != nil {
				return err
			}

			// Write the inserts column
			buffer.Reset(rowColumn)
			fill.Range(func(idx uint32) {
				buffer.PutOperation(commit.Insert, offset+idx)
			})
			if err := writer.WriteSelf(buffer); err != nil {
				return err
			}

			// Snapshot each column and write the buffer
			return c.cols.RangeUntil(func(column *column) error {
				if !column.Snapshot(chunk, buffer) {
					return nil // Skip indexes
				}
				return writer.WriteSelf(buffer)
			})
		})
	}); err != nil {
		return writer.Offset(), err
	}

	return writer.Offset(), writer.Flush()
}

// readState reads a collection snapshotted state from the underlying reader. It
// returns the last commit IDs for each chunk.
func (c *Collection) readState(src io.Reader) (map[commit.Chunk]uint64, error) {
	r := iostream.NewReader(src)
	commits := make(map[commit.Chunk]uint64)

	// Read the version and make sure it matches
	version, err := r.ReadUvarint()
	if err != nil || version != 0x1 {
		return nil, fmt.Errorf("column: unable to restore (version %d) %v", version, err)
	}

	// Read the number of columns
	columns, err := r.ReadUvarint()
	if err != nil {
		return nil, err
	}

	// Read each chunk
	return commits, r.ReadRange(func(chunk int, r *iostream.Reader) error {
		return c.Query(func(txn *Txn) error {
			txn.dirty.Set(uint32(chunk))

			// Read the last written commit ID for the chunk
			if commits[commit.Chunk(chunk)], err = r.ReadUvarint(); err != nil {
				return err
			}

			for i := uint64(0); i < columns; i++ {
				buffer := txn.owner.txns.acquirePage("")
				_, err := buffer.ReadFrom(r)
				switch {
				case err == io.EOF && i < columns:
					return errUnexpectedEOF
				case err != nil:
					return err
				default:
					txn.updates = append(txn.updates, buffer)
				}
			}

			return nil
		})
	})
}

// chunks returns the number of chunks and columns
func (c *Collection) chunks() int {
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.fill) == 0 {
		return 0
	}

	max, _ := c.fill.Max()
	return int(commit.ChunkAt(max) + 1)
}

// readChunk acquires appropriate locks for a chunk and executes a read callback.
// This is used for snapshotting purposes only.
func (c *Collection) readChunk(chunk commit.Chunk, fn func(uint64, commit.Chunk, bitmap.Bitmap) error) error {

	// Lock both the chunk and the fill list
	c.slock.RLock(uint(chunk))
	c.lock.Lock()
	defer c.slock.RUnlock(uint(chunk))
	defer c.lock.Unlock()
	return fn(c.commits[chunk], chunk, chunk.OfBitmap(c.fill))
}
