// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"fmt"

	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
	"github.com/kelindar/simd"
)

//go:generate go run ./codegen/main.go

// readNumber is a helper function for point reads
func readNumber[T simd.Number](txn *Txn, columnName string) (value T, found bool) {
	if column, ok := txn.columnAt(columnName); ok {
		if rdr, ok := column.Column.(*numericColumn[T]); ok {
			value, found = rdr.load(txn.cursor)
		}
	}
	return
}

// --------------------------- Generic Column ----------------------------

// numericColumn represents a numeric column
type numericColumn[T simd.Number] struct {
	chunks[T]
	option[T]
	write func(*commit.Buffer, uint32, T)
	apply func(*commit.Reader, bitmap.Bitmap, []T, option[T])
}

// makeNumeric creates a new vector for simd.Numbers
func makeNumeric[T simd.Number](
	write func(*commit.Buffer, uint32, T),
	apply func(*commit.Reader, bitmap.Bitmap, []T, option[T]),
	opts []func(*option[T]),
) *numericColumn[T] {
	return &numericColumn[T]{
		chunks: make(chunks[T], 0, 4),
		write:  write,
		apply:  apply,
		option: configure(opts, option[T]{
			Merge: func(value, delta T) T { return value + delta },
		}),
	}
}

// --------------------------- Accessors ----------------------------

// Contains checks whether the column has a value at a specified index.
func (c *numericColumn[T]) Contains(idx uint32) bool {
	chunk := commit.ChunkAt(idx)
	return c.chunks[chunk].fill.Contains(idx - chunk.Min())
}

// load retrieves a float64 value at a specified index
func (c *numericColumn[T]) load(idx uint32) (v T, ok bool) {
	chunk := commit.ChunkAt(idx)
	index := idx - chunk.Min()
	if int(chunk) < len(c.chunks) && c.chunks[chunk].fill.Contains(index) {
		v, ok = c.chunks[chunk].data[index], true
	}
	return
}

// Value retrieves a value at a specified index
func (c *numericColumn[T]) Value(idx uint32) (any, bool) {
	return c.load(idx)
}

// LoadFloat64 retrieves a float64 value at a specified index
func (c *numericColumn[T]) LoadFloat64(idx uint32) (float64, bool) {
	v, ok := c.load(idx)
	return float64(v), ok
}

// LoadInt64 retrieves an int64 value at a specified index
func (c *numericColumn[T]) LoadInt64(idx uint32) (int64, bool) {
	v, ok := c.load(idx)
	return int64(v), ok
}

// LoadUint64 retrieves an uint64 value at a specified index
func (c *numericColumn[T]) LoadUint64(idx uint32) (uint64, bool) {
	v, ok := c.load(idx)
	return uint64(v), ok
}

// --------------------------- Filtering ----------------------------

// filterNumbers filters down the values based on the specified predicate.
func filterNumbers[T, C simd.Number](column *numericColumn[T], chunk commit.Chunk, index bitmap.Bitmap, predicate func(C) bool) {
	if int(chunk) < len(column.chunks) {
		fill, data := column.chunkAt(chunk)
		index.And(fill)
		index.Filter(func(idx uint32) bool {
			return predicate(C(data[idx]))
		})
	}
}

// FilterFloat64 filters down the values based on the specified predicate.
func (c *numericColumn[T]) FilterFloat64(chunk commit.Chunk, index bitmap.Bitmap, predicate func(float64) bool) {
	filterNumbers(c, chunk, index, predicate)
}

// FilterInt64 filters down the values based on the specified predicate.
func (c *numericColumn[T]) FilterInt64(chunk commit.Chunk, index bitmap.Bitmap, predicate func(int64) bool) {
	filterNumbers(c, chunk, index, predicate)
}

// FilterUint64 filters down the values based on the specified predicate.
func (c *numericColumn[T]) FilterUint64(chunk commit.Chunk, index bitmap.Bitmap, predicate func(uint64) bool) {
	filterNumbers(c, chunk, index, predicate)
}

// --------------------------- Apply & Snapshot ----------------------------

// Apply applies a set of operations to the column.
func (c *numericColumn[T]) Apply(chunk commit.Chunk, r *commit.Reader) {
	fill, data := c.chunkAt(chunk)
	c.apply(r, fill, data, c.option)
}

// Snapshot writes the entire column into the specified destination buffer
func (c *numericColumn[T]) Snapshot(chunk commit.Chunk, dst *commit.Buffer) {
	fill, data := c.chunkAt(chunk)
	fill.Range(func(x uint32) {
		c.write(dst, chunk.Min()+x, data[x])
	})
}

// --------------------------- Reader/Writer ----------------------------

// rdNumber represents a read-only accessor for simd.Numbers
type rdNumber[T simd.Number] struct {
	reader *numericColumn[T]
	txn    *Txn
}

// Get loads the value at the current transaction cursor
func (s rdNumber[T]) Get() (T, bool) {
	return s.reader.load(s.txn.cursor)
}

// Sum computes a sum of the column values selected by this transaction
func (s rdNumber[T]) Sum() (sum T) {
	s.txn.initialize()
	s.txn.rangeRead(func(chunk commit.Chunk, index bitmap.Bitmap) {
		if int(chunk) < len(s.reader.chunks) {
			sum += bitmap.Sum(s.reader.chunks[chunk].data, index)
		}
	})
	return sum
}

// Avg computes an arithmetic mean of the column values selected by this transaction
func (s rdNumber[T]) Avg() float64 {
	sum, ct := T(0), 0
	s.txn.initialize()
	s.txn.rangeRead(func(chunk commit.Chunk, index bitmap.Bitmap) {
		if int(chunk) < len(s.reader.chunks) {
			sum += bitmap.Sum(s.reader.chunks[chunk].data, index)
			ct += index.Count()
		}
	})
	return float64(sum) / float64(ct)
}

// Min finds the smallest value from the column values selected by this transaction
func (s rdNumber[T]) Min() (min T, ok bool) {
	s.txn.initialize()
	s.txn.rangeRead(func(chunk commit.Chunk, index bitmap.Bitmap) {
		if int(chunk) < len(s.reader.chunks) {
			if v, hit := bitmap.Min(s.reader.chunks[chunk].data, index); hit && (v < min || !ok) {
				min = v
				ok = true
			}
		}
	})
	return
}

// Max finds the largest value from the column values selected by this transaction
func (s rdNumber[T]) Max() (max T, ok bool) {
	s.txn.initialize()
	s.txn.rangeRead(func(chunk commit.Chunk, index bitmap.Bitmap) {
		if int(chunk) < len(s.reader.chunks) {
			if v, hit := bitmap.Max(s.reader.chunks[chunk].data, index); hit && (v > max || !ok) {
				max = v
				ok = true
			}
		}
	})
	return
}

// readNumberOf creates a new numeric reader
func readNumberOf[T simd.Number](txn *Txn, columnName string) rdNumber[T] {
	column, ok := txn.columnAt(columnName)
	if !ok {
		panic(fmt.Errorf("column: column '%s' does not exist", columnName))
	}

	reader, ok := column.Column.(*numericColumn[T])
	if !ok {
		panic(fmt.Errorf("column: column '%s' is not of type %T", columnName, T(0)))
	}

	return rdNumber[T]{
		reader: reader,
		txn:    txn,
	}
}
