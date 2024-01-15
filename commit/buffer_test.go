// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package commit

import (
	"bytes"
	"testing"
	"time"
	"unsafe"

	"github.com/kelindar/bitmap"
	"github.com/stretchr/testify/assert"
)

/*
cpu: Intel(R) Core(TM) i7-9700K CPU @ 3.60GHz
BenchmarkQueue/u16-rw-8                      154           7691836 ns/op              19 B/op          0 allocs/op
BenchmarkQueue/u16-next-8                    214           5542922 ns/op               7 B/op          0 allocs/op
BenchmarkQueue/u32-rw-8                      152           7743216 ns/op              20 B/op          0 allocs/op
BenchmarkQueue/u32-next-8                    212           5616605 ns/op               7 B/op          0 allocs/op
BenchmarkQueue/u64-rw-8                      148           8000536 ns/op              20 B/op          0 allocs/op
BenchmarkQueue/u64-next-8                    194           6126377 ns/op               7 B/op          0 allocs/op
BenchmarkQueue/str-rw-8                       91          12935521 ns/op              33 B/op          0 allocs/op
BenchmarkQueue/str-next-8                     98          10901156 ns/op              15 B/op          0 allocs/op
BenchmarkQueue/bool-rw-8                     169           6950441 ns/op              18 B/op          0 allocs/op
BenchmarkQueue/bool-next-8                   228           5195821 ns/op               6 B/op          0 allocs/op
*/
func BenchmarkQueue(b *testing.B) {
	const count = 1000000

	run("u16-rw", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count*2; i += 2 {
			buf.PutUint16(Put, i, uint16(i))
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint16()
		}
	})

	run("u16-next", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count; i++ {
			buf.PutUint16(Put, i, uint16(i))
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint16()
		}
	})

	run("u32-rw", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count*2; i += 2 {
			buf.PutUint32(Put, i, i)
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint32()
		}
	})

	run("u32-next", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count; i++ {
			buf.PutUint32(Put, i, i)
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint32()
		}
	})

	run("u64-rw", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count*2; i += 2 {
			buf.PutUint64(Put, i, uint64(i))
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint64()
		}
	})

	run("u64-next", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count; i++ {
			buf.PutUint64(Put, i, uint64(i))
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Uint64()
		}
	})

	run("str-rw", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count*2; i += 2 {
			buf.PutString(Put, i, "hello world")
		}
		for r.Seek(buf); r.Next(); {
			_ = r.String()
		}
	})

	run("str-next", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count; i++ {
			buf.PutString(Put, i, "hello world")
		}
		for r.Seek(buf); r.Next(); {
			_ = r.String()
		}
	})

	run("bool-rw", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count*2; i += 2 {
			buf.PutBool(i, true)
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Bool()
		}
	})

	run("bool-next", b, count, func(buf *Buffer, r *Reader) {
		for i := uint32(0); i < count; i++ {
			buf.PutBool(i, true)
		}
		for r.Seek(buf); r.Next(); {
			_ = r.Bool()
		}
	})
}

// Run runs a single benchmark
func run(name string, b *testing.B, count int, fn func(buf *Buffer, r *Reader)) {
	b.Run(name, func(b *testing.B) {
		buf := NewBuffer(count * 20)
		r := NewReader()
		b.ReportAllocs()
		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			buf.Reset("test")
			fn(buf, r)
		}
	})
}

func TestSizeof(t *testing.T) {
	assert.Equal(t, 96, int(unsafe.Sizeof(Reader{})))
	assert.Equal(t, 80, int(unsafe.Sizeof(Buffer{})))
}

