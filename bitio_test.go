package bitio

import (
	"bytes"
	"testing"
)

type TestStruct1 struct {
	Val1 int `byte:"1"`
	Val2 int `byte:"2"`
	Val3 int `byte:"3"`
}

func TestReader(t *testing.T) {
	data := []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c}
	vals := TestStruct1{}
	if err := Read(&vals, bytes.NewReader(data)); err != nil {
		t.Fatal("Read error:", err)
	}

	if vals.Val1 != 0x0a {
		t.Fatalf("Read got %x, want %x ", vals.Val1, 0x0a)
	}
	if vals.Val2 != 0x1cff {
		t.Fatalf("Read got %x, want %x ", vals.Val1, 0x1cff)
	}
	if vals.Val3 != 0x1c01ff {
		t.Fatalf("Read got %x, want %x ", vals.Val1, 0x1c01ff)
	}
}
