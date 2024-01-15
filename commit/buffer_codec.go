// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package commit

import (
	"encoding/binary"
	"io"
	"reflect"
	"unsafe"

	"github.com/kelindar/iostream"
)

// --------------------------- WriteTo ----------------------------

// WriteTo writes data to w until there's no more data to write or when an error occurs. The return
// value n is the number of bytes written. Any error encountered during the write is also returned.
func (b *Buffer) WriteTo(dst io.Writer) (int64, error) {
	w := iostream.NewWriter(dst)
	if err := w.WriteString(b.Column); err != nil {
		return w.Offset(), err
	}

	if err := w.WriteInt32(b.last); err != nil {
		return w.Offset(), err
	}

	var temp [12]byte
	if err := w.WriteRange(len(b.chunks), func(i int, w *iostream.Writer) error {
		v := b.chunks[i]
		binary.BigEndian.PutUint32(temp[0:4], uint32(v.Chunk))
		binary.BigEndian.PutUint32(temp[4:8], v.Start)
		binary.BigEndian.PutUint32(temp[8:12], v.Value)
		_, err := w.Write(temp[:])
		return err
	}); err != nil {
		return w.Offset(), err
	}

	err := w.WriteBytes(b.buffer)
	return w.Offset(), err
}

// --------------------------- ReadFrom ----------------------------

// ReadFrom reads data from r until EOF or error. The return value n is the number of
// bytes read. Any error except EOF encountered during the read is also returned.
func (b *Buffer) ReadFrom(src io.Reader) (int64, error) {
	r := iostream.NewReader(src)
	var err error
	if b.Column, err = r.ReadString(); err != nil {
		return r.Offset(), err
	}

	if b.last, err = r.ReadInt32(); err != nil {
		return r.Offset(), err
	}

	if b.chunks, err = readChunksFrom(r); err != nil {
		return r.Offset(), err
	}

	if b.buffer, err = r.ReadBytes(); err != nil {
		return r.Offset(), err
	}

	if len(b.chunks) > 0 {
		last := b.chunks[len(b.chunks)-1]
		b.chunk = last.Chunk
	}

	return r.Offset(), nil
}

// readChunksFrom reads the list of chunks from the reader
func readChunksFrom(r *iostream.Reader) ([]header, error) {
	size, err := r.ReadUvarint()
	if err != nil {
		return nil, err
	}

	v := make([]header, size)
	var temp [12]byte
	for i := 0; i < int(size); i++ {
		if _, err := io.ReadFull(r, temp[:]); err != nil {
			return nil, err
		}

		v[i].Chunk = Chunk(binary.BigEndian.Uint32(temp[0:4]))
		v[i].Start = binary.BigEndian.Uint32(temp[4:8])
		v[i].Value = binary.BigEndian.Uint32(temp[8:12])
	}
	return v, nil
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
