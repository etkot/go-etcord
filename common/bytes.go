package common

import (
	"bytes"
	"encoding/binary"
	"io"
)

type Buffer struct {
	*bytes.Buffer
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