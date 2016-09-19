package bitio

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// converter interface
type converter interface {
	size() int
	set([]byte, uint) error
	get([]byte, uint) error
}

func converterFactory(rv reflect.Value, size int, len int) (converter, error) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return newIntConverter(rv, size)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return newUintConverter(rv, size)
	case reflect.String:
		return newStringConverter(rv, size)
	case reflect.Slice:
		return newSliceConverter(rv, size, len)
	default:
		return nil, fmt.Errorf("bitio: not support type %q", rv.Kind().String())
	}
}

// generate int converter
func newIntConverter(rv reflect.Value, size int) (converter, error) {
	return &numberConverter{rv, size, true}, nil
}

// generate uint converter
func newUintConverter(rv reflect.Value, size int) (converter, error) {
	return &numberConverter{rv, size, false}, nil
}

// number (int/uint) converter
type numberConverter struct {
	rval   reflect.Value
	bits   int
	signed bool
}

func (s *numberConverter) size() int {
	return s.bits
}

func (s *numberConverter) set(b []byte, leftpad uint) error {
	rightpad := uint(8*len(b)-s.size()) - leftpad

	if s.signed {
		value := toLittleEndianInt(b, leftpad, rightpad)
		s.rval.SetInt(value)
	} else {
		value := toLittleEndianUint(b, leftpad, rightpad)
		s.rval.SetUint(value)
	}
	return nil
}

func (s *numberConverter) get(b []byte, leftpad uint) error {
	if s.signed {
		value := s.rval.Int()
		fromLittleEndianInt(b, leftpad, s.size(), value)
	} else {
		value := s.rval.Uint()
		fromLittleEndianUint(b, leftpad, s.size(), value)
	}
	return nil
}

// generate string converter
func newStringConverter(rv reflect.Value, size int) (converter, error) {
	if size%8 != 0 {
		return nil, fmt.Errorf("bitio: string type size need to multiple of 8bit")
	}

	return &stringConverter{rv, size}, nil
}

// string converter
type stringConverter struct {
	rval reflect.Value
	bits int
}

func (s *stringConverter) size() int {
	return s.bits
}

func (s *stringConverter) set(b []byte, leftpad uint) error {
	value := lBitsShift(b, leftpad)
	padding := (8*len(b) - s.size())
	if padding >= 8 {
		value = cutSlice(value, len(value)-1)
	}

	s.rval.SetString(string(value))
	return nil
}

func (s *stringConverter) get(b []byte, leftpad uint) error {
	buf := []byte(s.rval.String())
	rBitsShiftCopy(b, buf, leftpad)
	return nil
}

// generate slice converter
func newSliceConverter(rv reflect.Value, size int, len int) (converter, error) {
	if len < 1 {
		return nil, fmt.Errorf("bitio: slice type need length")
	}

	// assign
	if rv.Len() < len {
		rv2 := reflect.MakeSlice(rv.Type(), len, len)
		reflect.Copy(rv2, rv)
		rv.Set(rv2)
	}

	elems := make([]converter, len)
	for i := 0; i < len; i++ {
		ev := rv.Index(i)
		ef, err := converterFactory(ev, size, 0)
		if err != nil {
			return nil, err
		}
		elems[i] = ef
	}

	return &sliceConverter{rv, elems}, nil
}

// slice converter
type sliceConverter struct {
	rval  reflect.Value
	elems []converter
}

func (s *sliceConverter) size() int {
	var sum int
	for _, e := range s.elems {
		sum += e.size()
	}
	return sum
}

func (s *sliceConverter) set(b []byte, leftpad uint) error {
	return s.each(b, leftpad, true)
}

func (s *sliceConverter) get(b []byte, leftpad uint) error {
	return s.each(b, leftpad, false)
}

func (s *sliceConverter) each(b []byte, leftpad uint, isSet bool) error {
	for _, e := range s.elems {
		begin := leftpad / 8
		rightpos := leftpad + uint(e.size())
		end := (rightpos + 7) / 8

		if isSet {
			if err := e.set(b[begin:end], leftpad); err != nil {
				return err
			}
		} else {
			if err := e.get(b[begin:end], leftpad); err != nil {
				return err
			}
		}

		leftpad += uint(e.size())
	}
	return nil

}

// generate struct fields converter map
func mkStructFieldsConvertMap(rv reflect.Value, rt reflect.Type) (map[string]converter, error) {
	fields := make(map[string]converter)
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
				return nil, fmt.Errorf("%s has invalid size %q byte(s)", f.Name, v)
			}
			size *= 8
		} else if v, ok := f.Tag.Lookup("bit"); ok {
			size, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("%s has invalid size %q bit(s)", f.Name, v)
			}
		} else {
			return nil, fmt.Errorf("%s need size hint", f.Name)
		}

		if v, ok := f.Tag.Lookup("len"); ok {
			len, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("%s has invalid length %q", f.Name, v)
			}
		}

		setfunc, err := converterFactory(v, size, len)
		if err != nil {
			return nil, err
		}
		fields[f.Name] = setfunc
	}

	return fields, nil
}

