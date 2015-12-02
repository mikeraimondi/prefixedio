package prefixedio

import (
	"encoding/binary"
	"fmt"
	"io"
)

const maxLen = 1048576

// PrefixedWrite writes the size of the bytes passed, then the bytes
func PrefixedWrite(w io.Writer, bytes []byte) (n int, err error) {
	if err = binary.Write(w, binary.BigEndian, uint64(len(bytes))); err != nil {
		return
	}
	return w.Write(bytes)
}

// PrefixedBuf is a size-prefixed buffer
type PrefixedBuf struct {
	size int64
	buf  []byte
}

// ReadFrom reads the size prefix (s) from rd, then overwrites the buffer with s bytes from rd
func (p *PrefixedBuf) ReadFrom(rd io.Reader) (n int64, err error) {
	if diff := 8 - len(p.buf); diff > 0 {
		p.buf = append(p.buf, make([]byte, diff)...)
	}
	_, err = rd.Read(p.buf[:8])
	if err != nil {
		return
	}
	p.size = int64(binary.BigEndian.Uint64(p.buf[:8])) //TODO ick
	if p.size == 0 {
		return
	}
	if p.size > maxLen {
		err = fmt.Errorf("Message too large at %v bytes", p.size)
		return
	}
	if diff := p.size - int64(len(p.buf)); diff > 0 {
		p.buf = append(p.buf, make([]byte, diff)...)
	}
	var nInt int
	nInt, err = rd.Read(p.buf[:p.size])
	n = int64(nInt)
	return
}

// Bytes returns the bytes from the buffer
func (p *PrefixedBuf) Bytes() []byte {
	return p.buf[:p.size]
}
