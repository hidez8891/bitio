package bitio_test

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/hidez8891/bitio"
)

func TestBitField_interface(t *testing.T) {
	// Only compile test

	var r io.Reader = &bitio.BitFieldReader{}
	_ = r
	var w io.Writer = &bitio.BitFieldWriter{}
	_ = w
}

type TestBitFieldInt1 struct {
	Val1 int `byte:"1"`
	Val2 int `byte:"2"`
	Val3 int `byte:"3"`
}

type TestBitFieldInt2 struct {
	Val1 int `bit:"1"`
	Val2 int `bit:"14"`
	Val3 int `bit:"1"`
}

type TestBitFieldInt3 struct {
	Val1 int `byte:"2" endian:"big"`
	Val2 int `byte:"2" endian:"little"`
}

type TestBitFieldUint1 struct {
	Val1 uint `byte:"1"`
	Val2 uint `byte:"2"`
	Val3 uint `byte:"3"`
}

type TestBitFieldUint2 struct {
	Val1 uint `bit:"1"`
	Val2 uint `bit:"14"`
	Val3 uint `bit:"1"`
}

type TestBitFieldUint3 struct {
	Val1 uint `byte:"2" endian:"big"`
	Val2 uint `byte:"2" endian:"little"`
}

type TestBitFieldString1 struct {
	Val1 string `byte:"1"`
	Val2 string `byte:"3"`
}

type TestBitFieldSlice1 struct {
	Val1 []int  `bit:"4" len:"2"`
	Val2 []byte `bit:"4" len:"2"`
}

type TestBitFieldCombination1 struct {
	Val1 int8  `byte:"1"`
	Val2 uint8 `byte:"1"`
}

type TestBitFieldCombination2 struct {
	Val1 int8   `bit:"4"`
	Val2 string `bit:"8"`
	Val3 uint8  `bit:"4"`
}

type TestData struct {
	raw  []byte
	ptr  interface{}
	exp  map[string]interface{}
	bits int
}

var tests = []TestData{
	{
		raw: []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c},
		ptr: &TestBitFieldInt1{},
		exp: map[string]interface{}{
			"Val1": 0x0a,
			"Val2": 0x1cff,
			"Val3": 0x1c01ff,
		},
		bits: 48,
	},
	{
		raw: []byte{0x97, 0x97},
		ptr: &TestBitFieldInt2{},
		exp: map[string]interface{}{
			"Val1": 0x1,    // 1
			"Val2": 0x0b2f, // 0010_1111 0010_11 [Little endian]
			"Val3": 0x1,    // 1
		},
		bits: 16,
	},
	{
		raw: []byte{0x0a, 0x01, 0x0a, 0x01},
		ptr: &TestBitFieldInt3{},
		exp: map[string]interface{}{
			"Val1": 0x0a01,
			"Val2": 0x010a,
		},
		bits: 32,
	},
	{
		raw: []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c},
		ptr: &TestBitFieldUint1{},
		exp: map[string]interface{}{
			"Val1": 0x0a,
			"Val2": 0x1cff,
			"Val3": 0x1c01ff,
		},
		bits: 48,
	},
	{
		raw: []byte{0x97, 0x97},
		ptr: &TestBitFieldUint2{},
		exp: map[string]interface{}{
			"Val1": 0x1,    // 1
			"Val2": 0x0b2f, // 0010_1111 0010_11 [Little endian]
			"Val3": 0x1,    // 1
		},
		bits: 16,
	},
	{
		raw: []byte{0x0a, 0x01, 0x0a, 0x01},
		ptr: &TestBitFieldUint3{},
		exp: map[string]interface{}{
			"Val1": 0x0a01,
			"Val2": 0x010a,
		},
		bits: 32,
	},
	{
		raw: []byte{'0', 'a', 'b', 'c'},
		ptr: &TestBitFieldString1{},
		exp: map[string]interface{}{
			"Val1": "0",
			"Val2": "abc",
		},
		bits: 32,
	},
	{
		raw: []byte{0xca, 0xca},
		ptr: &TestBitFieldSlice1{},
		exp: map[string]interface{}{
			"Val1": []int{0x0c, 0x0a},
			"Val2": []byte{0x0c, 0x0a},
		},
		bits: 16,
	},
	{
		raw: []byte{0x80, 0x80},
		ptr: &TestBitFieldCombination1{},
		exp: map[string]interface{}{
			"Val1": -128,
			"Val2": 128,
		},
		bits: 16,
	},
	{
		raw: []byte{0x16, 0x11},
		ptr: &TestBitFieldCombination2{},
		exp: map[string]interface{}{
			"Val1": 0x1,
			"Val2": "a",
			"Val3": 0x1,
		},
		bits: 16,
	},
}

func TestBitFieldReader_Read(t *testing.T) {
	for _, tt := range tests {
		r := bitio.NewBitFieldReader(bytes.NewReader(tt.raw))
		ptr := tt.ptr

		var (
			err error
			n   int
		)

		if n, err = r.ReadStruct(ptr); err != nil {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("Read %v error: %v", rt, err)
		}

		if n != tt.bits {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("Read %v read size %d, want %d", rt, n, tt.bits)
		}

		rv := reflect.ValueOf(ptr).Elem()
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			v := rv.Field(i)
			f := rt.Field(i)

			if f.PkgPath != "" {
				continue
			}

			if toStrCompare(v, tt.exp[f.Name]) == false {
				rt := reflect.TypeOf(ptr)
				t.Fatalf("%v %s read %v, want %v", rt, f.Name, v, tt.exp[f.Name])
			}
		}
	}
}

func TestBitFieldWriter_Write(t *testing.T) {
	for _, tt := range tests {
		var (
			err error
			n   int
		)

		ptr := tt.ptr
		r := bitio.NewBitFieldReader(bytes.NewReader(tt.raw))
		if _, err = r.ReadStruct(ptr); err != nil {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("Write test initialize %v error: %v", rt, err)
		}

		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitFieldWriter(b)
		if n, err = w.WriteStruct(ptr); err != nil {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("Write %v error: %v", rt, err)
		}

		if n != tt.bits {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("Write %v write size %d, want %d", rt, n, tt.bits)
		}

		if err = w.Flush(); err != nil {
			t.Fatalf("Write flush happen error %v", err)
		}

		if reflect.DeepEqual(b.Bytes(), tt.raw) == false {
			rt := reflect.TypeOf(ptr)
			t.Fatalf("%v write %v, want %v", rt, b.Bytes(), tt.raw)
		}
	}
}

func toStrCompare(a, b interface{}) bool {
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as == bs
}
