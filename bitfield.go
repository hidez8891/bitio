package bitio

import (
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

// Read reads data and returns read size.
// If error happen, err will be set.
func (obj *BitFieldReader) Read(p []byte) (int, error) {
	return obj.r.Read(p)
}

// ReadStruct reads bit-field data and returns read size.
// If error happen, err will be set.
func (obj *BitFieldReader) ReadStruct(p interface{}) (nBit int, err error) {
	// check argument type
	var rv reflect.Value
	if rv = reflect.ValueOf(p); rv.Kind() != reflect.Ptr {
		err = fmt.Errorf("ReadStruct: argument wants to pointer of struct")
		return
	}
	if rv = rv.Elem(); rv.Kind() != reflect.Struct {
		err = fmt.Errorf("ReadStruct: argument wants to pointer of struct")
		return
	}
	rt := rv.Type()

	// read bit-fields
	fieldValue := make(map[string]int)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		ptr := rv.Field(i)

		// skip unexport field
		if field.PkgPath != "" {
			continue
		}

		// get field configration
		var config *fieldConfig
		if config, err = getFieldConfig(ptr, field, fieldValue); err != nil {
			return
		}

		// read bit-filed
		var n int
		if n, err = readField(obj.r, config); err != nil {
			return
		}
		nBit += n

		// save field's value (ex: length's variable)
		switch field.Type.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldValue[field.Name] = int(ptr.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldValue[field.Name] = int(ptr.Uint())
		default:
			// unsave no number
		}
	}

	return
}

////////////////////////////////////////////////////////////////////////////////

// NewBitFieldWriter returns BitFieldWriter
func NewBitFieldWriter(w io.Writer) *BitFieldWriter {
	return NewBitFieldWriter2(NewBitWriteBuffer(w))
}

// NewBitFieldWriter2 returns BitFieldWriter
func NewBitFieldWriter2(w BitWriter) *BitFieldWriter {
	return &BitFieldWriter{
		w: w,
	}
}

// BitFieldWriter write bit-field data.
type BitFieldWriter struct {
	w BitWriter
}

// Write writes data len(p) size and returns write size.
// If error happen, err will be set.
func (obj *BitFieldWriter) Write(p []byte) (int, error) {
	return obj.w.Write(p)
}

// WriteStruct writes bit-field data and returns write size.
// If error happen, err will be set.
func (obj *BitFieldWriter) WriteStruct(p interface{}) (nBit int, err error) {
	// check argument type
	var rv reflect.Value
	if rv = reflect.ValueOf(p); rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		err = fmt.Errorf("WriteStruct: argument wants to struct")
	}
	rt := rv.Type()

	fieldValue := make(map[string]int)

	// save slice length for length's variable
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		ptr := rv.Field(i)

		// skip unexport field
		if field.PkgPath != "" {
			continue
		}

		// save slice length
		switch field.Type.Kind() {
		case reflect.Slice:
			if v, ok := field.Tag.Lookup("len"); ok {
				if _, err = strconv.Atoi(v); err != nil {
					fieldValue[v] = ptr.Len()
				}
			}
		default:
			// nothing to do
		}
	}

	// write bit-fields
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		ptr := rv.Field(i)

		// skip unexport field
		if field.PkgPath != "" {
			continue
		}

		// get field configration
		var config *fieldConfig
		if config, err = getFieldConfig(ptr, field, fieldValue); err != nil {
			return
		}

		// update field's value (ex: length's variable)
		if val, ok := fieldValue[field.Name]; ok {
			switch field.Type.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				ptr.SetInt(int64(val))
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				ptr.SetUint(uint64(val))
			default:
				// nothing to do
			}
		}

		// write bit-filed
		var n int
		if n, err = writeField(obj.w, config); err != nil {
			return
		}
		nBit += n
	}

	return
}

// Flush writes data if BitWriter is not empty.
// If error happen, err will be set.
func (obj *BitFieldWriter) Flush() error {
	return obj.w.Flush()
}

////////////////////////////////////////////////////////////////////////////////

// fieldConfig store bit-field configration.
type fieldConfig struct {
	ptr    reflect.Value
	bits   int
	len    int
	endian ByteOrder
}

