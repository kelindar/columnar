package commit

import (
	"fmt"
	"math"
	"reflect"
	"unsafe"
)

const (
	chunkShift = 14     // 16K elements
	size1      = 0      // 1 byte in size
	size2      = 1 << 4 // 2 bytes in size
	size4      = 2 << 4 // 4 bytes in size
	size8      = 3 << 4 // 8 bytes in size
	isNext     = 1 << 7 // is immediate next
	isString   = 1 << 6 // is variable-size string
)

// --------------------------- Delta log ----------------------------

// Queue represents a queue of delta operations.
type Queue struct {
	last   int32   // The last offset writte
	chunk  int32   // The current chunk
	buffer []byte  // The destination buffer
	chunks []int32 // The offsets of chunks
	_      [8]byte // padding
	Column string  // The column for the queue
}

// NewQueue creates a new queue to store individual operations.
func NewQueue(capacity int) *Queue {
	return &Queue{
		buffer: make([]byte, 0, capacity),
	}
}

// Reset resets the queue so it can be reused.
func (q *Queue) Reset(column string) {
	q.chunks = q.chunks[:0]
	q.buffer = q.buffer[:0]
	q.last = 0
	q.chunk = -1
	q.Column = column
}

// Put appends a value of any supported type onto the queue.
func (q *Queue) Put(op UpdateType, idx uint32, value interface{}) {
	switch v := value.(type) {
	case uint64:
		q.PutUint64(op, idx, v)
	case uint32:
		q.PutUint32(op, idx, v)
	case uint16:
		q.PutUint16(op, idx, v)
	default:
		panic(fmt.Errorf("column: unsupported type %T", value))
	}
}

// PutUint64 appends a uint64 value.
func (q *Queue) PutUint64(op UpdateType, idx uint32, value uint64) {
	q.writeChunk(idx)
	delta := int32(idx) - q.last
	q.last = int32(idx)
	if delta == 1 {
		q.buffer = append(q.buffer,
			byte(op)|size8|isNext,
			byte(value>>56), byte(value>>48), byte(value>>40), byte(value>>32),
			byte(value>>24), byte(value>>16), byte(value>>8), byte(value),
		)
		return
	}

	q.buffer = append(q.buffer,
		byte(op)|size8,
		byte(value>>56), byte(value>>48), byte(value>>40), byte(value>>32),
		byte(value>>24), byte(value>>16), byte(value>>8), byte(value),
	)
	q.writeOffset(uint32(delta))
}

// PutUint32 appends a uint32 value.
func (q *Queue) PutUint32(op UpdateType, idx uint32, value uint32) {
	q.writeChunk(idx)
	delta := int32(idx) - q.last
	q.last = int32(idx)
	if delta == 1 {
		q.buffer = append(q.buffer,
			byte(op)|size4|isNext,
			byte(value>>24), byte(value>>16), byte(value>>8), byte(value),
		)
		return
	}

	q.buffer = append(q.buffer,
		byte(op)|size4,
		byte(value>>24), byte(value>>16), byte(value>>8), byte(value),
	)
	q.writeOffset(uint32(delta))
}

// PutUint16 appends a uint16 value.
func (q *Queue) PutUint16(op UpdateType, idx uint32, value uint16) {
	q.writeChunk(idx)
	delta := int32(idx) - q.last
	q.last = int32(idx)
	if delta == 1 {
		q.buffer = append(q.buffer, byte(op)|size2|isNext, byte(value>>8), byte(value))
		return
	}

	q.buffer = append(q.buffer, byte(op)|size2, byte(value>>8), byte(value))
	q.writeOffset(uint32(delta))
}

// PutBool appends a boolean value.
func (q *Queue) PutBool(op UpdateType, idx uint32, value bool) {
	q.writeChunk(idx)
	delta := int32(idx) - q.last
	q.last = int32(idx)

	// let the compiler do its magic: https://github.com/golang/go/issues/6011
	v := 0
	if value {
		v = 1
	}

	if delta == 1 {
		q.buffer = append(q.buffer, byte(op)|size1|isNext, byte(v))
		return
	}

	q.buffer = append(q.buffer, byte(op)|size1, byte(v))
	q.writeOffset(uint32(delta))
}

// PutInt64 appends an int64 value.
func (q *Queue) PutInt64(op UpdateType, idx uint32, value int64) {
	q.PutUint64(op, idx, uint64(value))
}

// PutInt32 appends an int32 value.
func (q *Queue) PutInt32(op UpdateType, idx uint32, value int32) {
	q.PutUint32(op, idx, uint32(value))
}

// PutInt16 appends an int16 value.
func (q *Queue) PutInt16(op UpdateType, idx uint32, value int16) {
	q.PutUint16(op, idx, uint16(value))
}

// PutFloat64 appends a float64 value.
func (q *Queue) PutFloat64(op UpdateType, idx uint32, value float64) {
	q.PutUint64(op, idx, math.Float64bits(value))
}

// PutFloat32 appends an int32 value.
func (q *Queue) PutFloat32(op UpdateType, idx uint32, value float32) {
	q.PutUint32(op, idx, math.Float32bits(value))
}

// PutBytes appends a binary value.
func (q *Queue) PutBytes(op UpdateType, idx uint32, value []byte) {
	q.writeChunk(idx)
	delta := int32(idx) - q.last
	q.last = int32(idx)

	// Write a 2-byte length (max 65K slices)
	length := len(value)
	if delta == 1 {
		q.buffer = append(q.buffer,
			byte(op)|size2|isString|isNext,
			byte(length>>8), byte(length),
		)
		q.buffer = append(q.buffer, value...)
		return
	}

	q.buffer = append(q.buffer,
		byte(op)|size2|isString,
		byte(length>>8), byte(length),
	)

	// Write the the data itself and the offset
	q.buffer = append(q.buffer, value...)
	q.writeOffset(uint32(delta))
}

// PutString appends a string value.
func (q *Queue) PutString(op UpdateType, idx uint32, value string) {
	q.PutBytes(op, idx, toBytes(value))
}

// writeOffset writes the offset at the current head.
func (q *Queue) writeOffset(delta uint32) {
	for delta >= 0x80 {
		q.buffer = append(q.buffer, byte(delta)|0x80)
		delta >>= 7
	}

	q.buffer = append(q.buffer, byte(delta))
}

// writeChunk writes a chunk if changed
func (q *Queue) writeChunk(idx uint32) {
	if chunk := int32(idx >> chunkShift); q.chunk != chunk {
		q.chunks = append(q.chunks, int32(len(q.buffer)))
		q.chunk = chunk
	}
}

// toBytes converts a string to a byte slice without allocating.
func toBytes(v string) (b []byte) {
	strHeader := (*reflect.StringHeader)(unsafe.Pointer(&v))
	byteHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	byteHeader.Data = strHeader.Data

	l := len(v)
	byteHeader.Len = l
	byteHeader.Cap = l
	return
}
