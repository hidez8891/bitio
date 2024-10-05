package bitio_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hidez8891/bitio"
	"golang.org/x/exp/constraints"
)

type testDataRW[T constraints.Integer] struct {
	value T
	buf   []byte
	nBit  int
	order bitio.ByteOrder
}

func testRead[T constraints.Integer](t *testing.T, tests []testDataRW[T]) {
	t.Helper()

	for i, tt := range tests {
		br := bitio.NewBitReadBuffer(bytes.NewReader(tt.buf))

		var dst T
		err := bitio.Read(br, tt.nBit, tt.order, &dst)

		if err != nil {
			t.Fatalf("Read[%T] read fail:%v [testcase-%d]", dst, err, i)
		}
		if dst != tt.value {
			t.Fatalf("Read[%T] read %x, want %x [testcase-%d]", dst, dst, tt.value, i)
		}
	}
}

func TestRead(t *testing.T) {
	testRead(t, []testDataRW[int8]{
		{
			buf:   []byte{0xab},
			nBit:  4,
			order: bitio.LittleEndian,
			value: 0x0a,
		},
		{
			buf:   []byte{0xab},
			nBit:  4,
			order: bitio.BigEndian,
			value: 0x0a,
		},
		{
			buf:   []byte{0xab},
			nBit:  8,
			order: bitio.LittleEndian,
			value: -int8(^(uint8(0xab) - 1)),
		},
		{
			buf:   []byte{0xab},
			nBit:  8,
			order: bitio.BigEndian,
			value: -int8(^(uint8(0xab) - 1)),
		},
	})

	testRead(t, []testDataRW[uint8]{
		{
			buf:   []byte{0xab},
			nBit:  8,
			order: bitio.LittleEndian,
			value: 0xab,
		},
		{
			buf:   []byte{0xab},
			nBit:  8,
			order: bitio.BigEndian,
			value: 0xab,
		},
	})

	testRead(t, []testDataRW[int16]{
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  12,
			order: bitio.LittleEndian,
			value: 0x0cab,
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  12,
			order: bitio.BigEndian,
			value: 0x0abc,
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  16,
			order: bitio.LittleEndian,
			value: -int16(^(uint16(0xcdab) - 1)),
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  16,
			order: bitio.BigEndian,
			value: -int16(^(uint16(0xabcd) - 1)),
		},
	})

	testRead(t, []testDataRW[uint16]{
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  16,
			order: bitio.LittleEndian,
			value: 0xcdab,
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  16,
			order: bitio.BigEndian,
			value: 0xabcd,
		},
	})

	testRead(t, []testDataRW[int32]{
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  28,
			order: bitio.LittleEndian,
			value: 0x0c3412ab,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  28,
			order: bitio.BigEndian,
			value: 0x0ab1234c,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  32,
			order: bitio.LittleEndian,
			value: -int32(^(uint32(0xcd3412ab) - 1)),
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  32,
			order: bitio.BigEndian,
			value: -int32(^(uint32(0xab1234cd) - 1)),
		},
	})

	testRead(t, []testDataRW[uint32]{
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  32,
			order: bitio.LittleEndian,
			value: 0xcd3412ab,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
			nBit:  32,
			order: bitio.BigEndian,
			value: 0xab1234cd,
		},
	})

	testRead(t, []testDataRW[int64]{
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  60,
			order: bitio.LittleEndian,
			value: 0x0cbc9a78563412ab,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  60,
			order: bitio.BigEndian,
			value: 0x0ab123456789abcc,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  64,
			order: bitio.LittleEndian,
			value: -int64(^(uint64(0xcdbc9a78563412ab) - 1)),
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  64,
			order: bitio.BigEndian,
			value: -int64(^(uint64(0xab123456789abccd) - 1)),
		},
	})

	testRead(t, []testDataRW[uint64]{
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  64,
			order: bitio.LittleEndian,
			value: 0xcdbc9a78563412ab,
		},
		{
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
			nBit:  64,
			order: bitio.BigEndian,
			value: 0xab123456789abccd,
		},
	})
}

func testWrite[T constraints.Integer](t *testing.T, tests []testDataRW[T]) {
	t.Helper()

	for i, tt := range tests {
		b := new(bytes.Buffer)
		bw := bitio.NewBitWriteBuffer(b)

		err := bitio.Write(bw, tt.nBit, tt.order, tt.value)
		if err != nil {
			t.Fatalf("Write[%T] write fail:%v [testcase-%d]", tt.value, err, i)
		}

		err = bw.Flush()
		if err != nil {
			t.Fatalf("Write[%T] flush fail:%v [testcase-%d]", tt.value, err, i)
		}

		if reflect.DeepEqual(b.Bytes(), tt.buf) == false {
			t.Fatalf("Write[%T] write %#v, want %#v [testcase-%d]", tt.value, b.Bytes(), tt.buf, i)
		}
	}
}

