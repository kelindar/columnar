// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"fmt"
	"math"
	"sync"

	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
	"github.com/kelindar/intmap"
	"github.com/zeebo/xxh3"
)

// --------------------------- Enum ----------------------------

var _ Textual = new(columnEnum)

// columnEnum represents a string column
type columnEnum struct {
	chunks[uint32]
	seek *intmap.Sync // The hash->location table
	data []string     // The string data
}

// makeEnum creates a new column
func makeEnum() Column {
	return &columnEnum{
		chunks: make(chunks[uint32], 0, 4),
		seek:   intmap.NewSync(64, .95),
		data:   make([]string, 0, 64),
	}
}

// Apply applies a set of operations to the column.
func (c *columnEnum) Apply(chunk commit.Chunk, r *commit.Reader) {
	fill, locs := c.chunkAt(chunk)
	for r.Next() {
		offset := r.IndexAtChunk()
		switch r.Type {
		case commit.Put:
			fill[offset>>6] |= 1 << (offset & 0x3f)
			locs[offset] = c.findOrAdd(r.Bytes())
		case commit.Delete:
			fill.Remove(offset)
			// TODO: remove unused strings, need some reference counting for that
			// and can proably be done during vacuum() instead
		}
	}
}

// Search for the string or adds it and returns the offset
func (c *columnEnum) findOrAdd(v []byte) uint32 {
	target := uint32(xxh3.Hash(v))
	at, _ := c.seek.LoadOrStore(target, func() uint32 {
		c.data = append(c.data, string(v))
		return uint32(len(c.data)) - 1
	})
	return at
}

// readAt reads a string at a location
func (c *columnEnum) readAt(at uint32) string {
	return c.data[at]
}

// Value retrieves a value at a specified index
func (c *columnEnum) Value(idx uint32) (v interface{}, ok bool) {
	return c.LoadString(idx)
}

// LoadString retrieves a value at a specified index
func (c *columnEnum) LoadString(idx uint32) (v string, ok bool) {
	chunk := commit.ChunkAt(idx)
	index := idx - chunk.Min()
	if int(chunk) < len(c.chunks) && c.chunks[chunk].fill.Contains(index) {
		v, ok = c.readAt(c.chunks[chunk].data[index]), true
	}
	return
}

// FilterString filters down the values based on the specified predicate. The column for
// this filter must be a string.
func (c *columnEnum) FilterString(chunk commit.Chunk, index bitmap.Bitmap, predicate func(v string) bool) {
	if int(chunk) >= len(c.chunks) {
		return
	}

	fill, locs := c.chunkAt(chunk)
	cache := struct {
		index uint32 // Last seen offset
		value bool   // Last evaluated predicate
	}{
		index: math.MaxUint32,
		value: false,
	}

	// Do a quick ellimination of elements which are NOT contained in this column, this
	// allows us not to check contains during the filter itself
	index.And(fill)

	// Filters down the strings, if strings repeat we avoid reading every time by
	// caching the last seen index/value combination.
	index.Filter(func(idx uint32) bool {
		if at := locs[idx]; at != cache.index {
			cache.index = at
			cache.value = predicate(c.readAt(at))
			return cache.value
		}

		// The value is cached, avoid evaluating it
		return cache.value
	})
}

// Contains checks whether the column has a value at a specified index.
func (c *columnEnum) Contains(idx uint32) bool {
	chunk := commit.ChunkAt(idx)
	return c.chunks[chunk].fill.Contains(idx - chunk.Min())
}

// Snapshot writes the entire column into the specified destination buffer
func (c *columnEnum) Snapshot(chunk commit.Chunk, dst *commit.Buffer) {
	fill, locs := c.chunkAt(chunk)
	fill.Range(func(idx uint32) {
		dst.PutString(commit.Put, idx, c.readAt(locs[idx]))
	})
}

// rwEnum represents read-write accessor for enum
type rwEnum struct {
	rdString[*columnEnum]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwEnum) Set(value string) {
	s.writer.PutString(commit.Put, *s.cursor, value)
}

