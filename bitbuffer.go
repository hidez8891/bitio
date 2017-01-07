package bitio

import (
	"fmt"
	"io"
)

// BitReader is the interface bit/byte reading method
type BitReader interface {
	ReadBit(p *byte, bitSize int) (nBit int, err error)
	ReadBits(p []byte, bitSize int) (nBit int, err error)
	Read(p []byte) (nByte int, err error)
}

// BitWriter is the interface bit/byte writting method
type BitWriter interface {
	WriteBit(p byte, bitSize int) (nBit int, err error)
	WriteBits(p []byte, bitSize int) (nBit int, err error)
	Write(p []byte) (nByte int, err error)
}

////////////////////////////////////////////////////////////////////////////////

// NewBitReadBuffer returns BitReadBuffer
func NewBitReadBuffer(r io.Reader) *BitReadBuffer {
	return &BitReadBuffer{
		r:    r,
		buff: 0,
		left: 0,
	}
}

// BitReadBuffer is implemented by BitReader
type BitReadBuffer struct {
	r    io.Reader
	buff byte
	left int
}

// ReadBit reads single data (bitSize) and returns read size.
// If error happen, err will be set.
// Read data is saved right justified. (4bit = 0x0f)
func (obj *BitReadBuffer) ReadBit(b *byte, bitSize int) (nBit int, err error) {
	if b == nil {
		return 0, fmt.Errorf("bitio: argument *b is null pointer")
	}
	if bitSize > 8 {
		return 0, fmt.Errorf("bitio: ReadBit requires read size <= 8")
	}
	*b = 0

	if err = obj.tryRead(); err != nil {
		return 0, err
	}

	if obj.left < bitSize {
		bitSize -= obj.left

		*b = obj.buff >> uint(8-obj.left)
		*b <<= uint(bitSize)
		nBit += obj.left

		if err = obj.forceRead(); err != nil {
			return
		}
	}

	*b |= obj.buff >> uint(8-bitSize)
	nBit += bitSize

	obj.buff = obj.buff << uint(bitSize)
	obj.left -= bitSize

	return
}

// ReadBits reads data (bitSize) and returns read size.
// If error happen, err will be set.
// Read data is saved right justified. (12bit = 0x0f 0xff)
func (obj *BitReadBuffer) ReadBits(p []byte, bitSize int) (nBit int, err error) {
	var n int

	if len(p)*8 < bitSize {
		return 0, fmt.Errorf("bitio: argument p[] is %d bits, want %d bits", len(p)*8, bitSize)
	}

	byteSize := bitSize / 8
	bitSize %= 8

	if byteSize > 0 {
		if n, err = obj.Read(p[:byteSize]); err != nil {
			return
		}
		nBit += n
	}

	if bitSize > 0 {
		if n, err = obj.ReadBit(&p[byteSize], bitSize); err != nil {
			return
		}
		nBit += n

		p[byteSize] <<= uint(8 - bitSize)
		rightShift(p[0:byteSize+1], uint(8-bitSize))
	}

	return
}

// ReadBit reads data len(p) size and returns read size.
// If error happen, err will be set.
func (obj *BitReadBuffer) Read(p []byte) (nBit int, err error) {
	if err = obj.tryRead(); err != nil {
		return 0, err
	}
	if len(p) == 0 {
		return 0, err
	}

	n := obj.left
	p[0] = obj.buff >> uint(8-obj.left)
	obj.left = 0
	nBit += n

	for i := 1; i < len(p); i++ {
		if err = obj.forceRead(); err != nil {
			return
		}

		p[i] = obj.buff
		obj.left = 0
		nBit += 8
	}

	if n < 8 {
		if err = obj.forceRead(); err != nil {
			return
		}
		nn := 8 - n

		leftShift(p, uint(nn))
		p[len(p)-1] |= obj.buff >> uint(8-nn)
		nBit += nn

		obj.buff <<= uint(nn)
		obj.left -= nn
	}

	return
}

// tryRead reads 1 byte data if obj.buff is empty.
// If error happen, returns err.
func (obj *BitReadBuffer) tryRead() error {
	if obj.left == 0 {
		return obj.forceRead()
	}
	return nil
}

// forceRead reads 1 byte data.
// If error happen, returns err.
func (obj *BitReadBuffer) forceRead() error {
	b := make([]byte, 1)

	if _, err := obj.r.Read(b); err != nil {
		return err
	}

	obj.buff = b[0]
	obj.left = 8
	return nil
}