func TestReadWrite(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutInt16(Put, 10, 100)
	buf.PutInt16(Put, 11, 100)
	buf.PutInt32(Put, 20, 200)
	buf.PutInt32(Put, 21, 200)
	buf.PutInt64(Put, 30, 300)
	buf.PutInt64(Put, 31, 300)
	buf.PutUint16(Put, 40, 400)
	buf.PutUint16(Put, 41, 400)
	buf.PutUint32(Put, 50, 500)
	buf.PutUint32(Put, 51, 500)
	buf.PutUint64(Put, 60, 600)
	buf.PutUint64(Put, 61, 600)
	buf.PutFloat32(Put, 70, 700)
	buf.PutFloat32(Put, 71, 700)
	buf.PutFloat64(Put, 80, 800)
	buf.PutFloat64(Put, 81, 800)
	buf.PutString(Put, 90, "900")
	buf.PutString(Put, 91, "hello world")
	buf.PutBytes(Put, 100, []byte("binary"))
	buf.PutBool(110, true)
	buf.PutBool(111, false)
	buf.PutInt(Put, 120, 1000)
	buf.PutUint(Put, 130, 1100)
	buf.PutNumber(Put, 140, 12.34)

	// Read values back
	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, int16(100), r.Int16())
	assert.True(t, r.Next())
	assert.Equal(t, int16(100), r.Int16())
	assert.True(t, r.Next())
	assert.Equal(t, int32(200), r.Int32())
	assert.True(t, r.Next())
	assert.Equal(t, int32(200), r.Int32())
	assert.True(t, r.Next())
	assert.Equal(t, int64(300), r.Int64())
	assert.True(t, r.Next())
	assert.Equal(t, int64(300), r.Int64())
	assert.True(t, r.Next())
	assert.Equal(t, uint16(400), r.Uint16())
	assert.True(t, r.Next())
	assert.Equal(t, uint16(400), r.Uint16())
	assert.True(t, r.Next())
	assert.Equal(t, uint32(500), r.Uint32())
	assert.True(t, r.Next())
	assert.Equal(t, uint32(500), r.Uint32())
	assert.True(t, r.Next())
	assert.Equal(t, uint64(600), r.Uint64())
	assert.True(t, r.Next())
	assert.Equal(t, uint64(600), r.Uint64())
	assert.True(t, r.Next())
	assert.Equal(t, float32(700), r.Float32())
	assert.True(t, r.Next())
	assert.Equal(t, float32(700), r.Float32())
	assert.True(t, r.Next())
	assert.Equal(t, float64(800), r.Float64())
	assert.True(t, r.Next())
	assert.Equal(t, float64(800), r.Float64())
	assert.True(t, r.Next())
	assert.Equal(t, "900", r.String())
	assert.True(t, r.Next())
	assert.Equal(t, "hello world", r.String())
	assert.True(t, r.Next())
	assert.Equal(t, "binary", string(r.Bytes()))
	assert.True(t, r.Next())
	assert.Equal(t, true, r.Bool())
	assert.True(t, r.Next())
	assert.Equal(t, false, r.Bool())
	assert.True(t, r.Next())
	assert.Equal(t, int(1000), r.Int())
	assert.True(t, r.Next())
	assert.Equal(t, uint(1100), r.Uint())
	assert.True(t, r.Next())
	assert.Equal(t, 12.34, r.Number())
	assert.False(t, r.Next())
}

func TestBufferClone(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutInt16(Put, 10, 100)
	buf.PutString(Put, 20, "hello")

	cloned := buf.Clone()
	assert.EqualValues(t, buf, cloned)
}

func TestPutNil(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutAny(PutTrue, 0, nil)

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.True(t, r.Bool())
}

func TestPutTime(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutAny(Put, 0, time.Unix(0, 0))

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.NotEmpty(t, r.Bytes())
}

func TestPutBitmap(t *testing.T) {
	buf := NewBuffer(0)
	buf.PutBitmap(Insert, 0, bitmap.Bitmap{0xff})

	r := NewReader()
	r.Seek(buf)
	assert.True(t, r.Next())
	assert.Equal(t, Insert, r.Type)
}

func TestBufferWriteTo(t *testing.T) {
	input := NewBuffer(0)
	input.Column = "test"
	input.PutInt16(Put, 10, 100)
	input.PutString(Put, 20, "hello")

	buffer := bytes.NewBuffer(nil)
	n, err := input.WriteTo(buffer)
	assert.NoError(t, err)
	assert.Equal(t, int64(buffer.Len()), n)
	assert.Equal(t, int64(36), n)

	output := NewBuffer(0)
	m, _ := output.ReadFrom(buffer)
	assert.Equal(t, int64(buffer.Len()), m)
	assert.Equal(t, input, output)
}

func TestBufferWriteToFailures(t *testing.T) {
	buf := NewBuffer(0)
	buf.Column = "test"
	buf.PutInt16(Put, 10, 100)
	buf.PutString(Put, 20, "hello")

	for size := 0; size < 30; size++ {
		output := &limitWriter{Limit: size}
		_, err := buf.WriteTo(output)
		assert.Error(t, err)
	}
}

func TestBufferReadFromFailures(t *testing.T) {
	input := NewBuffer(0)
	input.Column = "test"
	input.PutInt16(Put, 10, 100)
	input.PutString(Put, 20, "hello")

	buffer := bytes.NewBuffer(nil)
	n, err := input.WriteTo(buffer)
	assert.NoError(t, err)

	for size := 0; size < int(n)-1; size++ {
		output := NewBuffer(0)
		_, err := output.ReadFrom(bytes.NewReader(buffer.Bytes()[:size]))
		assert.Error(t, err)
	}
}

func FuzzBufferString(f *testing.F) {
	f.Add(uint32(1), "test")

	f.Fuzz(func(t *testing.T, i uint32, v string) {
		buf := NewBuffer(0)
		buf.PutString(Put, i, v)

		r := NewReader()
		r.Seek(buf)
		assert.True(t, r.Next())
		assert.Equal(t, v, r.String())
	})
}

func TestOpString(t *testing.T) {
	for i := 0; i < 255; i++ {
		assert.NotEmpty(t, OpType(i).String())
	}
}
