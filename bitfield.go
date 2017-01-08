package bitio

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// NewBitFieldReader returns BitFieldReader
func NewBitFieldReader(r io.Reader) *BitFieldReader {
	return NewBitFieldReader2(NewBitReadBuffer(r))
}

// NewBitFieldReader2 returns BitFieldReader
func NewBitFieldReader2(r BitReader) *BitFieldReader {
	return &BitFieldReader{
		r: r,
	}
}

// BitFieldReader read bit-field data.
type BitFieldReader struct {
	r BitReader
}

// Read reads bit-field data and returns read size.
// If error happen, err will be set.
func (obj *BitFieldReader) Read(p interface{}) (nBit int, err error) {
	// check argument type
	var rv reflect.Value
	if rv = reflect.ValueOf(p); rv.Kind() != reflect.Ptr {
		err = fmt.Errorf("Read: argument wants to pointer of struct")
		return
	}
	if rv = rv.Elem(); rv.Kind() != reflect.Struct {
		err = fmt.Errorf("Read: argument wants to pointer of struct")
		return
	}
	rt := rv.Type()

	// read bit-fields
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		ptr := rv.Field(i)

		// skip unexport field
		if field.PkgPath != "" {
			continue
		}

		// bit-field size
		bits := 0
		if v, ok := field.Tag.Lookup("byte"); ok {
			if bits, err = strconv.Atoi(v); err != nil {
				err = fmt.Errorf("%s has invalid size %q byte(s)", field.Name, v)
			}
			bits *= 8
		} else if v, ok := field.Tag.Lookup("bit"); ok {
			if bits, err = strconv.Atoi(v); err != nil {
				err = fmt.Errorf("%s has invalid size %q bit(s)", field.Name, v)
			}
		} else {
			err = fmt.Errorf("%s need size hint", field.Name)
		}
		if err != nil {
			return
		}

		// bit-field block count
		len := 0
		if v, ok := field.Tag.Lookup("len"); ok {
			if len, err = strconv.Atoi(v); err != nil {
				err = fmt.Errorf("%s has invalid length %q", field.Name, v)
			}
		}
		if err != nil {
			return
		}

		// read bit-filed
		var (
			r fieldReader
			n int
		)
		if r, err = newFieldReader(obj.r, ptr, bits, len); err != nil {
			return
		}
		if n, err = r.read(); err != nil {
			return
		}
		nBit += n
	}

	return
}

////////////////////////////////////////////////////////////////////////////////

type fieldReader interface {
	read() (nBit int, err error)
}

func newFieldReader(r BitReader, rv reflect.Value, bits, len int) (fr fieldReader, err error) {
	if bits < 1 {
		return nil, fmt.Errorf("invalid bit-field size %d byte(s)", bits)
	}

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fr = &fieldIntReader{r, rv, bits}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fr = &fieldUintReader{r, rv, bits}
	case reflect.String:
		fr = &fieldStringReader{r, rv, bits}
	case reflect.Slice:
		if len < 1 {
			return nil, fmt.Errorf("Slice type needs positive length")
		}
		fr = &fieldSliceReader{r, rv, bits, len}
	default:
		return nil, fmt.Errorf("Not support bit-filed type %q", rv.Kind().String())
	}

	return
}

// fieldIntReader implemented fieldReader for Integer type
type fieldIntReader struct {
	r    BitReader
	ptr  reflect.Value
	bits int
}

func (obj *fieldIntReader) read() (nBit int, err error) {
	if obj.bits > 64 {
		err = fmt.Errorf("bit-field size needs <= 64bit")
		return
	}

	buf := make([]byte, 8)
	if nBit, err = obj.r.ReadBits(buf, obj.bits); err != nil {
		return
	}

	// little endian shift
	// 12bit: 0x0123 -> 0x1230 -> 0x1203
	if nBit%8 > 0 {
		leftShift(buf, uint(8-nBit%8))
		buf[nBit/8] >>= uint(8 - nBit%8)
	}

	// TODO read negative value
	obj.ptr.SetInt(int64(binary.LittleEndian.Uint64(buf)))
	return
}

// fieldUintReader implemented fieldReader for Unsigned Integer type
type fieldUintReader struct {
	r    BitReader
	ptr  reflect.Value
	bits int
}

func (obj *fieldUintReader) read() (nBit int, err error) {
	if obj.bits > 64 {
		err = fmt.Errorf("bit-field size needs <= 64bit")
		return
	}

	buf := make([]byte, 8)
	if nBit, err = obj.r.ReadBits(buf, obj.bits); err != nil {
		return
	}

	// little endian shift
	// 12bit: 0x0123 -> 0x1230 -> 0x1203
	if nBit%8 > 0 {
		leftShift(buf, uint(8-nBit%8))
		buf[nBit/8] >>= uint(8 - nBit%8)
	}

	obj.ptr.SetUint(binary.LittleEndian.Uint64(buf))
	return
}

// fieldStringReader implemented fieldReader for String type
type fieldStringReader struct {
	r    BitReader
	ptr  reflect.Value
	bits int
}

func (obj *fieldStringReader) read() (nBit int, err error) {
	if obj.bits%8 != 0 {
		err = fmt.Errorf("String type size needs to 8*n bits")
		return
	}

	buf := make([]byte, obj.bits/8)
	if nBit, err = obj.r.Read(buf); err != nil {
		nBit *= 8
		return
	}
	nBit *= 8

	obj.ptr.SetString(string(buf))
	return
}

// fieldSliceReader implemented fieldReader for Slice type
type fieldSliceReader struct {
	r    BitReader
	ptr  reflect.Value
	bits int
	len  int
}

func (obj *fieldSliceReader) read() (nBit int, err error) {
	if obj.len < 1 {
		err = fmt.Errorf("Slice type needs positive length")
		return
	}

	// (re-)allocate slice space
	if obj.ptr.Len() < obj.len {
		rv := reflect.MakeSlice(obj.ptr.Type(), obj.len, obj.len)
		reflect.Copy(rv, obj.ptr)
		obj.ptr.Set(rv)
	}

	// read slice bit-fields
	for i := 0; i < obj.len; i++ {
		var (
			r fieldReader
			n int
		)

		rv := obj.ptr.Index(i)
		if r, err = newFieldReader(obj.r, rv, obj.bits, 0); err != nil {
			return
		}

		if n, err = r.read(); err != nil {
			return
		}
		nBit += n
	}

	return
}
