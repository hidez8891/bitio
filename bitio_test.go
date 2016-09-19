package bitio

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

type TestStruct1 struct {
	Val1 int `byte:"1"`
	Val2 int `byte:"2"`
	Val3 int `byte:"3"`
}

type TestStruct2 struct {
	Val1 int `bit:"3"`
	Val2 int `bit:"4"`
}

type TestStruct3 struct {
	Val1 int `bit:"1"`
	Val2 int `bit:"14"`
	Val3 int `bit:"1"`
}

type TestStruct4 struct {
	Val1 int    `byte:"1"`
	Val2 string `byte:"3"`
}

type TestStruct5 struct {
	Val1 int    `bit:"4"`
	Val2 string `byte:"1"`
	Val3 int    `bit:"4"`
}

type readTester struct {
	src []byte
	dst interface{}
	ans map[string]interface{}
}

var readtests = []readTester{
	{
		src: []byte{0x0a, 0xff, 0x1c, 0xff, 0x01, 0x1c},
		dst: &TestStruct1{},
		ans: map[string]interface{}{
			"Val1": 0x0a,
			"Val2": 0x1cff,
			"Val3": 0x1c01ff,
		},
	},
	{
		src: []byte{0xfa}, // 1111_1010
		dst: &TestStruct2{},
		ans: map[string]interface{}{
			"Val1": 0x7, // 111
			"Val2": 0xd, // 1101
		},
	},
	{
		src: []byte{0x97, 0x97}, // 1001_0111 1001_0111
		dst: &TestStruct3{},
		ans: map[string]interface{}{
			"Val1": 0x1,    // 1
			"Val2": 0x0b2f, // 0010_1111 0010_11 [Little endian]
			"Val3": 0x1,    // 1
		},
	},
	{
		src: []byte{0x01, 'a', 'b', 'c'},
		dst: &TestStruct4{},
		ans: map[string]interface{}{
			"Val1": 0x1,
			"Val2": "abc",
		},
	},
	{
		src: []byte{0x16, 0x11},
		dst: &TestStruct5{},
		ans: map[string]interface{}{
			"Val1": 0x1,
			"Val2": "a",
			"Val3": 0x1,
		},
	},
}

func TestReader(t *testing.T) {
	for _, test := range readtests {
		dst := test.dst
		if err := Read(dst, bytes.NewReader(test.src)); err != nil {
			t.Fatal("Read error:", err)
		}

		rv := reflect.ValueOf(dst).Elem()
		rt := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			v := rv.Field(i)
			f := rt.Field(i)
			if f.PkgPath != "" {
				continue
			}

			name := f.Name
			if tostrCompare(v, test.ans[name]) == false {
				t.Fatalf("%s read %v, want %v", name, v, test.ans[name])
			}
		}
	}
}

func tostrCompare(a, b interface{}) bool {
	as := fmt.Sprintf("%v", a)
	bs := fmt.Sprintf("%v", b)
	return as == bs
}
