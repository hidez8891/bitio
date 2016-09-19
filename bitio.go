package bitio

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

type setter interface {
	size() int
	set([]byte, uint, uint) error
}

func setterFactory(rv reflect.Value, size int, len int) (setter, error) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return newIntSetter(rv, size)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return newUintSetter(rv, size)
	case reflect.String:
		return &stringSetter{rv, size}, nil
	case reflect.Slice:
		return newSliceSetter(rv, size, len)
	default:
		return nil, fmt.Errorf("bitio: not support type %q", rv.Kind().String())
	}
}

func newIntSetter(rv reflect.Value, size int) (setter, error) {
	return &numberSetter{rv, size, true}, nil
}

func newUintSetter(rv reflect.Value, size int) (setter, error) {
	return &numberSetter{rv, size, false}, nil
}

type numberSetter struct {
	rval   reflect.Value
	bits   int
	signed bool
}

func (s *numberSetter) size() int {
	return s.bits
}

func (s *numberSetter) set(b []byte, leftpad, rightpad uint) error {
	if s.signed {
		value := toLittleEndianInt(b, leftpad, rightpad)
		s.rval.SetInt(value)
	} else {
		value := toLittleEndianUint(b, leftpad, rightpad)
		s.rval.SetUint(value)
	}
	return nil
}

type stringSetter struct {
	rval reflect.Value
	bits int
}

func (s *stringSetter) size() int {
	return s.bits
}

func (s *stringSetter) set(b []byte, leftpad, rightpad uint) error {
	if s.bits%8 != 0 {
		return fmt.Errorf("bitio: string type size need to multiple of 8bit")
	}

	value := lBitsShift(b, leftpad)
	padding := leftpad + rightpad
	if padding >= 8 {
		padding -= 8
		value = append([]byte{}, value[:len(value)-1]...)
	}

	s.rval.SetString(string(value))
	return nil
}

func newSliceSetter(rv reflect.Value, size int, len int) (setter, error) {
	if len < 1 {
		return nil, fmt.Errorf("bitio: slice type need length")
	}

	// assign
	rv.Set(reflect.MakeSlice(rv.Type(), len, len))

	elems := make([]setter, len)
	for i := 0; i < len; i++ {
		ev := rv.Index(i)
		ef, err := setterFactory(ev, size, 0)
		if err != nil {
			return nil, err
		}
		elems[i] = ef
	}

	return &sliceSetter{rv, elems}, nil
}

type sliceSetter struct {
	rval  reflect.Value
	elems []setter
}

func (s *sliceSetter) size() int {
	var sum int
	for _, e := range s.elems {
		sum += e.size()
	}
	return sum
}

func (s *sliceSetter) set(b []byte, leftpad, _ uint) error {
	for _, e := range s.elems {
		begin := leftpad / 8
		rightpos := leftpad + uint(e.size())
		end := (rightpos + 7) / 8
		rightpad := end*8 - rightpos

		if err := e.set(b[begin:end], leftpad, rightpad); err != nil {
			return err
		}

		leftpad += uint(e.size())
	}
	return nil
}

// Read method read data from srcreader and save to dstptr.
// dstptr's type is needed pointer type of struct.
func Read(dstptr interface{}, srcreader io.Reader) error {
	rv := reflect.ValueOf(dstptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("bitio.Read: want to set pointer type of struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bitio.Read: want to set pointer type of struct")
	}
	rt := rv.Type()

	// read each filed size
	fields := make(map[string]setter)
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)
		v := rv.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		var (
			size int
			len  int
			err  error
		)

		if v, ok := f.Tag.Lookup("byte"); ok {
			size, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("bitio.Read: %s has invalid size %q byte(s)", f.Name, v)
			}
			size *= 8
		} else if v, ok := f.Tag.Lookup("bit"); ok {
			size, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("bitio.Read: %s has invalid size %q bit(s)", f.Name, v)
			}
		} else {
			return fmt.Errorf("bitio.Read: %s need size hint", f.Name)
		}

		if v, ok := f.Tag.Lookup("len"); ok {
			len, err = strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("bitio.Read: %s has invalid length %q", f.Name, v)
			}
		}

		setfunc, err := setterFactory(v, size, len)
		if err != nil {
			return err
		}
		fields[f.Name] = setfunc
	}

	// read from reader
	readsize := 0
	for _, v := range fields {
		readsize += v.size()
	}
	readsize += 7 // padding
	readsize /= 8

	buffer := make([]byte, readsize)
	n, err := srcreader.Read(buffer)
	if err != nil {
		return fmt.Errorf("bitio.Read: read failed")
	}
	if n != readsize {
		return fmt.Errorf("bitio.Read: read %d bytes, want %d bytes", n, readsize)
	}

	// save to struct
	posbit := int(0)
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		sizebit := fields[f.Name].size()
		lpos := posbit / 8
		rpos := (posbit + sizebit + 7) / 8
		lpad := uint(posbit - lpos*8)
		rpad := uint(rpos*8 - posbit - sizebit)
		posbit += sizebit
		if err := fields[f.Name].set(buffer[lpos:rpos], lpad, rpad); err != nil {
			return err
		}
	}

	return nil
}

func toLittleEndianInt(b []byte, leftpad, rightpad uint) int64 {
	return int64(toLittleEndianUint(b, leftpad, rightpad))
}

func toLittleEndianUint(b []byte, leftpad, rightpad uint) uint64 {
	// require: leftpad, rightpad < 8
	w := lBitsShift(b, leftpad)
	padding := leftpad + rightpad
	if padding >= 8 {
		padding -= 8
		w = append([]byte{}, w[:len(w)-1]...)
	}
	w[len(w)-1] >>= padding // alignment

	value := uint64(0)
	digit := uint(0)

	for _, w := range w {
		// little endian
		value |= uint64(w) << digit
		digit += 8
	}

	return value
}

func lBitsShift(src []byte, shift uint) []byte {
	// require shift < 8
	dst := make([]byte, len(src))

	mask := byte(0xff >> shift)
	for i, b := range src {
		dst[i] |= (b & mask) << shift
	}

	mask = ^mask
	for i, b := range src[1:] {
		dst[i] |= (b & mask) >> (8 - shift)
	}

	return dst
}
