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
