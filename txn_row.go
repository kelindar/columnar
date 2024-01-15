// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package column

import (
	"encoding"
	"fmt"

	"github.com/kelindar/column/commit"
)

// Row represents a cursor at a particular row offest in the transaction.
type Row struct {
	txn *Txn
}

// Index returns the index of the row
func (r Row) Index() uint32 {
	return r.txn.Index()
}

// --------------------------- Numbers ----------------------------

// Int loads a int value at a particular column
func (r Row) Int(columnName string) (v int, ok bool) {
	return readNumber[int](r.txn, columnName)
}

// SetInt stores a int value at a particular column
func (r Row) SetInt(columnName string, value int) {
	r.txn.Int(columnName).Set(value)
}

// MergeInt atomically merges a delta into int value at a particular column
func (r Row) MergeInt(columnName string, value int) {
	r.txn.Int(columnName).Merge(value)
}

// Int16 loads a int16 value at a particular column
func (r Row) Int16(columnName string) (v int16, ok bool) {
	return readNumber[int16](r.txn, columnName)
}

// SetInt16 stores a int16 value at a particular column
func (r Row) SetInt16(columnName string, value int16) {
	r.txn.Int16(columnName).Set(value)
}

// MergeInt16 atomically merges a delta into int16 value at a particular column
func (r Row) MergeInt16(columnName string, value int16) {
	r.txn.Int16(columnName).Merge(value)
}

// Int32 loads a int32 value at a particular column
func (r Row) Int32(columnName string) (v int32, ok bool) {
	return readNumber[int32](r.txn, columnName)
}

// SetInt32 stores a int32 value at a particular column
func (r Row) SetInt32(columnName string, value int32) {
	r.txn.Int32(columnName).Set(value)
}

// MergeInt32 atomically merges a delta into int32 value at a particular column
func (r Row) MergeInt32(columnName string, value int32) {
	r.txn.Int32(columnName).Merge(value)
}

// Int64 loads a int64 value at a particular column
func (r Row) Int64(columnName string) (v int64, ok bool) {
	return readNumber[int64](r.txn, columnName)
}

// SetInt64 stores a int64 value at a particular column
func (r Row) SetInt64(columnName string, value int64) {
	r.txn.Int64(columnName).Set(value)
}

// MergeInt64 atomically merges a delta into int64 value at a particular column
func (r Row) MergeInt64(columnName string, value int64) {
	r.txn.Int64(columnName).Merge(value)
}

// Uint loads a uint value at a particular column
func (r Row) Uint(columnName string) (v uint, ok bool) {
	return readNumber[uint](r.txn, columnName)
}

// SetUint stores a uint value at a particular column
func (r Row) SetUint(columnName string, value uint) {
	r.txn.Uint(columnName).Set(value)
}

// MergeUint atomically merges a delta into uint value at a particular column
func (r Row) MergeUint(columnName string, value uint) {
	r.txn.Uint(columnName).Merge(value)
}

// Uint16 loads a uint16 value at a particular column
func (r Row) Uint16(columnName string) (v uint16, ok bool) {
	return readNumber[uint16](r.txn, columnName)
}

// SetUint16 stores a uint16 value at a particular column
func (r Row) SetUint16(columnName string, value uint16) {
	r.txn.Uint16(columnName).Set(value)
}

// MergeUint16 atomically merges a delta into uint16 value at a particular column
func (r Row) MergeUint16(columnName string, value uint16) {
	r.txn.Uint16(columnName).Merge(value)
}

// Uint32 loads a uint32 value at a particular column
func (r Row) Uint32(columnName string) (v uint32, ok bool) {
	return readNumber[uint32](r.txn, columnName)
}

// SetUint32 stores a uint32 value at a particular column
func (r Row) SetUint32(columnName string, value uint32) {
	r.txn.Uint32(columnName).Set(value)
}

// MergeUint32 atomically merges a delta into uint32 value at a particular column
func (r Row) MergeUint32(columnName string, value uint32) {
	r.txn.Uint32(columnName).Merge(value)
}

// Uint64 loads a uint64 value at a particular column
func (r Row) Uint64(columnName string) (v uint64, ok bool) {
	return readNumber[uint64](r.txn, columnName)
}

