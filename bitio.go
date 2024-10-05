package bitio

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/exp/constraints"
)

// ByteOrder indicates the endianness of binary data.
type ByteOrder bool

const (
	BigEndian    ByteOrder = false
	LittleEndian ByteOrder = true
)

// Read read from BitReader and convert to T type value.
// Returns error if reading from reader fails or number of read bits is less than requested.
func Read[T constraints.Integer](br BitReader, nBit int, order ByteOrder, dst *T) error {
	tsize := int(unsafe.Sizeof(*dst))
	if tsize*8 < nBit {
		return fmt.Errorf("read size %d bit exceeds %T type size %d bit", nBit, *dst, tsize*8)
	}
	if tsize > 8 {
		return fmt.Errorf("unsupport %T type", *dst)
	}

	buf := make([]byte, 8)
	if n, err := br.ReadBits(buf, nBit); err != nil {
		return err
	} else if n != nBit {
		return fmt.Errorf("insufficient size of read, want %d bit, read %d bit", nBit, n)
	}

	var value uint64
	if order == LittleEndian {
		// little endian
		// 12bit: 0x123 = 0x*****231 -> 0x****2301 -> 0x2301****
		if nBit%8 > 0 {
			leftShift(buf, uint(8-nBit%8))
			buf[len(buf)-1] >>= uint(8 - nBit%8)
		}
		leftShift(buf, uint(8*(8-(nBit+7)/8)))
		value = binary.LittleEndian.Uint64(buf)
	} else {
		// big endian (no shift)
		// 12bit: 0x123 = 0x*****123
		value = binary.BigEndian.Uint64(buf)
	}

	*dst = T(value)
	return nil
}

// ReadSlice read from BitReader and convert to T type slice.
// Return error if element read failed.
func ReadSlice[T constraints.Integer](br BitReader, elemBit int, order ByteOrder, dst []T) error {
	length := len(dst)
	for i := 0; i < length; i++ {
		if err := Read(br, elemBit, order, &dst[i]); err != nil {
			return err
		}
	}
	return nil
}

// Write write T type value to BitWriter as specified number of bits.
// Return error if writing to writer fails or number of write bits is exceeds T size.
func Write[T constraints.Integer](bw BitWriter, nBit int, order ByteOrder, src T) error {
	tsize := int(unsafe.Sizeof(src))
	if tsize*8 < nBit {
		return fmt.Errorf("write size %d bit exceeds %T type size %d bit", nBit, src, tsize*8)
	}
	if tsize > 8 {
		return fmt.Errorf("unsupport %T type", src)
	}

	buf := make([]byte, 8)
	if order == LittleEndian {
		// little endian
		// 12bit: 0x123 = 0x2301**** -> 0x****2301 -> 0x*****231
		binary.LittleEndian.PutUint64(buf, uint64(src))
		rightShift(buf, 8*uint(8-(nBit+7)/8))
		if nBit%8 > 0 {
			buf[len(buf)-1] <<= uint(8 - nBit%8)
			rightShift(buf, uint(8-nBit%8))
		}
	} else {
		// big endian (no shift)
		// 12bit: 0x123 = 0x****0123
		binary.BigEndian.PutUint64(buf, uint64(src))
	}

	if n, err := bw.WriteBits(buf, nBit); err != nil {
		return err
	} else if n != nBit {
		return fmt.Errorf("insufficient size of write, want %d bit, write %d bit", nBit, n)
	}

	return nil
}

// WriteSlice writes a slice of T to BitWriter.
// Return error if element write fails.
func WriteSlice[T constraints.Integer](bw BitWriter, elemBit int, order ByteOrder, src []T) error {
	length := len(src)
	for i := 0; i < length; i++ {
		if err := Write(bw, elemBit, order, src[i]); err != nil {
			return err
		}
	}
	return nil
}
