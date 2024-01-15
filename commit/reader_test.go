// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package commit

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	buf := NewBuffer(0)
	buf.Reset("test")
	for i := uint32(0); i < 10; i++ {
		buf.PutUint64(Put, i, 2*uint64(i))
	}

	i := 0
	assert.Equal(t, 91, len(buf.buffer))

	r := NewReader()
	for r.Seek(buf); r.Next(); {
		assert.Equal(t, Put, r.Type)
		assert.Equal(t, i, int(r.Offset))
		assert.Equal(t, int(i*2), int(r.Uint64()))
		i++
	}
}

func TestRandom(t *testing.T) {
	seq := make([]uint32, 1024)
	for i := 0; i < len(seq); i++ {
		seq[i] = uint32(rand.Int31n(10000000))
	}

	buf := NewBuffer(0)
	for i := uint32(0); i < 1000; i++ {
		buf.PutUint32(Put, seq[i], uint32(rand.Int31()))
	}

	i := 0
	r := NewReader()
	for r.Seek(buf); r.Next(); {
		assert.Equal(t, Put, r.Type)
		assert.Equal(t, int(seq[i]), int(r.Offset))
		i++
	}
}

func TestRange(t *testing.T) {
	const count = 10000

	seq := make([]uint32, count)
	for i := 0; i < len(seq); i++ {
		seq[i] = uint32(rand.Int31n(1000000))
	}

	buf := NewBuffer(0)
	for i := uint32(0); i < count; i++ {
		buf.PutUint32(Put, seq[i], uint32(rand.Int31()))
	}

	r := NewReader()
	for i := 0; i < 100; i++ {
		r.Range(buf, Chunk(i), func(r *Reader) {
			for r.Next() {
				assert.Equal(t, Put, r.Type)
				assert.Equal(t, i, int(r.Offset>>chunkShift))
			}
		})
	}
}

func TestReadSwap(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutAny(Put, 10, int16(100))
	buf.PutAny(Put, 20, int32(200))
	buf.PutAny(Put, 30, int64(300))
	buf.PutAny(Put, 40, uint16(400))
	buf.PutAny(Put, 50, uint32(500))
	buf.PutAny(Put, 60, uint64(600))
	buf.PutAny(Put, 70, float32(700))
	buf.PutAny(Put, 80, float64(800))
	buf.PutAny(Put, 90, "900")
	buf.PutAny(Put, 100, []byte("binary"))
	buf.PutAny(Put, 110, true)
	buf.PutAny(Put, 120, int8(100))
	buf.PutAny(Put, 130, uint8(100))
	buf.PutAny(Put, 140, int(100))
	buf.PutAny(Put, 150, uint(100))
	buf.PutAny(Put, 160, float64(100))

	// Should only have 1 chunk
	assert.False(t, buf.IsEmpty())
	assert.Equal(t, 1, len(buf.chunks))
	buf.RangeChunks(func(chunk Chunk) {
		assert.Equal(t, Chunk(0), chunk)
	})

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, int16(100), r.Int16())
	assert.True(t, r.Next())
	assert.Equal(t, int32(200), r.Int32())
	assert.True(t, r.Next())
	assert.Equal(t, int64(300), r.Int64())
	assert.True(t, r.Next())
	assert.Equal(t, uint16(400), r.Uint16())
	assert.True(t, r.Next())
	assert.Equal(t, uint32(500), r.Uint32())
	assert.True(t, r.Next())
	assert.Equal(t, uint64(600), r.Uint64())
	assert.True(t, r.Next())
	assert.Equal(t, float32(700), r.Float32())
	assert.True(t, r.Next())
	assert.Equal(t, float64(800), r.Float64())
	assert.True(t, r.Next())
	assert.Equal(t, "900", r.String())
	assert.True(t, r.Next())
	assert.Equal(t, "binary", string(r.Bytes()))
	assert.True(t, r.Next())
	assert.Equal(t, true, r.Bool())
	assert.True(t, r.Next())
	assert.Equal(t, int16(100), r.Int16())
	assert.True(t, r.Next())
	assert.Equal(t, uint16(100), r.Uint16())
	assert.True(t, r.Next())
	assert.Equal(t, int(100), r.Int())
	assert.True(t, r.Next())
	assert.Equal(t, uint(100), r.Uint())

	// Rewind back and swap values
	r.Rewind()
	assert.True(t, r.Next())
	r.SwapInt16(99)
	assert.Equal(t, int16(99), r.Int16())
	assert.True(t, r.Next())
	r.SwapInt32(199)
	assert.Equal(t, int32(199), r.Int32())
	assert.True(t, r.Next())
	r.SwapInt64(299)
	assert.Equal(t, int64(299), r.Int64())
	assert.True(t, r.Next())
	r.SwapUint16(399)
	assert.Equal(t, uint16(399), r.Uint16())
	assert.True(t, r.Next())
	r.SwapUint32(499)
	assert.Equal(t, uint32(499), r.Uint32())
	assert.True(t, r.Next())
	r.SwapUint64(599)
	assert.Equal(t, uint64(599), r.Uint64())
	assert.True(t, r.Next())
	r.SwapFloat32(699)
	assert.Equal(t, float32(699), r.Float32())
	assert.True(t, r.Next())
	r.SwapFloat64(799)
	assert.Equal(t, float64(799), r.Float64())
	assert.True(t, r.Next())
	assert.True(t, r.Next())
	assert.True(t, r.Next())
	r.SwapBool(true)
	assert.Equal(t, true, r.Bool())
	assert.True(t, r.Next())
	assert.True(t, r.Next())
	assert.True(t, r.Next())
	r.SwapInt(300)
	assert.Equal(t, int(300), r.Int())
	assert.True(t, r.Next())
	r.SwapUint(400)
	assert.Equal(t, uint(400), r.Uint())
	assert.True(t, r.Next())
}

