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

var bitfieldTests = []struct {
	name string
	raw  []byte
	ptr  interface{}
	exp  map[string]interface{}
	bits int
}{
	{
		name: "bool field 01",
		raw:  []byte{0x80},
		ptr: &struct {
			Val1 bool `bit:"1"`
			Val2 bool `bit:"1"`
		}{},
		exp: map[string]interface{}{
			"Val1": true,
			"Val2": false,
		},
		bits: 2,
	},
	{
		name: "bool field 02",
		raw:  []byte{0x40},
		ptr: &struct {
			Val1 bool `bit:"1"`
			Val2 bool `bit:"1"`
		}{},
		exp: map[string]interface{}{
			"Val1": false,
			"Val2": true,
		},
		bits: 2,
	},
	{
		name: "bool field 03",
		raw:  []byte{0x00, 0x01, 0x00, 0x00},
		ptr: &struct {
			Val1 bool `byte:"2" endian:"big"`
			Val2 bool `byte:"2" endian:"little"`
		}{},
		exp: map[string]interface{}{
			"Val1": true,
			"Val2": false,
		},
		bits: 32,
	},
	{
		name: "bool field 04",
		raw:  []byte{0x00, 0x00, 0x01, 0x00},
		ptr: &struct {
			Val1 bool `byte:"2" endian:"big"`
			Val2 bool `byte:"2" endian:"little"`
		}{},
		exp: map[string]interface{}{
			"Val1": false,
			"Val2": true,
		},
		bits: 32,
	},
	{
		name: "int field 01",
		raw:  []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c},
		ptr: &struct {
			Val1 int `byte:"1"`
			Val2 int `byte:"2"`
			Val3 int `byte:"3"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x0a,
			"Val2": 0x1cff,
			"Val3": 0x1c01ff,
		},
		bits: 48,
	},
	{
		name: "int field 02",
		raw:  []byte{0x97, 0x97},
		ptr: &struct {
			Val1 int `bit:"1"`
			Val2 int `bit:"14"`
			Val3 int `bit:"1"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x1,    // 1
			"Val2": 0x0b2f, // 0010_1111 00_1011 [Little endian]
			"Val3": 0x1,    // 1
		},
		bits: 16,
	},
	{
		name: "int field 03",
		raw:  []byte{0x0a, 0x01, 0x0a, 0x01},
		ptr: &struct {
			Val1 int `byte:"2" endian:"big"`
			Val2 int `byte:"2" endian:"little"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x0a01,
			"Val2": 0x010a,
		},
		bits: 32,
	},
	{
		name: "uint field 01",
		raw:  []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c},
		ptr: &struct {
			Val1 uint `byte:"1"`
			Val2 uint `byte:"2"`
			Val3 uint `byte:"3"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x0a,
			"Val2": 0x1cff,
			"Val3": 0x1c01ff,
		},
		bits: 48,
	},
	{
		name: "uint field 02",
		raw:  []byte{0x97, 0x97},
		ptr: &struct {
			Val1 uint `bit:"1"`
			Val2 uint `bit:"14"`
			Val3 uint `bit:"1"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x1,    // 1
			"Val2": 0x0b2f, // 0010_1111 00_1011 [Little endian]
			"Val3": 0x1,    // 1
		},
		bits: 16,
	},
	{
		name: "uint field 03",
		raw:  []byte{0x0a, 0x01, 0x0a, 0x01},
		ptr: &struct {
			Val1 uint `byte:"2" endian:"big"`
			Val2 uint `byte:"2" endian:"little"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x0a01,
			"Val2": 0x010a,
		},
		bits: 32,
	},
	{
		name: "string field 01",
		raw:  []byte{'0', 'a', 'b', 'c'},
		ptr: &struct {
			Val1 string `byte:"1"`
			Val2 string `byte:"3"`
		}{},
		exp: map[string]interface{}{
			"Val1": "0",
			"Val2": "abc",
		},
		bits: 32,
	},
	{
		name: "slice field 01",
		raw:  []byte{0xca, 0xca},
		ptr: &struct {
			Val1 []int  `bit:"4" len:"2"`
			Val2 []byte `bit:"4" len:"2"`
		}{},
		exp: map[string]interface{}{
			"Val1": []int{0x0c, 0x0a},
			"Val2": []byte{0x0c, 0x0a},
		},
		bits: 16,
	},
	{
		name: "combination 01",
		raw:  []byte{0x80, 0x80},
		ptr: &struct {
			Val1 int8  `byte:"1"`
			Val2 uint8 `byte:"1"`
		}{},
		exp: map[string]interface{}{
			"Val1": -128,
			"Val2": 128,
		},
		bits: 16,
	},
	{
		name: "combination 02",
		raw:  []byte{0x16, 0x11},
		ptr: &struct {
			Val1 int8   `bit:"4"`
			Val2 string `bit:"8"`
			Val3 uint8  `bit:"4"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x1,
			"Val2": "a",
			"Val3": 0x1,
		},
		bits: 16,
	},
	{
		name: "uint boundaries 01",
		raw:  []byte{0x43, 0x52, 0x01},
		ptr: &struct {
			Val1 uint16 `bit:"12" endian:"big"`
			Val2 uint16 `bit:"12" endian:"big"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x435,
			"Val2": 0x201,
		},
		bits: 24,
	},
	{
		name: "uint boundaries 02",
		raw:  []byte{0x43, 0x52, 0x01},
		ptr: &struct {
			Val1 uint16 `bit:"12" endian:"big"`
			Val2 uint16 `bit:"12" endian:"little"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x435,
			"Val2": 0x120,
		},
		bits: 24,
	},
	{
		name: "variable length 01",
		raw:  []byte{0x05, 0x11, 0x22, 0x33, 0x44, 0x55},
		ptr: &struct {
			Val1 uint8  `bit:"8"`
			Val2 []byte `byte:"1" len:"Val1"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x05,
			"Val2": []byte{0x11, 0x22, 0x33, 0x44, 0x55},
		},
		bits: 48,
	},
	{
		name: "variable length 02",
		raw:  []byte{0x51, 0x23, 0x45},
		ptr: &struct {
			Val1 uint8  `bit:"4"`
			Val2 []byte `bit:"4" len:"Val1"`
		}{},
		exp: map[string]interface{}{
			"Val1": 0x05,
			"Val2": []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		},
		bits: 24,
	},
}