func TestWrite(t *testing.T) {
	testWrite(t, []testDataRW[int8]{
		{
			value: -int8(^(uint8(0xab) - 1)),
			nBit:  4,
			order: bitio.LittleEndian,
			buf:   []byte{0xb0},
		},
		{
			value: -int8(^(uint8(0xab) - 1)),
			nBit:  4,
			order: bitio.BigEndian,
			buf:   []byte{0xb0},
		},
		{
			value: -int8(^(uint8(0xab) - 1)),
			nBit:  8,
			order: bitio.LittleEndian,
			buf:   []byte{0xab},
		},
		{
			value: -int8(^(uint8(0xab) - 1)),
			nBit:  8,
			order: bitio.BigEndian,
			buf:   []byte{0xab},
		},
	})

	testWrite(t, []testDataRW[uint8]{
		{
			value: 0xab,
			nBit:  8,
			order: bitio.LittleEndian,
			buf:   []byte{0xab},
		},
		{
			value: 0xab,
			nBit:  8,
			order: bitio.BigEndian,
			buf:   []byte{0xab},
		},
	})

	testWrite(t, []testDataRW[int16]{
		{
			value: -int16(^(uint16(0xabcd) - 1)),
			nBit:  12,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xb0},
		},
		{
			value: -int16(^(uint16(0xabcd) - 1)),
			nBit:  12,
			order: bitio.BigEndian,
			buf:   []byte{0xbc, 0xd0},
		},
		{
			value: -int16(^(uint16(0xabcd) - 1)),
			nBit:  16,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xab},
		},
		{
			value: -int16(^(uint16(0xabcd) - 1)),
			nBit:  16,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0xcd},
		},
	})

	testWrite(t, []testDataRW[uint16]{
		{
			value: 0xabcd,
			nBit:  16,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xab},
		},
		{
			value: 0xabcd,
			nBit:  16,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0xcd},
		},
	})

	testWrite(t, []testDataRW[int32]{
		{
			value: -int32(^(uint32(0xab1234cd) - 1)),
			nBit:  28,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0x34, 0x12, 0xb0},
		},
		{
			value: -int32(^(uint32(0xab1234cd) - 1)),
			nBit:  28,
			order: bitio.BigEndian,
			buf:   []byte{0xb1, 0x23, 0x4c, 0xd0},
		},
		{
			value: -int32(^(uint32(0xab1234cd) - 1)),
			nBit:  32,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0x34, 0x12, 0xab},
		},
		{
			value: -int32(^(uint32(0xab1234cd) - 1)),
			nBit:  32,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
		},
	})

	testWrite(t, []testDataRW[uint32]{
		{
			value: 0xab1234cd,
			nBit:  32,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0x34, 0x12, 0xab},
		},
		{
			value: 0xab1234cd,
			nBit:  32,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0x12, 0x34, 0xcd},
		},
	})

	testWrite(t, []testDataRW[int64]{
		{
			value: -int64(^(uint64(0xab123456789abccd) - 1)),
			nBit:  60,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xbc, 0x9a, 0x78, 0x56, 0x34, 0x12, 0xb0},
		},
		{
			value: -int64(^(uint64(0xab123456789abccd) - 1)),
			nBit:  60,
			order: bitio.BigEndian,
			buf:   []byte{0xb1, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcc, 0xd0},
		},
		{
			value: -int64(^(uint64(0xab123456789abccd) - 1)),
			nBit:  64,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xbc, 0x9a, 0x78, 0x56, 0x34, 0x12, 0xab},
		},
		{
			value: -int64(^(uint64(0xab123456789abccd) - 1)),
			nBit:  64,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
		},
	})

	testWrite(t, []testDataRW[uint64]{
		{
			value: 0xab123456789abccd,
			nBit:  64,
			order: bitio.LittleEndian,
			buf:   []byte{0xcd, 0xbc, 0x9a, 0x78, 0x56, 0x34, 0x12, 0xab},
		},
		{
			value: 0xab123456789abccd,
			nBit:  64,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xcd},
		},
	})
}

type testDataSliceRW[T constraints.Integer] struct {
	value []T
	buf   []byte
	nBit  int
	order bitio.ByteOrder
}