// SetUint64 stores a uint64 value at a particular column
func (r Row) SetUint64(columnName string, value uint64) {
	r.txn.Uint64(columnName).Set(value)
}

// MergeUint64 atomically merges a delta into uint64 value at a particular column
func (r Row) MergeUint64(columnName string, value uint64) {
	r.txn.Uint64(columnName).Merge(value)
}

// Float32 loads a float32 value at a particular column
func (r Row) Float32(columnName string) (v float32, ok bool) {
	return readNumber[float32](r.txn, columnName)
}

// SetFloat32 stores a float32 value at a particular column
func (r Row) SetFloat32(columnName string, value float32) {
	r.txn.Float32(columnName).Set(value)
}

// MergeFloat32 atomically merges a delta into float32 value at a particular column
func (r Row) MergeFloat32(columnName string, value float32) {
	r.txn.Float32(columnName).Merge(value)
}

// Float64 loads a float64 value at a particular column
func (r Row) Float64(columnName string) (float64, bool) {
	return readNumber[float64](r.txn, columnName)
}

// SetFloat64 stores a float64 value at a particular column
func (r Row) SetFloat64(columnName string, value float64) {
	r.txn.Float64(columnName).Set(value)
}

// MergeFloat64 atomically merges a delta into float64 value at a particular column
func (r Row) MergeFloat64(columnName string, value float64) {
	r.txn.Float64(columnName).Merge(value)
}

// --------------------------- Strings ----------------------------

// Key loads a primary key value at a particular column
func (r Row) Key() (v string, ok bool) {
	if pk := r.txn.owner.pk; pk != nil {
		v, ok = pk.LoadString(r.txn.cursor)
	}
	return
}

// SetKey stores a primary key value at a particular column
func (r Row) SetKey(key string) {
	r.txn.Key().Set(key)
}

// String loads a string value at a particular column
func (r Row) String(columnName string) (v string, ok bool) {
	return readStringOf[*columnString](r.txn, columnName).Get()
}

// SetString stores a string value at a particular column
func (r Row) SetString(columnName string, value string) {
	r.txn.String(columnName).Set(value)
}

// MergeString merges a string value at a particular column
func (r Row) MergeString(columnName string, value string) {
	r.txn.String(columnName).Merge(value)
}

// Enum loads a string value at a particular column
func (r Row) Enum(columnName string) (v string, ok bool) {
	return readStringOf[*columnEnum](r.txn, columnName).Get()
}

// SetEnum stores a string value at a particular column
func (r Row) SetEnum(columnName string, value string) {
	r.txn.Enum(columnName).Set(value)
}

// --------------------------- Records ----------------------------

// Record loads a record value at a particular column
func (r Row) Record(columnName string) (any, bool) {
	return readRecordOf(r.txn, columnName).Get()
}

// SetRecord stores a record value at a particular column
func (r Row) SetRecord(columnName string, value encoding.BinaryMarshaler) error {
	return r.txn.Record(columnName).Set(value)
}

// MergeRecord merges a record value at a particular column
func (r Row) MergeRecord(columnName string, delta encoding.BinaryMarshaler) error {
	return r.txn.Record(columnName).Merge(delta)
}

// --------------------------- Map ----------------------------

// SetMany stores a set of columns for a given map
func (r Row) SetMany(value map[string]any) error {
	for k, v := range value {
		if _, ok := r.txn.columnAt(k); !ok {
			return fmt.Errorf("unable to set '%s', no such column", k)
		}

		if err := r.txn.bufferFor(k).PutAny(commit.Put, r.txn.cursor, v); err != nil {
			return err
		}
	}
	return nil
}

// --------------------------- Others ----------------------------

// Bool loads a bool value at a particular column
func (r Row) Bool(columnName string) bool {
	return readBoolOf(r.txn, columnName).Get()
}

// SetBool stores a bool value at a particular column
func (r Row) SetBool(columnName string, value bool) {
	r.txn.Bool(columnName).Set(value)
}

// Any loads a bool value at a particular column
func (r Row) Any(columnName string) (any, bool) {
	return readAnyOf(r.txn, columnName).Get()
}

// SetAny stores a bool value at a particular column
func (r Row) SetAny(columnName string, value interface{}) {
	r.txn.Any(columnName).Set(value)
}