func TestBitFieldReader_Read(t *testing.T) {
	for _, tt := range bitfieldTests {
		r := bitio.NewBitFieldReader(bytes.NewReader(tt.raw))
		ptr := tt.ptr

		var err error
		var n int

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
				t.Fatalf("%v %s read %#v, want %#v", rt, f.Name, v, tt.exp[f.Name])
			}
		}
	}
}

func TestBitFieldWriter_Write(t *testing.T) {
	for _, tt := range bitfieldTests {
		var err error
		var n int

		ptr := tt.ptr
		r := bitio.NewBitFieldReader(bytes.NewReader(tt.raw))
		if _, err = r.ReadStruct(ptr); err != nil {
			t.Fatalf("Write test initialize %q error: %v", tt.name, err)
		}

		b := bytes.NewBuffer([]byte{})
		w := bitio.NewBitFieldWriter(b)
		if n, err = w.WriteStruct(ptr); err != nil {
			t.Fatalf("Write %q error: %v", tt.name, err)
		}

		if n != tt.bits {
			t.Fatalf("Write %q write size %d, want %d", tt.name, n, tt.bits)
		}

		if err = w.Flush(); err != nil {
			t.Fatalf("Write flush happen error %v", err)
		}

		if reflect.DeepEqual(b.Bytes(), tt.raw) == false {
			t.Fatalf("%q write %#v, want %#v", tt.name, b.Bytes(), tt.raw)
		}
	}
}

func TestBitFieldWriter_Write_VariableLengthSlice(t *testing.T) {
	ptr := &struct {
		Val1 uint8  `bit:"4"`
		Val2 []byte `bit:"4" len:"Val1"`
	}{
		Val1: 0, // unset length
		Val2: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}
	bits := 24
	exp := []byte{0x51, 0x23, 0x45}
	name := "variable length's variable"

	b := bytes.NewBuffer([]byte{})
	w := bitio.NewBitFieldWriter(b)

	var err error
	var n int

	if n, err = w.WriteStruct(ptr); err != nil {
		t.Fatalf("Write %q error: %v", name, err)
	}

	if n != bits {
		t.Fatalf("Write %q write size %d, want %d", name, n, bits)
	}

	if err = w.Flush(); err != nil {
		t.Fatalf("Write flush happen error %v", err)
	}

	if reflect.DeepEqual(b.Bytes(), exp) == false {
		t.Fatalf("%q write %#v, want %#v", name, b.Bytes(), exp)
	}
}

func toStrCompare(a, b interface{}) bool {
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as == bs
}