func testReadSlice[T constraints.Integer](t *testing.T, tests []testDataSliceRW[T]) {
	t.Helper()

	for i, tt := range tests {
		br := bitio.NewBitReadBuffer(bytes.NewReader(tt.buf))

		dst := make([]T, len(tt.value))
		err := bitio.ReadSlice(br, tt.nBit, tt.order, dst)

		if err != nil {
			t.Fatalf("ReadSlice[%T] read fail:%v [testcase-%d]", dst, err, i)
		}
		if reflect.DeepEqual(dst, tt.value) == false {
			t.Fatalf("ReadSlice[%T] read %#v, want %#v [testcase-%d]", dst, dst, tt.value, i)
		}
	}
}

func TestReadSlice(t *testing.T) {
	testReadSlice(t, []testDataSliceRW[uint8]{
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  4,
			order: bitio.LittleEndian,
			value: []uint8{0x0a, 0x0b, 0x0c, 0x0d},
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  4,
			order: bitio.BigEndian,
			value: []uint8{0x0a, 0x0b, 0x0c, 0x0d},
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  8,
			order: bitio.LittleEndian,
			value: []uint8{0xab, 0xcd},
		},
		{
			buf:   []byte{0xab, 0xcd},
			nBit:  8,
			order: bitio.BigEndian,
			value: []uint8{0xab, 0xcd},
		},
	})

	testReadSlice(t, []testDataSliceRW[uint16]{
		{
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
			nBit:  12,
			order: bitio.LittleEndian,
			value: []uint16{0x312, 0x645, 0x978, 0xb1a},
		},
		{
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
			nBit:  12,
			order: bitio.BigEndian,
			value: []uint16{0x123, 0x456, 0x789, 0x1ab},
		},
		{
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
			nBit:  16,
			order: bitio.LittleEndian,
			value: []uint16{0x3412, 0x7856, 0xab91},
		},
		{
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
			nBit:  16,
			order: bitio.BigEndian,
			value: []uint16{0x1234, 0x5678, 0x91ab},
		},
	})
}

func testWriteSlice[T constraints.Integer](t *testing.T, tests []testDataSliceRW[T]) {
	t.Helper()

	for i, tt := range tests {
		b := new(bytes.Buffer)
		bw := bitio.NewBitWriteBuffer(b)

		err := bitio.WriteSlice(bw, tt.nBit, tt.order, tt.value)
		if err != nil {
			t.Fatalf("WriteSlice[%T] write fail:%v [testcase-%d]", tt.value, err, i)
		}

		err = bw.Flush()
		if err != nil {
			t.Fatalf("WriteSlice[%T] flush fail:%v [testcase-%d]", tt.value, err, i)
		}

		if reflect.DeepEqual(b.Bytes(), tt.buf) == false {
			t.Fatalf("WriteSlice[%T] write %#v, want %#v [testcase-%d]", tt.value, b.Bytes(), tt.buf, i)
		}
	}
}

func TestWriteSlice(t *testing.T) {
	testWriteSlice(t, []testDataSliceRW[uint8]{
		{
			value: []uint8{0x0a, 0x0b, 0x0c, 0x0d},
			nBit:  4,
			order: bitio.LittleEndian,
			buf:   []byte{0xab, 0xcd},
		},
		{
			value: []uint8{0x0a, 0x0b, 0x0c, 0x0d},
			nBit:  4,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0xcd},
		},
		{
			value: []uint8{0xab, 0xcd},
			nBit:  8,
			order: bitio.LittleEndian,
			buf:   []byte{0xab, 0xcd},
		},
		{
			value: []uint8{0xab, 0xcd},
			nBit:  8,
			order: bitio.BigEndian,
			buf:   []byte{0xab, 0xcd},
		},
	})

	testWriteSlice(t, []testDataSliceRW[uint16]{
		{
			value: []uint16{0x312, 0x645, 0x978, 0xb1a},
			nBit:  12,
			order: bitio.LittleEndian,
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
		},
		{
			value: []uint16{0x123, 0x456, 0x789, 0x1ab},
			nBit:  12,
			order: bitio.BigEndian,
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
		},
		{
			value: []uint16{0x3412, 0x7856, 0xab91},
			nBit:  16,
			order: bitio.LittleEndian,
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
		},
		{
			value: []uint16{0x1234, 0x5678, 0x91ab},
			nBit:  16,
			order: bitio.BigEndian,
			buf:   []byte{0x12, 0x34, 0x56, 0x78, 0x91, 0xab},
		},
	})
}
