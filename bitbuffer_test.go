package bitio

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestBitReadBuffer_ReadBit(t *testing.T) {
	var tests = []struct {
		data []byte
		bits int
		exp  byte
	}{
		{[]byte{0xff}, 1, 0x01},
		{[]byte{0xff}, 7, 0x7f},
		{[]byte{0xff}, 8, 0xff},
	}

	for _, tt := range tests {
		var n int
		var b byte
		var err error

		r := NewBitReadBuffer(bytes.NewReader(tt.data))
		if n, err = r.ReadBit(&b, tt.bits); err != nil {
			t.Fatalf("ReadBit happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("ReadBit read size %d, want %d", n, tt.bits)
		}

		if b != tt.exp {
			t.Fatalf("ReadBit read data %x, want %x", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_ReadBit_Loop(t *testing.T) {
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101"
	data := binaryToByteArray(str)

	r := NewBitReadBuffer(bytes.NewReader(data))
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
		{[]byte{0xff, 0xff}, 1, []byte{0x01}},
		{[]byte{0xff, 0xff}, 7, []byte{0x7f}},
		{[]byte{0xff, 0xff}, 8, []byte{0xff}},
		{[]byte{0xff, 0xff}, 9, []byte{0x01, 0xff}},
		{[]byte{0xff, 0xff}, 15, []byte{0x7f, 0xff}},
		{[]byte{0xff, 0xff}, 16, []byte{0xff, 0xff}},
	}

	for _, tt := range tests {
		var n int
		var b []byte
		var err error

		r := NewBitReadBuffer(bytes.NewReader(tt.data))
		b = make([]byte, (tt.bits+7)/8)

		if n, err = r.ReadBits(b, tt.bits); err != nil {
			t.Fatalf("ReadBits happen error %v", err)
		}

		if n != tt.bits {
			t.Fatalf("ReadBits read size %d, want %d", n, tt.bits)
		}

		if reflect.DeepEqual(b, tt.exp) == false {
			t.Fatalf("ReadBits read data %x, want %x", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_ReadBits_Loop(t *testing.T) {
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

	r := NewBitReadBuffer(bytes.NewReader(data))
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
		if i <= 8 {
			exp[0] = byte(i)
		} else {
			exp[1] = byte(i)
		}

		if reflect.DeepEqual(b, exp) == false {
			t.Fatalf("ReadBits read data %x, want %x", b, exp)
		}
	}
}

func TestBitReadBuffer_Read(t *testing.T) {
	var tests = []struct {
		data []byte
		size int
		exp  []byte
	}{
		{[]byte{0xff, 0xff}, 1, []byte{0xff}},
		{[]byte{0xff, 0xff}, 2, []byte{0xff, 0xff}},
	}

	for _, tt := range tests {
		var n int
		var b []byte
		var err error

		r := NewBitReadBuffer(bytes.NewReader(tt.data))
		b = make([]byte, tt.size)

		if n, err = r.Read(b); err != nil {
			t.Fatalf("Read happen error %v", err)
		}

		if n != tt.size {
			t.Fatalf("Read read size %d, want %d", n, tt.size)
		}

		if reflect.DeepEqual(b, tt.exp) == false {
			t.Fatalf("Read read data %x, want %x", b, tt.exp)
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
		[]byte{0x97, 0x97},
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
		r := NewBitReadBuffer(bytes.NewReader(data))

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
				t.Fatalf("BitReadBuffer read data %x, want %x", b, tt.exp)
			}
		}
	}
}

func BenchmarkBitReadBuffer_Read_FixedAlign_small(b *testing.B) {
	readFixedAlign(b, 2<<5)
}

func BenchmarkBitReadBuffer_Read_NoFixedAlign_small(b *testing.B) {
	readNoFixedAlign(b, 2<<5)
}

func BenchmarkBitReadBuffer_Read_FixedAlign_middle(b *testing.B) {
	readFixedAlign(b, 2<<10)
}

func BenchmarkBitReadBuffer_Read_NoFixedAlign_middle(b *testing.B) {
	readNoFixedAlign(b, 2<<10)
}

func BenchmarkBitReadBuffer_Read_FixedAlign_large(b *testing.B) {
	readFixedAlign(b, 2<<16)
}

func BenchmarkBitReadBuffer_Read_NoFixedAlign_large(b *testing.B) {
	readNoFixedAlign(b, 2<<16)
}

func readFixedAlign(b *testing.B, size int) {
	r := NewBitReadBuffer(&Infinity{})
	p := make([]byte, size)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Read(p)
	}
}

func readNoFixedAlign(b *testing.B, size int) {
	r := NewBitReadBuffer(&Infinity{})
	p := make([]byte, size)

	// put off align by 1bit
	r.ReadBit(&p[0], 1)

	b.SetBytes(int64(size))
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
		{0xff, 1, []byte{0x80}},
		{0xff, 7, []byte{0xfe}},
		{0xff, 8, []byte{0xff}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := NewBitWriteBuffer(b)
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
			t.Fatalf("WriteBit write data %x, want %x", b.Bytes(), tt.exp)
		}
	}
}

func TestBitWriteBuffer_WriteBit_Loop(t *testing.T) {
	str := "" +
		"1" +
		"10" +
		"011" +
		"0100" +
		"00101"
	exp := binaryToByteArray(str)

	b := bytes.NewBuffer([]byte{})
	w := NewBitWriteBuffer(b)
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
		t.Fatalf("WriteBit write data %x, want %x", b.Bytes(), exp)
	}
}

func TestBitWriteBuffer_WriteBits(t *testing.T) {
	var tests = []struct {
		data []byte
		bits int
		exp  []byte
	}{
		{[]byte{0xff, 0xff}, 1, []byte{0x80}},
		{[]byte{0xff, 0xff}, 7, []byte{0xfe}},
		{[]byte{0xff, 0xff}, 8, []byte{0xff}},
		{[]byte{0xff, 0xff}, 9, []byte{0xff, 0x80}},
		{[]byte{0xff, 0xff}, 15, []byte{0xff, 0xfe}},
		{[]byte{0xff, 0xff}, 16, []byte{0xff, 0xff}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := NewBitWriteBuffer(b)

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
			t.Fatalf("WriteBits write data %x, want %x", b.Bytes(), tt.exp)
		}
	}
}

func TestBitWriteBuffer_WriteBits_Loop(t *testing.T) {
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
	w := NewBitWriteBuffer(b)
	for i := 1; i <= 10; i++ {
		var n int
		var err error

		data := []byte{0x00, 0x00}
		if i <= 8 {
			data[0] = byte(i)
		} else {
			data[1] = byte(i)
		}

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
		t.Fatalf("WriteBits write data %x, want %x", b.Bytes(), exp)
	}
}

func TestBitWriteBuffer_Write(t *testing.T) {
	var tests = []struct {
		data []byte
		exp  []byte
	}{
		{[]byte{0xfe}, []byte{0xfe}},
		{[]byte{0xff, 0xfe}, []byte{0xff, 0xfe}},
	}

	for _, tt := range tests {
		var n int
		var err error

		b := bytes.NewBuffer([]byte{})
		w := NewBitWriteBuffer(b)

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
			t.Fatalf("Write write data %x, want %x", b.Bytes(), tt.exp)
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
		[]byte{0x97, 0x97},
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
		w := NewBitWriteBuffer(b)

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
			t.Fatalf("BitWriteBuffer write data %x, want %x", b.Bytes(), exp)
		}
	}
}

func BenchmarkBitWriteBuffer_Write_FixedAlign_small(b *testing.B) {
	writeFixedAlign(b, 2<<5)
}

func BenchmarkBitWriteBuffer_Write_NoFixedAlign_small(b *testing.B) {
	writeNoFixedAlign(b, 2<<5)
}

func BenchmarkBitWriteBuffer_Write_FixedAlign_middle(b *testing.B) {
	writeFixedAlign(b, 2<<10)
}

func BenchmarkBitWriteBuffer_Write_NoFixedAlign_middle(b *testing.B) {
	writeNoFixedAlign(b, 2<<10)
}

func BenchmarkBitWriteBuffer_Write_FixedAlign_large(b *testing.B) {
	writeFixedAlign(b, 2<<16)
}

func BenchmarkBitWriteBuffer_Write_NoFixedAlign_large(b *testing.B) {
	writeNoFixedAlign(b, 2<<16)
}

func writeFixedAlign(b *testing.B, size int) {
	p := make([]byte, size)
	w := NewBitWriteBuffer(ioutil.Discard)

	b.SetBytes(int64(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(p)
	}
}

func writeNoFixedAlign(b *testing.B, size int) {
	p := make([]byte, size)
	w := NewBitWriteBuffer(ioutil.Discard)

	// put off align by 1bit
	w.WriteBit(p[0], 1)

	b.SetBytes(int64(size))
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
