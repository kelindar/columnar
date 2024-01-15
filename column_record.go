// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"encoding"
	"reflect"
	"sync"
	"unsafe"

	"github.com/kelindar/column/commit"
)

type recordType interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

// --------------------------- Record ----------------------------

// columnRecord represents a typed column that is persisted using binary marshaler
type columnRecord struct {
	columnString
	pool *sync.Pool
}

// ForRecord creates a new column that contains a type marshaled into/from binary. It requires
// a constructor for the type as well as optional merge function. If merge function is
// set to nil, "overwrite" strategy will be used.
func ForRecord[T recordType](new func() T, opts ...func(*option[T])) Column {
	mergeFunc := configure(opts, option[T]{
		Merge: func(value, delta T) T { return delta },
	}).Merge

	pool := &sync.Pool{
		New: func() any { return new() },
	}

	// Merge function that decodes, merges and re-encodes records into their
	// respective binary representation.
	mergeRecord := func(v, d string) string {
		value := pool.Get().(T)
		delta := pool.Get().(T)
		defer pool.Put(value)
		defer pool.Put(delta)

		// Unmarshal the existing value
		err1 := value.UnmarshalBinary(s2b(v))
		err2 := delta.UnmarshalBinary(s2b(d))
		if err1 != nil || err2 != nil {
			return v
		}

		// Apply the user-defined merging strategy and marshal it back
		merged := mergeFunc(value, delta)
		if encoded, err := merged.MarshalBinary(); err == nil {
			return b2s(&encoded)
		}
		return v
	}

	return &columnRecord{
		pool: pool,
		columnString: columnString{
			chunks: make(chunks[string], 0, 4),
			option: option[string]{
				Merge: mergeRecord,
			},
		},
	}
}

// Value returns the value at the given index
// TODO: should probably get rid of this and use an `rdRecord` instead
func (c *columnRecord) Value(idx uint32) (out any, has bool) {
	if v, ok := c.columnString.Value(idx); ok {
		out = c.pool.New()
		has = out.(encoding.BinaryUnmarshaler).UnmarshalBinary(s2b(v.(string))) == nil
	}
	return
}

// --------------------------- Writer ----------------------------

// rwRecord represents read-write accessor for primary keys.
type rwRecord struct {
	rdRecord
	writer *commit.Buffer
}

// Set sets the value at the current transaction index
func (s rwRecord) Set(value encoding.BinaryMarshaler) error {
	return s.write(commit.Put, value.MarshalBinary)
}

// Merge atomically merges a delta to the value at the current transaction cursor
func (s rwRecord) Merge(delta encoding.BinaryMarshaler) error {
	return s.write(commit.Merge, delta.MarshalBinary)
}

// write writes the operation
func (s rwRecord) write(op commit.OpType, encodeDelta func() ([]byte, error)) error {
	v, err := encodeDelta()
	if err == nil {
		s.writer.PutBytes(op, *s.cursor, v)
	}
	return err
}

// As creates a read-write accessor for a specific record type.
func (txn *Txn) Record(columnName string) rwRecord {
	return rwRecord{
		rdRecord: readRecordOf(txn, columnName),
		writer:   txn.bufferFor(columnName),
	}
}

// --------------------------- Reader ----------------------------

// rdRecord represents a read-only accessor for records
type rdRecord reader[*columnRecord]

// Get loads the value at the current transaction index
func (s rdRecord) Get() (any, bool) {
	value := s.reader.pool.New().(encoding.BinaryUnmarshaler)
	if s.Unmarshal(value.UnmarshalBinary) {
		return value, true
	}

	return nil, false
}

// Unmarshal loads the value at the current transaction index using a
// specified function to decode the value.
func (s rdRecord) Unmarshal(decode func(data []byte) error) bool {
	encoded, ok := s.reader.LoadString(*s.cursor)
	if !ok {
		return false
	}

	return decode(s2b(encoded)) == nil
}

// readRecordOf creates a read-only accessor for readers
func readRecordOf(txn *Txn, columnName string) rdRecord {
	return rdRecord(readerFor[*columnRecord](txn, columnName))
}

// --------------------------- Convert ----------------------------

// b2s converts byte slice to a string without allocating.
func b2s(b *[]byte) string {
	return *(*string)(unsafe.Pointer(b))
}

// s2b converts a string to a byte slice without allocating.
func s2b(v string) (b []byte) {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	return
}
