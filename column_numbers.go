// This code was generated, DO NOT EDIT.
// Any changes will be lost if this file is regenerated.

package column

import (
	"github.com/kelindar/bitmap"
	"github.com/kelindar/column/commit"
)

// --------------------------- Int ----------------------------

// makeInts creates a new vector for ints
func makeInts(opts ...func(*option[int])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value int) { buffer.PutInt(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []int, opts option[int]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Int()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapInt(opts.Merge(data[offset], r.Int()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwInt represents a read-write cursor for int
type rwInt struct {
	rdNumber[int]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwInt) Set(value int) {
	s.writer.PutInt(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwInt) Merge(delta int) {
	s.writer.PutInt(commit.Merge, s.txn.cursor, delta)
}

// Int returns a read-write accessor for int column
func (txn *Txn) Int(columnName string) rwInt {
	return rwInt{
		rdNumber: readNumberOf[int](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Int16 ----------------------------

// makeInt16s creates a new vector for int16s
func makeInt16s(opts ...func(*option[int16])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value int16) { buffer.PutInt16(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []int16, opts option[int16]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Int16()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapInt16(opts.Merge(data[offset], r.Int16()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwInt16 represents a read-write cursor for int16
type rwInt16 struct {
	rdNumber[int16]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwInt16) Set(value int16) {
	s.writer.PutInt16(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwInt16) Merge(delta int16) {
	s.writer.PutInt16(commit.Merge, s.txn.cursor, delta)
}

// Int16 returns a read-write accessor for int16 column
func (txn *Txn) Int16(columnName string) rwInt16 {
	return rwInt16{
		rdNumber: readNumberOf[int16](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Int32 ----------------------------

// makeInt32s creates a new vector for int32s
func makeInt32s(opts ...func(*option[int32])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value int32) { buffer.PutInt32(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []int32, opts option[int32]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Int32()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapInt32(opts.Merge(data[offset], r.Int32()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwInt32 represents a read-write cursor for int32
type rwInt32 struct {
	rdNumber[int32]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwInt32) Set(value int32) {
	s.writer.PutInt32(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwInt32) Merge(delta int32) {
	s.writer.PutInt32(commit.Merge, s.txn.cursor, delta)
}

// Int32 returns a read-write accessor for int32 column
func (txn *Txn) Int32(columnName string) rwInt32 {
	return rwInt32{
		rdNumber: readNumberOf[int32](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Int64 ----------------------------

// makeInt64s creates a new vector for int64s
func makeInt64s(opts ...func(*option[int64])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value int64) { buffer.PutInt64(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []int64, opts option[int64]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Int64()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapInt64(opts.Merge(data[offset], r.Int64()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwInt64 represents a read-write cursor for int64
type rwInt64 struct {
	rdNumber[int64]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwInt64) Set(value int64) {
	s.writer.PutInt64(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwInt64) Merge(delta int64) {
	s.writer.PutInt64(commit.Merge, s.txn.cursor, delta)
}

// Int64 returns a read-write accessor for int64 column
func (txn *Txn) Int64(columnName string) rwInt64 {
	return rwInt64{
		rdNumber: readNumberOf[int64](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Uint ----------------------------

// makeUints creates a new vector for uints
func makeUints(opts ...func(*option[uint])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value uint) { buffer.PutUint(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []uint, opts option[uint]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Uint()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapUint(opts.Merge(data[offset], r.Uint()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwUint represents a read-write cursor for uint
type rwUint struct {
	rdNumber[uint]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwUint) Set(value uint) {
	s.writer.PutUint(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwUint) Merge(delta uint) {
	s.writer.PutUint(commit.Merge, s.txn.cursor, delta)
}

// Uint returns a read-write accessor for uint column
func (txn *Txn) Uint(columnName string) rwUint {
	return rwUint{
		rdNumber: readNumberOf[uint](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Uint16 ----------------------------

// makeUint16s creates a new vector for uint16s
func makeUint16s(opts ...func(*option[uint16])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value uint16) { buffer.PutUint16(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []uint16, opts option[uint16]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Uint16()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapUint16(opts.Merge(data[offset], r.Uint16()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwUint16 represents a read-write cursor for uint16
type rwUint16 struct {
	rdNumber[uint16]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwUint16) Set(value uint16) {
	s.writer.PutUint16(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwUint16) Merge(delta uint16) {
	s.writer.PutUint16(commit.Merge, s.txn.cursor, delta)
}

// Uint16 returns a read-write accessor for uint16 column
func (txn *Txn) Uint16(columnName string) rwUint16 {
	return rwUint16{
		rdNumber: readNumberOf[uint16](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Uint32 ----------------------------

// makeUint32s creates a new vector for uint32s
func makeUint32s(opts ...func(*option[uint32])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value uint32) { buffer.PutUint32(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []uint32, opts option[uint32]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Uint32()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapUint32(opts.Merge(data[offset], r.Uint32()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwUint32 represents a read-write cursor for uint32
type rwUint32 struct {
	rdNumber[uint32]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwUint32) Set(value uint32) {
	s.writer.PutUint32(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwUint32) Merge(delta uint32) {
	s.writer.PutUint32(commit.Merge, s.txn.cursor, delta)
}

// Uint32 returns a read-write accessor for uint32 column
func (txn *Txn) Uint32(columnName string) rwUint32 {
	return rwUint32{
		rdNumber: readNumberOf[uint32](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Uint64 ----------------------------

// makeUint64s creates a new vector for uint64s
func makeUint64s(opts ...func(*option[uint64])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value uint64) { buffer.PutUint64(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []uint64, opts option[uint64]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Uint64()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapUint64(opts.Merge(data[offset], r.Uint64()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwUint64 represents a read-write cursor for uint64
type rwUint64 struct {
	rdNumber[uint64]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwUint64) Set(value uint64) {
	s.writer.PutUint64(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwUint64) Merge(delta uint64) {
	s.writer.PutUint64(commit.Merge, s.txn.cursor, delta)
}

// Uint64 returns a read-write accessor for uint64 column
func (txn *Txn) Uint64(columnName string) rwUint64 {
	return rwUint64{
		rdNumber: readNumberOf[uint64](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Float32 ----------------------------

// makeFloat32s creates a new vector for float32s
func makeFloat32s(opts ...func(*option[float32])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value float32) { buffer.PutFloat32(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []float32, opts option[float32]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Float32()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapFloat32(opts.Merge(data[offset], r.Float32()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwFloat32 represents a read-write cursor for float32
type rwFloat32 struct {
	rdNumber[float32]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwFloat32) Set(value float32) {
	s.writer.PutFloat32(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwFloat32) Merge(delta float32) {
	s.writer.PutFloat32(commit.Merge, s.txn.cursor, delta)
}

// Float32 returns a read-write accessor for float32 column
func (txn *Txn) Float32(columnName string) rwFloat32 {
	return rwFloat32{
		rdNumber: readNumberOf[float32](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Float64 ----------------------------

// makeFloat64s creates a new vector for float64s
func makeFloat64s(opts ...func(*option[float64])) Column {
	return makeNumeric(
		func(buffer *commit.Buffer, idx uint32, value float64) { buffer.PutFloat64(commit.Put, idx, value) },
		func(r *commit.Reader, fill bitmap.Bitmap, data []float64, opts option[float64]) {
			for r.Next() {
				offset := r.IndexAtChunk()
				switch r.Type {
				case commit.Put:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.Float64()
				case commit.Merge:
					fill[offset>>6] |= 1 << (offset & 0x3f)
					data[offset] = r.SwapFloat64(opts.Merge(data[offset], r.Float64()))
				case commit.Delete:
					fill.Remove(offset)
				}
			}
		}, opts,
	)
}

// rwFloat64 represents a read-write cursor for float64
type rwFloat64 struct {
	rdNumber[float64]
	writer *commit.Buffer
}

// Set sets the value at the current transaction cursor
func (s rwFloat64) Set(value float64) {
	s.writer.PutFloat64(commit.Put, s.txn.cursor, value)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwFloat64) Merge(delta float64) {
	s.writer.PutFloat64(commit.Merge, s.txn.cursor, delta)
}

// Float64 returns a read-write accessor for float64 column
func (txn *Txn) Float64(columnName string) rwFloat64 {
	return rwFloat64{
		rdNumber: readNumberOf[float64](txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}
