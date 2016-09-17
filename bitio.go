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

	buffer := make([]byte, readsize)
	n, err := srcreader.Read(buffer)
	if err != nil {
		return fmt.Errorf("bitio.Read: read data from srcreader failed")
	}
	if n != readsize {
		return fmt.Errorf("bitio.Read: read %d bytes, want %d bytes", n, readsize)
	}

	// save to struct
	pos := 0
	for i := 0; i < rv.NumField(); i++ {
		f := rt.Field(i)
		v := rv.Field(i)

		if f.PkgPath != "" {
			// unexport field
			continue
		}

		size := fields[f.Name]
		value := int64(0)
		digit := uint(0)
		for _, b := range buffer[pos : pos+size] {
			// little endian
			value |= int64(b) << digit
			digit += 8
		}
		pos += size

		v.SetInt(value)
	}

	return nil
}