func getFieldConfig(ptr reflect.Value, field reflect.StructField, fieldValue map[string]int) (*fieldConfig, error) {
	var err error

	// bit-field size
	bits := 0
	if v, ok := field.Tag.Lookup("byte"); ok {
		if bits, err = strconv.Atoi(v); err != nil {
			return nil, fmt.Errorf("%s has invalid size %q byte(s)", field.Name, v)
		}
		bits *= 8
	} else if v, ok := field.Tag.Lookup("bit"); ok {
		if bits, err = strconv.Atoi(v); err != nil {
			return nil, fmt.Errorf("%s has invalid size %q bit(s)", field.Name, v)
		}
	} else {
		return nil, fmt.Errorf("%s need size hint", field.Name)
	}

	// bit-field block count
	len := 0
	if v, ok := field.Tag.Lookup("len"); ok {
		if len, err = strconv.Atoi(v); err == nil {
			// nothing to do
		} else if val, ok := fieldValue[v]; ok {
			// load length's variable
			len = val
		} else {
			return nil, fmt.Errorf("%s has invalid length %q", field.Name, v)
		}
	}

	// bit-field endian
	endian := LittleEndian
	if v, ok := field.Tag.Lookup("endian"); ok {
		switch v {
		case "big":
			endian = BigEndian
		case "little":
			endian = LittleEndian
		default:
			return nil, fmt.Errorf("%s has invalid endian %q", field.Name, v)
		}
	}

	config := &fieldConfig{
		ptr:    ptr,
		bits:   bits,
		len:    len,
		endian: endian,
	}
	return config, nil
}

////////////////////////////////////////////////////////////////////////////////

func readField(r BitReader, config *fieldConfig) (n int, err error) {
	if config.bits < 1 {
		err = fmt.Errorf("invalid bit-field size %d byte(s)", config.bits)
		return
	}

	switch config.ptr.Kind() {
	case reflect.Bool:
		{
			var v int
			if err = Read(r, config.bits, config.endian, &v); err != nil {
				return
			}
			n = config.bits
			config.ptr.SetBool(v != 0)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			var v int64
			if err = Read(r, config.bits, config.endian, &v); err != nil {
				return
			}
			n = config.bits
			config.ptr.SetInt(v)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			var v uint64
			if err = Read(r, config.bits, config.endian, &v); err != nil {
				return
			}
			n = config.bits
			config.ptr.SetUint(v)
		}

	case reflect.String:
		{
			if config.bits%8 != 0 {
				err = fmt.Errorf("string type size needs to 8*n bits, set %d bits", config.bits)
				return
			}

			v := make([]byte, config.bits/8)
			if err = ReadSlice(r, 8, config.endian, v); err != nil {
				return
			}
			n = config.bits
			config.ptr.SetString(string(v))
		}

	case reflect.Slice:
		{
			if config.len < 1 {
				err = fmt.Errorf("slice type needs positive length, set %d length", config.len)
				return
			}

			// (re-)allocate slice space
			if config.ptr.Len() < config.len {
				rv := reflect.MakeSlice(config.ptr.Type(), config.len, config.len)
				reflect.Copy(rv, config.ptr)
				config.ptr.Set(rv)
			}

			// read slice elements
			for i := 0; i < config.len; i++ {
				_, err = readField(r, &fieldConfig{
					ptr:    config.ptr.Index(i),
					bits:   config.bits,
					endian: config.endian,
				})
				if err != nil {
					return
				}
			}

			n = config.bits * config.len
		}

	default:
		{
			err = fmt.Errorf("unsupport bit-filed type %q", config.ptr.Kind().String())
			return
		}
	}

	return
}

func writeField(w BitWriter, config *fieldConfig) (n int, err error) {
	if config.bits < 1 {
		err = fmt.Errorf("invalid bit-field size %d byte(s)", config.bits)
		return
	}

	switch config.ptr.Kind() {
	case reflect.Bool:
		{
			v := 0
			if config.ptr.Bool() {
				v = 1
			}

			if err = Write(w, config.bits, config.endian, v); err != nil {
				return
			}
			n = config.bits
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			v := int64(config.ptr.Int())
			if err = Write(w, config.bits, config.endian, v); err != nil {
				return
			}
			n = config.bits
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			v := uint64(config.ptr.Uint())
			if err = Write(w, config.bits, config.endian, v); err != nil {
				return
			}
			n = config.bits
		}

	case reflect.String:
		{
			if config.bits%8 != 0 {
				err = fmt.Errorf("string type size needs to 8*n bits, set %d bits", config.bits)
				return
			}
			size := config.bits / 8

			v := []byte(config.ptr.String())
			if len(v) < size {
				v = append(v, make([]byte, size-len(v))...)
			}

			if err = WriteSlice(w, 8, config.endian, v[:size]); err != nil {
				return
			}
			n = config.bits
		}

	case reflect.Slice:
		{
			if config.len < 1 {
				err = fmt.Errorf("slice type needs positive length, set %d length", config.len)
				return
			}

			// (re-)allocate slice space
			if config.ptr.Len() < config.len {
				rv := reflect.MakeSlice(config.ptr.Type(), config.len, config.len)
				reflect.Copy(rv, config.ptr)
				config.ptr.Set(rv)
			}

			// write slice elements
			for i := 0; i < config.len; i++ {
				_, err = writeField(w, &fieldConfig{
					ptr:    config.ptr.Index(i),
					bits:   config.bits,
					endian: config.endian,
				})
				if err != nil {
					return
				}
			}

			n = config.bits * config.len
		}

	default:
		{
			err = fmt.Errorf("unsupport bit-filed type %q", config.ptr.Kind().String())
			return
		}
	}

	return
}