// Read method read data from srcreader and save to dstptr.
// dstptr's type is needed pointer type of struct.
func Read(dstptr interface{}, srcreader io.Reader) error {
	// check
	rv := reflect.ValueOf(dstptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("bitio.Read: want to set pointer type of struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bitio.Read: want to set pointer type of struct")
	}
	rt := rv.Type()

	// generate struct fields converter map
	fields, err := mkStructFieldsConvertMap(rv, rt)
	if err != nil {
		return fmt.Errorf("bitio.Read: %v", err)
	}

	// calc read size
	readsize := 0
	for _, v := range fields {
		readsize += v.size()
	}
	readsize += 7 // padding
	readsize /= 8

	// read from reader
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
		begin := posbit / 8
		end := (posbit + sizebit + 7) / 8
		lpad := uint(posbit - begin*8)
		posbit += sizebit
		if err := fields[f.Name].set(buffer[begin:end], lpad); err != nil {
			return err
		}
	}

	return nil
}

// Write method write data to dstwriter.
// srcptr's type is needed type of struct (or pointer).
func Write(dstwriter io.Writer, srcptr interface{}) error {
	// check
	rv := reflect.ValueOf(srcptr)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem() // *strct -> strct
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bitio.Write: want to set struct")
	}
	rt := rv.Type()

	// generate struct fields converter map
	fields, err := mkStructFieldsConvertMap(rv, rt)
	if err != nil {
		return fmt.Errorf("bitio.Write: %v", err)
	}

	// calc write size
	writesize := 0
	for _, v := range fields {
		writesize += v.size()
	}
	writesize += 7 // padding
	writesize /= 8

	// read from struct
	buffer := make([]byte, writesize)
	posbit := int(0)
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		sizebit := fields[f.Name].size()
		begin := posbit / 8
		end := (posbit + sizebit + 7) / 8
		lpad := uint(posbit - begin*8)
		posbit += sizebit
		if err := fields[f.Name].get(buffer[begin:end], lpad); err != nil {
			return err
		}
	}

	// write to writer
	n, err := dstwriter.Write(buffer)
	if err != nil {
		return fmt.Errorf("bitio.Write: write failed")
	}
	if n != writesize {
		return fmt.Errorf("bitio.Write: write %d bytes, want %d bytes", n, writesize)
	}

	return nil
}

// []byte -> int64
func toLittleEndianInt(b []byte, leftpad, rightpad uint) int64 {
	return int64(toLittleEndianUint(b, leftpad, rightpad))
}

// []byte -> uint64
func toLittleEndianUint(b []byte, leftpad, rightpad uint) uint64 {
	// require: leftpad, rightpad < 8
	w := lBitsShift(b, leftpad)
	padding := leftpad + rightpad
	if padding >= 8 {
		padding -= 8
		w = append([]byte{}, w[:len(w)-1]...)
	}
	w[len(w)-1] >>= padding // alignment

	buf := make([]byte, 8)
	copy(buf, w)

	return binary.LittleEndian.Uint64(buf)
}

// dst = []byte(int64)
func fromLittleEndianInt(dst []byte, leftpad uint, size int, value int64) {
	fromLittleEndianUint(dst, leftpad, size, uint64(value))
}

// dst = []byte(uint64)
func fromLittleEndianUint(dst []byte, leftpad uint, size int, value uint64) {
	// require: leftpad < 8
	bv := make([]byte, 8)
	binary.LittleEndian.PutUint64(bv, value)
	bv = cutSlice(bv, (size+7)/8)

	// allign
	if size%8 != 0 {
		index := size / 8
		bv[index] <<= 8 - uint(size)%8
	}

	rBitsShiftCopy(dst, bv, leftpad)
}

// dst == src << shift
// require: shift < 8
func lBitsShift(src []byte, shift uint) []byte {
	dst := make([]byte, len(src))
	lBitsShiftCopy(dst, src, shift)
	return dst
}

// dst <- src << shift
// require: shift < 8
func lBitsShiftCopy(dst, src []byte, shift uint) {
	mask := byte(0xff >> shift)
	for i, b := range src {
		dst[i] |= (b & mask) << shift
	}

	mask = ^mask
	for i, b := range src[1:] {
		dst[i] |= (b & mask) >> (8 - shift)
	}
}

// dst <- src >> shift
// require: shift < 8
func rBitsShiftCopy(dst, src []byte, shift uint) {
	for i, b := range src {
		dst[i] |= b >> shift
	}

	mask := byte(0xff >> (8 - shift))
	for i, b := range src[:len(src)-1] {
		dst[i+1] |= (b & mask) << (8 - shift)
	}
	if len(dst) > len(src) {
		i := len(src) - 1
		dst[i+1] |= (src[i] & mask) << (8 - shift)
	}
}

// cut slice
func cutSlice(src []byte, length int) []byte {
	if len(src) <= length {
		return src
	}
	return append([]byte{}, src[:length]...)
}