// Enum returns a enumerable column accessor
func (txn *Txn) Enum(columnName string) rwEnum {
	return rwEnum{
		rdString: readStringOf[*columnEnum](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- String ----------------------------

var _ Textual = new(columnString)

// columnString represents a string column
type columnString struct {
	chunks[string]
	option[string]
}

// makeString creates a new string column
func makeStrings(opts ...func(*option[string])) Column {
	return &columnString{
		chunks: make(chunks[string], 0, 4),
		option: configure(opts, option[string]{
			Merge: func(_, delta string) string { return delta },
		}),
	}
}

// Apply applies a set of operations to the column.
func (c *columnString) Apply(chunk commit.Chunk, r *commit.Reader) {
	fill, data := c.chunkAt(chunk)
	from := chunk.Min()

	// Update the values of the column, for this one we can only process stores
	for r.Next() {
		offset := r.Offset - int32(from)
		switch r.Type {
		case commit.Put:
			fill[offset>>6] |= 1 << (offset & 0x3f)
			data[offset] = string(r.Bytes())
		case commit.Merge:
			fill[offset>>6] |= 1 << (offset & 0x3f)
			data[offset] = r.SwapString(c.Merge(data[offset], r.String()))
		case commit.Delete:
			fill.Remove(uint32(offset))
		}
	}
}

// Value retrieves a value at a specified index
func (c *columnString) Value(idx uint32) (v interface{}, ok bool) {
	return c.LoadString(idx)
}

// Contains checks whether the column has a value at a specified index.
func (c *columnString) Contains(idx uint32) bool {
	chunk := commit.ChunkAt(idx)
	index := idx - chunk.Min()
	return c.chunks[chunk].fill.Contains(index)
}

// LoadString retrieves a value at a specified index
func (c *columnString) LoadString(idx uint32) (v string, ok bool) {
	chunk := commit.ChunkAt(idx)
	index := idx - chunk.Min()

	if int(chunk) < len(c.chunks) && c.chunks[chunk].fill.Contains(index) {
		v, ok = c.chunks[chunk].data[index], true
	}
	return
}

// FilterString filters down the values based on the specified predicate. The column for
// this filter must be a string.
func (c *columnString) FilterString(chunk commit.Chunk, index bitmap.Bitmap, predicate func(v string) bool) {
	if int(chunk) < len(c.chunks) {
		fill, data := c.chunkAt(chunk)
		index.And(fill)
		index.Filter(func(idx uint32) bool {
			return predicate(data[idx])
		})
	}
}

// Snapshot writes the entire column into the specified destination buffer
func (c *columnString) Snapshot(chunk commit.Chunk, dst *commit.Buffer) {
	fill, data := c.chunkAt(chunk)
	fill.Range(func(x uint32) {
		dst.PutString(commit.Put, chunk.Min()+x, data[x])
	})
}

// rwString represents read-write accessor for strings
type rwString struct {
	rdString[*columnString]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwString) Set(value string) {
	s.writer.PutString(commit.Put, *s.cursor, value)
}

// Merge merges the value at the current transaction cursor
func (s rwString) Merge(value string) {
	s.writer.PutString(commit.Merge, *s.cursor, value)
}

// String returns a string column accessor
func (txn *Txn) String(columnName string) rwString {
	return rwString{
		rdString: readStringOf[*columnString](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Key ----------------------------

// columnKey represents the primary key column implementation
type columnKey struct {
	columnString
	name string            // Name of the column
	lock sync.RWMutex      // Lock to protect the lookup table
	seek map[string]uint32 // Lookup table for O(1) index seek
}

// makeKey creates a new primary key column
func makeKey() Column {
	return &columnKey{
		seek: make(map[string]uint32, 64),
		columnString: columnString{
			chunks: make(chunks[string], 0, 4),
		},
	}
}

// Apply applies a set of operations to the column.
func (c *columnKey) Apply(chunk commit.Chunk, r *commit.Reader) {
	fill, data := c.chunkAt(chunk)
	from := chunk.Min()

	for r.Next() {
		offset := r.Offset - int32(from)
		switch r.Type {
		case commit.Put:
			value := string(r.Bytes())

			fill[offset>>6] |= 1 << (offset & 0x3f)
			data[offset] = value
			c.lock.Lock()
			c.seek[value] = uint32(r.Offset)
			c.lock.Unlock()

		case commit.Delete:
			fill.Remove(uint32(offset))
			c.lock.Lock()
			delete(c.seek, string(data[offset]))
			c.lock.Unlock()
		}
	}
}

// OffsetOf returns the offset for a particular value
func (c *columnKey) OffsetOf(v string) (uint32, bool) {
	c.lock.RLock()
	idx, ok := c.seek[v]
	c.lock.RUnlock()
	return idx, ok
}

// rwKey represents read-write accessor for primary keys.
type rwKey struct {
	cursor *uint32
	writer *commit.Buffer
	reader *columnKey
}

// Set sets the value at the current transaction index
func (s rwKey) Set(value string) error {
	if _, ok := s.reader.OffsetOf(value); !ok {
		s.writer.PutString(commit.Put, *s.cursor, value)
		return nil
	}

	return fmt.Errorf("column: unable to set duplicate key '%s'", value)
}

// Get loads the value at the current transaction index
func (s rwKey) Get() (string, bool) {
	return s.reader.LoadString(*s.cursor)
}

// Enum returns a enumerable column accessor
func (txn *Txn) Key() rwKey {
	if txn.owner.pk == nil {
		panic(fmt.Errorf("column: primary key column does not exist"))
	}

	return rwKey{
		cursor: &txn.cursor,
		writer: txn.bufferFor(txn.owner.pk.name),
		reader: txn.owner.pk,
	}
}

// --------------------------- Reader ----------------------------

// rdString represents a read-only accessor for strings
type rdString[T Textual] reader[T]

// Get loads the value at the current transaction cursor
func (s rdString[T]) Get() (string, bool) {
	return s.reader.LoadString(*s.cursor)
}

// readStringOf creates a new string reader
func readStringOf[T Textual](txn *Txn, columnName string) rdString[T] {
	return rdString[T](readerFor[T](txn, columnName))
}
