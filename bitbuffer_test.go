package bitio

import (
	"bytes"
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
			t.Fatalf("ReadBit happen error %v", err)
		}

		if n != i {
			t.Fatalf("ReadBit read size %d, want %d", n, i)
		}

		exp := []byte{0x00, 0x00}
		if i <= 8 {
			exp[0] = byte(i)
		} else {
			exp[1] = byte(i)
		}

		if reflect.DeepEqual(b, exp) == false {
			t.Fatalf("ReadBit read data %x, want %x", b, exp)
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

		if n != tt.size*8 {
			t.Fatalf("Read read size %d, want %d", n, tt.size*8)
		}

		if reflect.DeepEqual(b, tt.exp) == false {
			t.Fatalf("Read read data %x, want %x", b, tt.exp)
		}
	}
}

func TestBitReadBuffer_Read_Combination(t *testing.T) {
	str := "" +
		"01" +
		"0011_0110" +
		"111" +
		"00_0111_0011" +
		"00_0111_0011" +
		"0011_0110"
	data := binaryToByteArray(str)

	var tests = []struct {
		bits int
		exp  []byte
	}{
		{2, []byte{0x01}},
		{8, []byte{0x36}},
		{3, []byte{0x07}},
		{10, []byte{0x00, 0x73}},
		{10, []byte{0x00, 0x73}},
		{8, []byte{0x36}},
	}

	r := NewBitReadBuffer(bytes.NewReader(data))
	for _, tt := range tests {
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
