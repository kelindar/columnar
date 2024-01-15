// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
)

// columnBool represents a boolean column
type columnBool struct {
	data bitmap.Bitmap
}

// makeBools creates a new boolean column
func makeBools() Column {
	return &columnBool{
		data: make(bitmap.Bitmap, 0, 4),
	}
}

// Grow grows the size of the column until we have enough to store
func (c *columnBool) Grow(idx uint32) {
	c.data.Grow(idx)
}

// Apply applies a set of operations to the column.
func (c *columnBool) Apply(chunk commit.Chunk, r *commit.Reader) {
	for r.Next() {
		v := uint64(1) << (r.Offset & 0x3f)
		switch r.Type {
		case commit.PutTrue:
			c.data[r.Offset>>6] |= v
		case commit.PutFalse: // also "delete"
			c.data[r.Offset>>6] &^= v
		}
	}
}

// Value retrieves a value at a specified index
func (c *columnBool) Value(idx uint32) (interface{}, bool) {
	value := c.data.Contains(idx)
	return value, value
}

// Contains checks whether the column has a value at a specified index.
func (c *columnBool) Contains(idx uint32) bool {
	return c.data.Contains(idx)
}

// Index returns the fill list for the column
func (c *columnBool) Index(chunk commit.Chunk) bitmap.Bitmap {
	return chunk.OfBitmap(c.data)
}

// Snapshot writes the entire column into the specified destination buffer
func (c *columnBool) Snapshot(chunk commit.Chunk, dst *commit.Buffer) {
	dst.PutBitmap(commit.PutTrue, chunk, c.data)
}

// --------------------------- Writer ----------------------------

// rwBool represents read-write accessor for boolean values
type rwBool struct {
	rdBool
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwBool) Set(value bool) {
	s.writer.PutBool(*s.cursor, value)
}

// Bool returns a bool column accessor
func (txn *Txn) Bool(columnName string) rwBool {
	return rwBool{
		rdBool: readBoolOf(txn, columnName),
		writer: txn.bufferFor(columnName),
	}
}

// --------------------------- Reader ----------------------------

// rdBool represents a read-only accessor for boolean values
type rdBool reader[Column]

// Get loads the value at the current transaction cursor
func (s rdBool) Get() bool {
	return s.reader.Contains(*s.cursor)
}

// readBoolOf creates a new boolean reader
func readBoolOf(txn *Txn, columnName string) rdBool {
	return rdBool(readerFor[Column](txn, columnName))
}
