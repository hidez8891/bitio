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
	Flush() error
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
// Input data is stored left justified. (4bit = 0xf0)
// Output data is stored right justified. (4bit = 0x0f)
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
// Input data is stored left justified. (12bit = 0xff 0xf0)
// Output data is stored right justified. (12bit = 0x0f 0xff)
func (obj *BitReadBuffer) ReadBits(p []byte, bitSize int) (nBit int, err error) {
	if len(p)*8 < bitSize {
		return 0, fmt.Errorf("bitio: argument p[] is %d bits, want %d bits", len(p)*8, bitSize)
	}

	if bitSize <= obj.left {
		p[len(p)-1] = obj.buff >> uint(8-bitSize)
		nBit = bitSize

		obj.buff <<= uint(bitSize)
		obj.left -= bitSize
		return
	}

	readBit := bitSize - obj.left
	readByte := (readBit + 7) / 8
	pp := make([]byte, readByte)
	if _, err = obj.r.Read(pp); err != nil {
		return
	}

	copy(p, pp)
	rightShift(p, uint(obj.left))
	p[0] |= obj.buff
	rightShift(p, uint(8*len(p)-bitSize))

	obj.buff = 0
	obj.left = 0
	if readBit%8 > 0 {
		obj.left = 8 - readBit%8
		obj.buff = pp[len(pp)-1] << uint(8-obj.left)
	}

	nBit = bitSize
	return
}

// Read reads data len(p) size and returns read size.
// If error happen, err will be set.
func (obj *BitReadBuffer) Read(p []byte) (nByte int, err error) {
	nBit, err := obj.ReadBits(p, len(p)*8)
	return nBit / 8, err
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

////////////////////////////////////////////////////////////////////////////////

// NewBitWriteBuffer returns BitWriteBuffer
func NewBitWriteBuffer(w io.Writer) *BitWriteBuffer {
	return &BitWriteBuffer{
		w:    w,
		buff: 0,
		left: 0,
	}
}

// BitWriteBuffer is implemented by BitWriteer
type BitWriteBuffer struct {
	w    io.Writer
	buff byte
	left int
}

// WriteBit writes single data (bitSize) and returns write size.
// If error happen, err will be set.
// Input data is stored left justified. (4bit = 0x0f)
// Output data is stored right justified. (4bit = 0xf0)
func (obj *BitWriteBuffer) WriteBit(p byte, bitSize int) (nBit int, err error) {
	p <<= uint(8 - bitSize)

	if obj.left+bitSize > 8 {
		n := 8 - obj.left
		obj.buff <<= uint(n)
		obj.buff |= p >> uint(8-n)
		obj.left += n

		if err = obj.forceWrite(); err != nil {
			return
		}
		nBit += n

		p <<= uint(n)
		bitSize -= n
	}

	obj.buff <<= uint(bitSize)
	obj.buff |= p >> uint(8-bitSize)
	obj.left += bitSize

	if err = obj.tryWrite(); err != nil {
		return
	}
	nBit += bitSize

	return
}

// WriteBits writes data (bitSize) and returns write size.
// If error happen, err will be set.
// Input data is stored right justified. (12bit = 0x0f 0xff)
// Output data is stored left justified. (12bit = 0xff 0xf0)
func (obj *BitWriteBuffer) WriteBits(p []byte, bitSize int) (nBit int, err error) {
	if len(p)*8 < bitSize {
		return 0, fmt.Errorf("bitio: argument p[] is %d bits, want %d bits", len(p)*8, bitSize)
	}

	byteSize := (bitSize + 7) / 8
	buf := make([]byte, byteSize)
	copy(buf, p[len(p)-byteSize:])

	shift := uint(bitSize) % 8
	if shift > 0 {
		leftShift(buf, 8-shift)
		buf[len(buf)-1] >>= 8 - shift
	}

	for _, b := range buf {
		var n int

		n, err = obj.WriteBit(b, min(8, bitSize))
		if err != nil {
			return
		}

		bitSize -= n
		nBit += n
	}

	return
}

// Write writes data len(p) size and returns write size.
// If error happen, err will be set.
func (obj *BitWriteBuffer) Write(p []byte) (nByte int, err error) {
	nBit, err := obj.WriteBits(p, len(p)*8)
	return nBit / 8, err
}

// Flush writes data if obj.buff is not empty.
// If error happen, err will be set.
func (obj *BitWriteBuffer) Flush() error {
	return obj.forceWrite()
}

// tryWrite writes 1 byte data if obj.buff is full.
// If error happen, returns err.
func (obj *BitWriteBuffer) tryWrite() error {
	if obj.left == 8 {
		return obj.forceWrite()
	}
	return nil
}

// forceWrite writes 1 byte data. (0 right padding)
// If error happen, returns err.
func (obj *BitWriteBuffer) forceWrite() error {
	if obj.left == 0 {
		return nil
	}

	obj.buff <<= uint(8 - obj.left)
	b := make([]byte, 1)
	b[0] = obj.buff

	if _, err := obj.w.Write(b); err != nil {
		return err
	}

	obj.buff = 0
	obj.left = 0
	return nil
}
