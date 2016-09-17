package bitio

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// Read method read data from srcreader and save to dstptr.
// dstptr's type is needed pointer type of struct.
func Read(dstptr interface{}, srcreader io.Reader) error {
	rv := reflect.ValueOf(dstptr)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("bitio.Read: dstptr need to set pointer type of struct")
	}

	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("bitio.Read: dstptr need to set pointer type of struct")
	}
	rt := rv.Type()

	// read each filed size
	fields := make(map[string]int)
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		if v, ok := f.Tag.Lookup("byte"); ok {
			size, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("bitio.Read: %s has invalid size %q byte(s)", f.Name, v)
			}
			fields[f.Name] = size * 8
		} else if v, ok := f.Tag.Lookup("bit"); ok {
			size, err := strconv.Atoi(v)
			if err != nil {
				return fmt.Errorf("bitio.Read: %s has invalid size %q bit(s)", f.Name, v)
			}
			fields[f.Name] = size
		} else {
			return fmt.Errorf("bitio.Read: %s need size hint", f.Name)
		}
	}

	// read from reader
	readsize := 0
	for _, v := range fields {
		readsize += v
	}
	readsize += 7 // padding
	readsize /= 8

	buffer := make([]byte, readsize)
	n, err := srcreader.Read(buffer)
	if err != nil {
		return fmt.Errorf("bitio.Read: read data from srcreader failed")
	}
	if n != readsize {
		return fmt.Errorf("bitio.Read: read %d bytes, want %d bytes", n, readsize)
	}

	// save to struct
	posbit := int(0)
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)
		v := rv.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		sizebit := fields[f.Name]
		lpos := posbit / 8
		rpos := (posbit + sizebit + 7) / 8
		lpad := uint(posbit - lpos*8)
		rpad := uint(rpos*8 - posbit - sizebit)
		posbit += sizebit
		value := toLittleEndianInt(buffer[lpos:rpos], lpad, rpad)
		v.SetInt(value)
	}

	return nil
}

func toLittleEndianInt(b []byte, leftpad, rightpad uint) int64 {
	// require: leftpad, rightpad < 8
	w := lBitsShift(b, leftpad)
	padding := leftpad + rightpad
	if padding >= 8 {
		padding -= 8
		w = append([]byte{}, w[:len(w)-1]...)
	}
	w[len(w)-1] >>= padding // alignment

	value := int64(0)
	digit := uint(0)

	for _, w := range w {
		// little endian
		value |= int64(w) << digit
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
