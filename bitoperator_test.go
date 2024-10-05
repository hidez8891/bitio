package bitio

import (
	"reflect"
	"testing"
)

func TestLeftShift(t *testing.T) {
	var tests = []struct {
		src  []byte
		bits uint
		dst  []byte
	}{
		{[]byte{0x11, 0x22, 0x33}, 4, []byte{0x12, 0x23, 0x30}},
		{[]byte{0x11, 0x22, 0x33}, 8, []byte{0x22, 0x33, 0x00}},
		{[]byte{0x11, 0x22, 0x33}, 12, []byte{0x23, 0x30, 0x00}},
		{[]byte{0x11, 0x22, 0x33}, 16, []byte{0x33, 0x00, 0x00}},
		{[]byte{0x11, 0x22, 0x33}, 20, []byte{0x30, 0x00, 0x00}},
		{[]byte{0x11, 0x22, 0x33}, 24, []byte{0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		leftShift(tt.src, tt.bits)

		if reflect.DeepEqual(tt.src, tt.dst) == false {
			t.Fatalf("leftShift returns %#v, want %#v", tt.src, tt.dst)
		}
	}
}

func BenchmarkLeftShift_8b(b *testing.B) {
	benchmarkLeftShift(b, 8)
}

func BenchmarkLeftShift_32b(b *testing.B) {
	benchmarkLeftShift(b, 32)
}

func BenchmarkLeftShift_1024b(b *testing.B) {
	benchmarkLeftShift(b, 1024)
}

func benchmarkLeftShift(b *testing.B, bufSize int) {
	buf := make([]byte, bufSize)

	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		leftShift(buf, 10)
	}
}

func TestRightShift(t *testing.T) {
	var tests = []struct {
		src  []byte
		bits uint
		dst  []byte
	}{
		{[]byte{0x11, 0x22, 0x33}, 4, []byte{0x01, 0x12, 0x23}},
		{[]byte{0x11, 0x22, 0x33}, 8, []byte{0x00, 0x11, 0x22}},
		{[]byte{0x11, 0x22, 0x33}, 12, []byte{0x00, 0x01, 0x12}},
		{[]byte{0x11, 0x22, 0x33}, 16, []byte{0x00, 0x00, 0x11}},
		{[]byte{0x11, 0x22, 0x33}, 20, []byte{0x00, 0x00, 0x01}},
		{[]byte{0x11, 0x22, 0x33}, 24, []byte{0x00, 0x00, 0x00}},
	}

	for _, tt := range tests {
		rightShift(tt.src, tt.bits)

		if reflect.DeepEqual(tt.src, tt.dst) == false {
			t.Fatalf("rightShift returns %#v, want %#v", tt.src, tt.dst)
		}
	}
}

func BenchmarkRightShift_8b(b *testing.B) {
	benchmarkRightShift(b, 8)
}

func BenchmarkRightShift_32b(b *testing.B) {
	benchmarkRightShift(b, 32)
}

func BenchmarkRightShift_1024b(b *testing.B) {
	benchmarkRightShift(b, 1024)
}

func benchmarkRightShift(b *testing.B, bufSize int) {
	buf := make([]byte, bufSize)

	b.SetBytes(int64(len(buf)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rightShift(buf, 10)
	}
}
