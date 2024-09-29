package bitio_test

import (
	"bytes"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/hidez8891/bitio"
)

func TestBit_interface(t *testing.T) {
	// Only compile test

	r := &bitio.BitReadBuffer{}
	var r1 bitio.BitReader = r
	_ = r1
	var r2 io.Reader = r
	_ = r2

	w := &bitio.BitWriteBuffer{}
	var w1 bitio.BitWriter = w
	_ = w1
	var w2 io.Writer = w
	_ = w2
}

func TestBitReadBuffer_ReadBit(t *testing.T) {
	var tests = []struct {
		data []byte
		bits int
		exp  byte
	}{
		// read bit (full bit)
		{[]byte{0xff}, 1, 0x01},
		{[]byte{0xff}, 4, 0x0f},
		{[]byte{0xff}, 7, 0x7f},
		{[]byte{0xff}, 8, 0xff},
		// read bit
		{[]byte{0xab}, 1, 0x01},
		{[]byte{0xab}, 4, 0x0a},
		{[]byte{0xab}, 7, 0x55},
		{[]byte{0xab}, 8, 0xab},
	}

	for _, tt := range tests {
		var n int
		var b byte
		var err error

		r := bitio.NewBitReadBuffer(bytes.NewReader(tt.data))
		if n, err = r.ReadBit(&b, tt.bits); err != nil {
			t.Fatalf("ReadBit happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("ReadBit read size %d, want %d", n, tt.bits)
		}

		if b != tt.exp {
			t.Fatalf("ReadBit read data %#v, want %#v", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_ReadBit_nbit(t *testing.T) {
	// read number N from N bit
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101"
	data := binaryToByteArray(str)

	r := bitio.NewBitReadBuffer(bytes.NewReader(data))
	for i := 1; i <= 5; i++ {
		var n int
		var b byte
		var err error

		if n, err = r.ReadBit(&b, i); err != nil {
			t.Fatalf("ReadBit happen error %v", err)
		}

		if n != i {
			t.Fatalf("ReadBit read size %d, want %d", n, i)
		}

		if b != byte(i) {
			t.Fatalf("ReadBit read data %x, want %x", b, byte(i))
		}
	}
}

func TestBitReadBuffer_ReadBits(t *testing.T) {
	var tests = []struct {
		data []byte
		bits int
		exp  []byte
	}{
		// read bit (full bit)
		{[]byte{0xff, 0xff}, 1, []byte{0x01}},
		{[]byte{0xff, 0xff}, 7, []byte{0x7f}},
		{[]byte{0xff, 0xff}, 8, []byte{0xff}},
		{[]byte{0xff, 0xff}, 9, []byte{0x01, 0xff}},
		{[]byte{0xff, 0xff}, 15, []byte{0x7f, 0xff}},
		{[]byte{0xff, 0xff}, 16, []byte{0xff, 0xff}},
		// read bit
		{[]byte{0xab, 0xcd}, 1, []byte{0x01}},
		{[]byte{0xab, 0xcd}, 7, []byte{0x55}},
		{[]byte{0xab, 0xcd}, 8, []byte{0xab}},
		{[]byte{0xab, 0xcd}, 9, []byte{0x01, 0x57}},
		{[]byte{0xab, 0xcd}, 15, []byte{0x55, 0xe6}},
		{[]byte{0xab, 0xcd}, 16, []byte{0xab, 0xcd}},
		// read bit (over size buffer)
		{[]byte{0xab, 0xcd}, 1, []byte{0x00, 0x00, 0x01}},
		{[]byte{0xab, 0xcd}, 7, []byte{0x00, 0x00, 0x55}},
		{[]byte{0xab, 0xcd}, 8, []byte{0x00, 0x00, 0xab}},
		{[]byte{0xab, 0xcd}, 9, []byte{0x00, 0x01, 0x57}},
		{[]byte{0xab, 0xcd}, 15, []byte{0x00, 0x55, 0xe6}},
		{[]byte{0xab, 0xcd}, 16, []byte{0x00, 0xab, 0xcd}},
	}

	for _, tt := range tests {
		var n int
		var b []byte
		var err error

		r := bitio.NewBitReadBuffer(bytes.NewReader(tt.data))
		b = make([]byte, len(tt.exp))

		if n, err = r.ReadBits(b, tt.bits); err != nil {
			t.Fatalf("ReadBits happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("ReadBits read size %d, want %d", n, tt.bits)
		}

		if reflect.DeepEqual(b, tt.exp) == false {
			t.Fatalf("ReadBits read data %#v, want %#v", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_ReadBits_nbit(t *testing.T) {
	// read number N from N bit
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101" +
		"000110" +
		"0000111" +
		"00001000" +
		"000001001" +
		"0000001010"
	data := binaryToByteArray(str)

	r := bitio.NewBitReadBuffer(bytes.NewReader(data))
	for i := 1; i <= 10; i++ {
		var n int
		var err error

		b := make([]byte, 2)
		if n, err = r.ReadBits(b, i); err != nil {
			t.Fatalf("ReadBits happen error %v", err)
		}

		if n != i {
			t.Fatalf("ReadBits read size %d, want %d", n, i)
		}

		exp := []byte{0x00, 0x00}
		exp[1] = byte(i)

		if reflect.DeepEqual(b, exp) == false {
			t.Fatalf("ReadBits read data %#v, want %#v", b, exp)
		}
	}
}

func TestBitReadBuffer_Read(t *testing.T) {
	var tests = []struct {
		data []byte
		size int
		exp  []byte
	}{
		// read byte (full byte)
		{[]byte{0xff, 0xff}, 1, []byte{0xff}},
		{[]byte{0xff, 0xff}, 2, []byte{0xff, 0xff}},
		// read byte
		{[]byte{0x12, 0x34}, 1, []byte{0x12}},
		{[]byte{0x12, 0x34}, 2, []byte{0x12, 0x34}},
	}

	for _, tt := range tests {
		var n int
		var b []byte
		var err error

		r := bitio.NewBitReadBuffer(bytes.NewReader(tt.data))
		b = make([]byte, tt.size)

		if n, err = r.Read(b); err != nil {
			t.Fatalf("Read happen error %v", err)
		}

		if n != tt.size {
			t.Fatalf("Read read size %d, want %d", n, tt.size)
		}

		if reflect.DeepEqual(b, tt.exp) == false {
			t.Fatalf("Read read data %#v, want %#v", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_Read_Combination(t *testing.T) {
	datas := [][]byte{
		binaryToByteArray("" +
			"01" +
			"0011_0110" +
			"111" +
			"01_0111_0011" +
			"01_0111_0011" +
			"0011_0110"),
		{0x97, 0x97},
	}

	var tests = [][]struct {
		bits int
		exp  []byte
	}{
		{
			{2, []byte{0x01}},
			{8, []byte{0x36}},
			{3, []byte{0x07}},
			{10, []byte{0x01, 0x73}},
			{10, []byte{0x01, 0x73}},
			{8, []byte{0x36}},
		},
		{
			{1, []byte{0x01}},
			{14, []byte{0x0b, 0xcb}},
			{1, []byte{0x01}},
		},
	}

	for i := 0; i < len(tests); i++ {
		data := datas[i]
		test := tests[i]
		r := bitio.NewBitReadBuffer(bytes.NewReader(data))

		for _, tt := range test {
			var n int
			var b []byte
			var err error

			b = make([]byte, (tt.bits+7)/8)

			if tt.bits < 8 {
				// test ReadBit
				if n, err = r.ReadBit(&b[0], tt.bits); err != nil {
					t.Fatalf("BitReadBuffer happen error %v", err)
				}
			} else if tt.bits == 8 {
				// test Read
				if n, err = r.Read(b); err != nil {
					t.Fatalf("BitReadBuffer happen error %v", err)
				}
				n *= 8
			} else {
				// test ReadBits
				if n, err = r.ReadBits(b, tt.bits); err != nil {
					t.Fatalf("BitReadBuffer happen error %v", err)
				}
			}

			if n != tt.bits {
				t.Fatalf("BitReadBuffer read size %d, want %d", n, tt.bits)
			}

			if reflect.DeepEqual(b, tt.exp) == false {
				t.Fatalf("BitReadBuffer read data %#v, want %#v", b, tt.exp)
			}
		}
	}
}

func BenchmarkBitReadBuffer_Read_Aligned_32b(b *testing.B) {
	readAligned(b, 2<<5)
}

func BenchmarkBitReadBuffer_Read_UnAligned_32b(b *testing.B) {
	readUnAligned(b, 2<<5)
}

func BenchmarkBitReadBuffer_Read_Aligned_1024b(b *testing.B) {
	readAligned(b, 2<<10)
}

func BenchmarkBitReadBuffer_Read_UnAligned_1024b(b *testing.B) {
	readUnAligned(b, 2<<10)
}

func BenchmarkBitReadBuffer_Read_Aligned_65536b(b *testing.B) {
	readAligned(b, 2<<16)
}

func BenchmarkBitReadBuffer_Read_UnAligned_65536b(b *testing.B) {
	readUnAligned(b, 2<<16)
}

func readAligned(b *testing.B, bufSize int) {
	r := bitio.NewBitReadBuffer(&Infinity{})
	p := make([]byte, bufSize)

	b.SetBytes(int64(bufSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(p)
	}
}

func readUnAligned(b *testing.B, bufSize int) {
	r := bitio.NewBitReadBuffer(&Infinity{})
	p := make([]byte, bufSize)

	// put off align by 1bit
	r.ReadBit(&p[0], 1)

	b.SetBytes(int64(bufSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(p)
	}
}

////////////////////////////////////////////////////////////////////////////////

func TestBitWriteBuffer_WriteBit(t *testing.T) {
	var tests = []struct {
		data byte
		bits int
		exp  []byte
	}{
		// write bit (full bit)
		{0xff, 1, []byte{0x80}},
		{0xff, 4, []byte{0xf0}},
		{0xff, 7, []byte{0xfe}},
		{0xff, 8, []byte{0xff}},
		// write bit
		{0xab, 1, []byte{0x80}},
		{0xab, 4, []byte{0xb0}},
		{0xab, 7, []byte{0x56}},
		{0xab, 8, []byte{0xab}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitWriteBuffer(b)
		if n, err = w.WriteBit(tt.data, tt.bits); err != nil {
			t.Fatalf("WriteBit happen error %v", err)
		}
		if err = w.Flush(); err != nil {
			t.Fatalf("WriteBit happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("WriteBit write size %d, want %d", n, tt.bits)
		}

		if reflect.DeepEqual(b.Bytes(), tt.exp) == false {
			t.Fatalf("WriteBit write data %#v, want %#v", b.Bytes(), tt.exp)
		}
	}
}

func TestBitWriteBuffer_WriteBit_nbit(t *testing.T) {
	// write number N to N bit
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101"
	exp := binaryToByteArray(str)

	b := bytes.NewBuffer([]byte{})
	w := bitio.NewBitWriteBuffer(b)
	for i := 1; i <= 5; i++ {
		var n int
		var err error

		if n, err = w.WriteBit(byte(i), i); err != nil {
			t.Fatalf("WriteBit happen error %v", err)
		}

		if n != i {
			t.Fatalf("WriteBit write size %d, want %d", n, i)
		}
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("WriteBit happen error %v", err)
	}

	if reflect.DeepEqual(b.Bytes(), exp) == false {
		t.Fatalf("WriteBit write data %#v, want %#v", b.Bytes(), exp)
	}
}

func TestBitWriteBuffer_WriteBits(t *testing.T) {
	var tests = []struct {
		data []byte
		bits int
		exp  []byte
	}{
		// write bit (full bit)
		{[]byte{0xff, 0xff}, 1, []byte{0x80}},
		{[]byte{0xff, 0xff}, 7, []byte{0xfe}},
		{[]byte{0xff, 0xff}, 8, []byte{0xff}},
		{[]byte{0xff, 0xff}, 9, []byte{0xff, 0x80}},
		{[]byte{0xff, 0xff}, 15, []byte{0xff, 0xfe}},
		{[]byte{0xff, 0xff}, 16, []byte{0xff, 0xff}},
		// write bit
		{[]byte{0xab, 0xcd}, 1, []byte{0x80}},
		{[]byte{0xab, 0xcd}, 7, []byte{0x9a}},
		{[]byte{0xab, 0xcd}, 8, []byte{0xcd}},
		{[]byte{0xab, 0xcd}, 9, []byte{0xe6, 0x80}},
		{[]byte{0xab, 0xcd}, 15, []byte{0x57, 0x9a}},
		{[]byte{0xab, 0xcd}, 16, []byte{0xab, 0xcd}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitWriteBuffer(b)

		if n, err = w.WriteBits(tt.data, tt.bits); err != nil {
			t.Fatalf("WriteBits happen error %v", err)
		}

		if err = w.Flush(); err != nil {
			t.Fatalf("WriteBits happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("WriteBits write size %d, want %d", n, tt.bits)
		}

		if reflect.DeepEqual(b.Bytes(), tt.exp) == false {
			t.Fatalf("WriteBits write data %#v, want %#v", b.Bytes(), tt.exp)
		}
	}
}

func TestBitWriteBuffer_WriteBits_nbit(t *testing.T) {
	// write number N to N bit
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101" +
		"000110" +
		"0000111" +
		"00001000" +
		"000001001" +
		"0000001010"
	exp := binaryToByteArray(str)

	b := bytes.NewBuffer([]byte{})
	w := bitio.NewBitWriteBuffer(b)
	for i := 1; i <= 10; i++ {
		var n int
		var err error

		data := make([]byte, (i+7)/8)
		data[len(data)-1] = byte(i)

		if n, err = w.WriteBits(data, i); err != nil {
			t.Fatalf("WriteBits happen error %v", err)
		}

		if n != i {
			t.Fatalf("WriteBits write size %d, want %d", n, i)
		}
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("WriteBits happen error %v", err)
	}

	if reflect.DeepEqual(b.Bytes(), exp) == false {
		t.Fatalf("WriteBits write data %#v, want %#v", b.Bytes(), exp)
	}
}

func TestBitWriteBuffer_Write(t *testing.T) {
	var tests = []struct {
		data []byte
		exp  []byte
	}{
		// write byte (full byte)
		{[]byte{0xff}, []byte{0xff}},
		{[]byte{0xff, 0xff}, []byte{0xff, 0xff}},
		// write byte
		{[]byte{0xab}, []byte{0xab}},
		{[]byte{0xab, 0xcd}, []byte{0xab, 0xcd}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitWriteBuffer(b)

		if n, err = w.Write(tt.data); err != nil {
			t.Fatalf("Write happen error %v", err)
		}

		if n != len(tt.data) {
			t.Fatalf("Write write size %d, want %d", n, len(tt.data))
		}

		if err = w.Flush(); err != nil {
			t.Fatalf("Write happen error %v", err)
		}

		if reflect.DeepEqual(b.Bytes(), tt.exp) == false {
			t.Fatalf("Write write data %#v, want %#v", b.Bytes(), tt.exp)
		}
	}
}

func TestBitWriteBuffer_Write_Combination(t *testing.T) {
	exps := [][]byte{
		binaryToByteArray("" +
			"01" +
			"0011_0110" +
			"111" +
			"01_0111_0011" +
			"01_0111_0011" +
			"0011_0110"),
		{0x97, 0x97},
	}

	var tests = [][]struct {
		bits int
		data []byte
	}{
		{
			{2, []byte{0x01}},
			{8, []byte{0x36}},
			{3, []byte{0x07}},
			{10, []byte{0x01, 0x73}},
			{10, []byte{0x01, 0x73}},
			{8, []byte{0x36}},
		},
		{
			{1, []byte{0x01}},
			{14, []byte{0x0b, 0xcb}},
			{1, []byte{0x01}},
		},
	}

	for i := 0; i < len(tests); i++ {
		exp := exps[i]
		test := tests[i]
		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitWriteBuffer(b)

		for _, tt := range test {
			var n int
			var err error

			if tt.bits < 8 {
				// test WriteBit
				if n, err = w.WriteBit(tt.data[0], tt.bits); err != nil {
					t.Fatalf("BitWriteBuffer happen error %v", err)
				}
			} else if tt.bits == 8 {
				// test Write
				if n, err = w.Write(tt.data); err != nil {
					t.Fatalf("BitWriteBuffer happen error %v", err)
				}
				n *= 8
			} else {
				// test WriteBits
				if n, err = w.WriteBits(tt.data, tt.bits); err != nil {
					t.Fatalf("BitWriteBuffer happen error %v", err)
				}
			}

			if n != tt.bits {
				t.Fatalf("BitWriteBuffer write size %d, want %d", n, tt.bits)
			}
		}

		if err := w.Flush(); err != nil {
			t.Fatalf("Write happen error %v", err)
		}

		if reflect.DeepEqual(b.Bytes(), exp) == false {
			t.Fatalf("BitWriteBuffer write data %#v, want %#v", b.Bytes(), exp)
		}
	}
}

func BenchmarkBitWriteBuffer_Write_Aligned_32b(b *testing.B) {
	writeAligned(b, 2<<5)
}

func BenchmarkBitWriteBuffer_Write_UnAligned_32b(b *testing.B) {
	writeUnAligned(b, 2<<5)
}

func BenchmarkBitWriteBuffer_Write_Aligned_1024b(b *testing.B) {
	writeAligned(b, 2<<10)
}

func BenchmarkBitWriteBuffer_Write_UnAligned_1024b(b *testing.B) {
	writeUnAligned(b, 2<<10)
}

func BenchmarkBitWriteBuffer_Write_Aligned_65536b(b *testing.B) {
	writeAligned(b, 2<<16)
}

func BenchmarkBitWriteBuffer_Write_UnAligned_65536b(b *testing.B) {
	writeUnAligned(b, 2<<16)
}

func writeAligned(b *testing.B, bufSize int) {
	p := make([]byte, bufSize)
	w := bitio.NewBitWriteBuffer(io.Discard)

	b.SetBytes(int64(bufSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(p)
	}
}

func writeUnAligned(b *testing.B, bufSize int) {
	p := make([]byte, bufSize)
	w := bitio.NewBitWriteBuffer(io.Discard)

	// put off align by 1bit
	w.WriteBit(p[0], 1)

	b.SetBytes(int64(bufSize))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(p)
	}
}

////////////////////////////////////////////////////////////////////////////////

func binaryToByteArray(str string) []byte {
	str = strings.Replace(str, "_", "", -1)

	if len(str)%8 != 0 {
		str += strings.Repeat("0", 8-len(str)%8)
	}

	b := make([]byte, len(str)/8)
	for i := 0; i < len(b); i++ {
		t, _ := strconv.ParseInt(str[:8], 2, 0)
		b[i] = byte(t)
		str = str[8:]
	}

	return b
}

type Infinity struct{}

func (obj *Infinity) Read(p []byte) (int, error) {
	for i := 0; i < len(p); i++ {
		p[i] = 0xed
	}
	return len(p), nil
}
