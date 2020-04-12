package common

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Buffer struct {
	*bytes.Buffer
}

func NewBuffer(buf []byte) Buffer {
	return Buffer{Buffer: bytes.NewBuffer(buf)}
}

func (buf Buffer) WriteUint16(n uint16) {
	tmp := make([]byte, 2)
	binary.BigEndian.PutUint16(tmp, n)
	buf.Write(tmp)
}

func (buf Buffer) WriteInt16(n int16) {
	buf.WriteUint16(uint16(n))
}

func (buf Buffer) ReadUint32() (uint32, error) {
	b := buf.Next(4)
	if len(b) != 4 {
		return 0, io.EOF
	}
	return binary.BigEndian.Uint32(b), nil
}

func (buf Buffer) ReadUint16() (uint16, error) {
	b := buf.Next(2)
	if len(b) != 2 {
		return 0, io.EOF
	}
	return binary.BigEndian.Uint16(b), nil
}

func (buf Buffer) ReadInt16() (int16, error) {
	n, err := buf.ReadUint16()
	return int16(n), err
}