func TestWriteUnsupported(t *testing.T) {
	buf := NewBuffer(0)
	assert.Error(t, buf.PutAny(Put, 10, complex64(1)))
}

func TestReaderIface(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutFloat64(Put, 777, float64(1))

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, float64(1), r.Float())
	assert.Equal(t, uint32(777), r.Index())
}

func TestReadIntMixedSize(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutInt16(Put, 0, 10)
	buf.PutInt32(Put, 1, 20)
	buf.PutInt64(Put, 2, 30)
	buf.PutString(Put, 3, "hello")

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, 10, r.Int())
	assert.True(t, r.Next())
	assert.Equal(t, 20, r.Int())
	assert.True(t, r.Next())
	assert.Equal(t, 30, r.Int())
	assert.True(t, r.Next())
	assert.Panics(t, func() {
		r.Int()
	})
}

func TestReadFloatMixedSize(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutFloat32(Put, 0, 10)
	buf.PutFloat64(Put, 1, 20)
	buf.PutString(Put, 3, "hello")

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, 10.0, r.Float())
	assert.True(t, r.Next())
	assert.Equal(t, 20.0, r.Float())
	assert.True(t, r.Next())
	assert.Panics(t, func() {
		r.Float()
	})
}

func TestReadSize(t *testing.T) {
	buf := NewBuffer(0)
	buf.Reset("test")
	buf.PutBool(123, true)

	r := NewReader()
	r.readFixed(buf.buffer[0])
	assert.Equal(t, 0, r.i1-r.i0)
}

func TestIndexAtChunk(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutFloat32(Put, 10000, 10)
	buf.PutFloat32(Put, 20000, 10)
	buf.PutFloat32(Put, 30000, 10)

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, uint32(10000), r.IndexAtChunk())
	assert.True(t, r.Next())
	assert.Equal(t, uint32(3616), r.IndexAtChunk())
}

func TestSwapOpChange(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutInt32(Merge, 10, int32(1))
	assert.Equal(t, []byte{0x23, 0x0, 0x0, 0x0, 0x1, 0xa}, buf.buffer)

	// Swap the value, this should also change the type
	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, Merge, r.Type)
	r.SwapInt32(int32(2))
	assert.Equal(t, int32(2), r.Int32())

	// Once swapped, op type should be changed to "Put"
	r.Seek(buf)
	assert.Equal(t, []byte{0x22, 0x0, 0x0, 0x0, 0x2, 0xa}, buf.buffer)
	assert.True(t, r.Next())
	assert.Equal(t, Put, r.Type)
}

func TestMergeBytes(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutBytes(Merge, 10, []byte("A"))
	assert.Equal(t, []byte{0x53, 0x0, 0x1, 0x41, 0xa}, buf.buffer)

	// Swap the value, this should also change the type
	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, Merge, r.Type)
	r.SwapBytes([]byte("B"))

	// Once swapped, op type should be changed to "Put"
	r.Seek(buf)
	assert.Equal(t, []byte{0x52, 0x0, 0x1, 0x42, 0xa}, buf.buffer)
	assert.True(t, r.Next())
	assert.Equal(t, Put, r.Type)
}

func TestMergeStrings(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutBytes(Merge, 30, []byte("5"))
	buf.PutBytes(Merge, 40, []byte("6"))
	buf.PutBytes(Merge, 10, []byte("2"))
	buf.PutBytes(Merge, 20, []byte("3"))

	var scanned []string

	// Swap the value, this should also change the type
	r := NewReader()
	r.Range(buf, 0, func(r *Reader) {
		for r.Rewind(); r.Next(); {
			i, _ := strconv.Atoi(r.String())
			r.SwapString(strconv.Itoa(i * i))
		}
	})

	r.Range(buf, 0, func(r *Reader) {
		for r.Rewind(); r.Next(); {
			scanned = append(scanned, fmt.Sprintf("(%s) %v", r.Type, r.String()))
		}
	})

	assert.Equal(t, []string{
		"(skip) 5",
		"(skip) 6",
		"(put) 4",
		"(put) 9",
		"(put) 25",
		"(put) 36",
	}, scanned)
}

func TestReaderIsUpsert(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutFloat32(Put, 0, 10)
	buf.PutFloat32(Delete, 0, 0)

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.True(t, r.IsUpsert())
	assert.True(t, r.Next())
	assert.True(t, r.IsDelete())
}
