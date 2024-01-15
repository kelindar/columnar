// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package commit

import (
	"io"
	"sync/atomic"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/kelindar/iostream"
)

// --------------------------- ID ----------------------------

var id uint64 = uint64(time.Now().UnixNano())

// Next returns the next commit ID
func Next() uint64 {
	return atomic.AddUint64(&id, 1)
}

// --------------------------- Chunk ----------------------------

const (
	bitmapShift = chunkShift - 6
	bitmapSize  = 1 << bitmapShift
	chunkShift  = 14 // 16K
	chunkSize   = 1 << chunkShift
)

// Chunk represents a chunk number
type Chunk uint32

// ChunkAt returns the chunk number at a given index
func ChunkAt(index uint32) Chunk {
	return Chunk(index >> chunkShift)
}

// OfBitmap computes a chunk for a given bitmap
func (c Chunk) OfBitmap(v bitmap.Bitmap) bitmap.Bitmap {
	const shift = chunkShift - 6
	x1 := min(int32(c+1)<<shift, int32(len(v)))
	x0 := min(int32(c)<<shift, x1)
	return v[x0:x1]
}

// Min returns the min offset at which the chunk should be starting
func (c Chunk) Min() uint32 {
	return uint32(int32(c) << chunkShift)
}

// Max returns the max offset at which the chunk should be ending
func (c Chunk) Max() uint32 {
	return c.Min() + chunkSize - 1
}

// Range iterates over a chunk given a bitmap
func (c Chunk) Range(v bitmap.Bitmap, fn func(idx uint32)) {
	offset := c.Min()
	output := c.OfBitmap(v)
	output.Range(func(idx uint32) {
		fn(offset + idx)
	})
}

// min returns a minimum of two numbers without branches.
func min(v1, v2 int32) int32 {
	return v2 + ((v1 - v2) & ((v1 - v2) >> 31))
}

// --------------------------- Commit ----------------------------

// Commit represents an individual transaction commit. If multiple chunks are committed
// in the same transaction, it would result in multiple commits per transaction.
type Commit struct {
	ID      uint64    // The commit ID
	Chunk   Chunk     // The chunk number
	Updates []*Buffer // The update buffers
}

// Clone clones a commit into a new one
func (c *Commit) Clone() (clone Commit) {
	clone.Chunk = c.Chunk
	for _, u := range c.Updates {
		if len(u.buffer) > 0 {
			clone.Updates = append(clone.Updates, u.Clone())
		}
	}
	return
}

// WriteTo writes data to w until there's no more data to write or when an error occurs. The return
// value n is the number of bytes written. Any error encountered during the write is also returned.
func (c *Commit) WriteTo(dst io.Writer) (int64, error) {
	w := iostream.NewWriter(dst)

	// Write the chunk ID
	if err := w.WriteUvarint(uint64(c.Chunk)); err != nil {
		return w.Offset(), err
	}

	// Write the commit ID
	if err := w.WriteUvarint(c.ID); err != nil {
		return w.Offset(), err
	}

	// Write all of the columns for the current chunk
	reader := NewReader()
	if err := w.WriteRange(len(c.Updates), func(i int, w *iostream.Writer) error {
		buffer := c.Updates[i]

		// Write the column name for this buffer
		if err := w.WriteString(buffer.Column); err != nil {
			return err
		}

		// Write the number of shards in case of interleaved buffer
		shards := uint64(0)
		reader.Range(buffer, c.Chunk, func(r *Reader) {
			shards++
		})
		if err := w.WriteUvarint(shards); err != nil {
			return err
		}

		// Write chunk information
		offset := uint32(0)
		reader.Range(buffer, c.Chunk, func(r *Reader) {
			_ = w.WriteUint32(uint32(r.Offset)) // Value
			_ = w.WriteUint32(offset)           // Offset
			offset += uint32(len(r.buffer))
		})

		// Write buffer length
		if err := w.WriteUvarint(uint64(offset)); err != nil {
			return err
		}

		// Write all chunk bytes together
		reader.Range(buffer, c.Chunk, func(r *Reader) {
			_, _ = w.Write(r.buffer)
		})
		return nil
	}); err != nil {
		return w.Offset(), err
	}

	return w.Offset(), nil
}

// ReadFrom reads data from r until EOF or error. The return value n is the number of
// bytes read. Any error except EOF encountered during the read is also returned.
func (c *Commit) ReadFrom(src io.Reader) (int64, error) {
	r := iostream.NewReader(src)

	// Read chunk ID
	chunk, err := r.ReadUvarint()
	c.Chunk = Chunk(chunk)
	if err != nil {
		return r.Offset(), err
	}

	// Read commit ID
	if c.ID, err = r.ReadUvarint(); err != nil {
		return r.Offset(), err
	}

	// Read each update buffer in the commit
	if err := r.ReadRange(func(i int, r *iostream.Reader) error {
		buffer := NewBuffer(256)
		c.Updates = append(c.Updates, buffer)

		// Read the column name
		column, err := r.ReadString()
		if err != nil {
			return err
		}

		// Read the chunks array
		buffer.Reset(column)
		r.ReadRange(func(i int, r *iostream.Reader) error {
			header := header{
				Chunk: Chunk(chunk),
			}

			// Previous offset and index in the byte array
			if header.Value, err = r.ReadUint32(); err != nil {
				return err
			}
			if header.Start, err = r.ReadUint32(); err != nil {
				return err
			}

			buffer.chunks = append(buffer.chunks, header)
			return nil
		})

		// Read the combined buffer
		buffer.buffer, err = r.ReadBytes()
		return err
	}); err != nil {
		return r.Offset(), err
	}

	return r.Offset(), nil
}
