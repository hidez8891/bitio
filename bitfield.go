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
		err = fmt.Errorf("Read: argument wants to pointer of struct")
		return
	}
	if rv = rv.Elem(); rv.Kind() != reflect.Struct {
		err = fmt.Errorf("Read: argument wants to pointer of struct")
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

		// read field configration
		var config *fieldConfig
		if config, err = readFieldConfig(ptr, field, fieldValue); err != nil {
			return
		}

		// read bit-filed
		var (
			r fieldReader
			n int
		)
		if r, err = newFieldReader(obj.r, config); err != nil {
			return
		}
		if n, err = r.read(); err != nil {
			return
		}
		nBit += n

		// save value
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
		err = fmt.Errorf("Write: argument wants to struct")
	}
	rt := rv.Type()

	// write bit-fields
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		ptr := rv.Field(i)

		// skip unexport field
		if field.PkgPath != "" {
			continue
		}

		// read field configration
		var config *fieldConfig
		if config, err = readFieldConfig(ptr, field, nil); err != nil {
			return
		}

		// write bit-filed
		var (
			w fieldWriter
			n int
		)
		if w, err = newFieldWriter(obj.w, config); err != nil {
			return
		}
		if n, err = w.write(); err != nil {
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

type fieldReader interface {
	read() (nBit int, err error)
}

func newFieldReader(r BitReader, config *fieldConfig) (fr fieldReader, err error) {
	if config.bits < 1 {
		return nil, fmt.Errorf("invalid bit-field size %d byte(s)", config.bits)
	}

	switch config.ptr.Kind() {
	case reflect.Bool:
		fr = newFieldBoolReader(r, config)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fr = newFieldIntReader(r, config)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fr = newFieldUintReader(r, config)
	case reflect.String:
		fr = newFieldStringReader(r, config)
	case reflect.Slice:
		if config.len < 1 {
			return nil, fmt.Errorf("Slice type needs positive length")
		}
		fr = newFieldSliceReader(r, config)
	default:
		return nil, fmt.Errorf("Not support bit-filed type %q", config.ptr.Kind().String())
	}

	return
}

// fieldBoolReader implemented fieldReader for Bool type
type fieldBoolReader struct {
	r *fieldUintReader
}

func newFieldBoolReader(r BitReader, config *fieldConfig) *fieldBoolReader {
	return &fieldBoolReader{
		r: newFieldUintReader(r, config),
	}
}

func (obj *fieldBoolReader) read() (nBit int, err error) {
	var value uint64

	value, nBit, err = obj.r.readValue()
	if err == nil {
		obj.r.ptr.SetBool(value != 0)
	}

	return
}

// fieldIntReader implemented fieldReader for Integer type
type fieldIntReader struct {
	r *fieldUintReader
}

func newFieldIntReader(r BitReader, config *fieldConfig) *fieldIntReader {
	return &fieldIntReader{
		r: newFieldUintReader(r, config),
	}
}

func (obj *fieldIntReader) read() (nBit int, err error) {
	var value uint64

	value, nBit, err = obj.r.readValue()
	if err == nil {
		obj.r.ptr.SetInt(int64(value))
	}

	return
}

// fieldUintReader implemented fieldReader for Unsigned Integer type
type fieldUintReader struct {
	r      BitReader
	ptr    reflect.Value
	bits   int
	endian int
}

func newFieldUintReader(r BitReader, config *fieldConfig) *fieldUintReader {
	return &fieldUintReader{
		r:      r,
		ptr:    config.ptr,
		bits:   config.bits,
		endian: config.endian,
	}
}

func (obj *fieldUintReader) readValue() (value uint64, nBit int, err error) {
	if obj.bits > 64 {
		err = fmt.Errorf("bit-field size needs <= 64bit")
		return
	}

	buf := make([]byte, 8)
	if nBit, err = obj.r.ReadBits(buf, obj.bits); err != nil {
		return
	}

	if obj.endian == endianLittle {
		// little endian shift
		// 12bit: 0x0123 -> 0x1230 -> 0x1203
		if nBit%8 > 0 {
			leftShift(buf, uint(8-nBit%8))
			buf[nBit/8] >>= uint(8 - nBit%8)
		}

		value = binary.LittleEndian.Uint64(buf)
	} else {
		// big endian shift
		// 12bit: 0x0123**** -> 0x****1230
		nByte := (nBit + 7) / 8
		rightShift(buf, uint(8*(8-nByte)))

		value = binary.BigEndian.Uint64(buf)
	}

	return
}

func (obj *fieldUintReader) read() (nBit int, err error) {
	var value uint64

	value, nBit, err = obj.readValue()
	if err == nil {
		obj.ptr.SetUint(value)
	}

	return
}

// fieldStringReader implemented fieldReader for String type
type fieldStringReader struct {
	r    BitReader
	ptr  reflect.Value
	bits int
}

func newFieldStringReader(r BitReader, config *fieldConfig) *fieldStringReader {
	return &fieldStringReader{
		r:    r,
		ptr:  config.ptr,
		bits: config.bits,
	}
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
	r      BitReader
	ptr    reflect.Value
	bits   int
	len    int
	endian int
}

func newFieldSliceReader(r BitReader, config *fieldConfig) *fieldSliceReader {
	return &fieldSliceReader{
		r:      r,
		ptr:    config.ptr,
		bits:   config.bits,
		len:    config.len,
		endian: config.endian,
	}
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

		ptr := obj.ptr.Index(i)
		if r, err = newFieldReader(obj.r, &fieldConfig{ptr, obj.bits, 0, obj.endian}); err != nil {
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

const (
	endianBig    = 0
	endianLittle = 1
)

// fieldConfig store bit-field configration.
type fieldConfig struct {
	ptr    reflect.Value
	bits   int
	len    int
	endian int
}

func readFieldConfig(ptr reflect.Value, field reflect.StructField, fieldValue map[string]int) (*fieldConfig, error) {
	var err error

	if fieldValue == nil {
		fieldValue = make(map[string]int) // dummy
	}

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
			// OK
		} else if val, ok := fieldValue[v]; ok {
			// OK
			len = val
		} else {
			return nil, fmt.Errorf("%s has invalid length %q", field.Name, v)
		}
	}

	// bit-field endian
	endian := endianLittle
	if v, ok := field.Tag.Lookup("endian"); ok {
		switch v {
		case "big":
			endian = endianBig
		case "little":
			endian = endianLittle
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

type fieldWriter interface {
	write() (nBit int, err error)
}

func newFieldWriter(w BitWriter, config *fieldConfig) (fw fieldWriter, err error) {
	if config.bits < 1 {
		return nil, fmt.Errorf("invalid bit-field size %d byte(s)", config.bits)
	}

	switch config.ptr.Kind() {
	case reflect.Bool:
		fw = newFieldBoolWriter(w, config)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fw = newFieldIntWriter(w, config)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fw = newFieldUintWriter(w, config)
	case reflect.String:
		fw = newFieldStringWriter(w, config)
	case reflect.Slice:
		if config.len < 1 {
			return nil, fmt.Errorf("Slice type needs positive length")
		}
		fw = newFieldSliceWriter(w, config)
	default:
		return nil, fmt.Errorf("Not support bit-filed type %q", config.ptr.Kind().String())
	}

	return
}

// fieldBoolWriter implemented fieldWriter for Bool type
type fieldBoolWriter struct {
	w *fieldUintWriter
}

func newFieldBoolWriter(w BitWriter, config *fieldConfig) *fieldBoolWriter {
	return &fieldBoolWriter{
		w: newFieldUintWriter(w, config),
	}
}

func (obj *fieldBoolWriter) write() (nBit int, err error) {
	value := 0
	if obj.w.ptr.Bool() == true {
		value = 1
	}
	return obj.w.writeValue(uint64(value))
}

// fieldIntWriter implemented fieldWriter for Integer type
type fieldIntWriter struct {
	w *fieldUintWriter
}

func newFieldIntWriter(w BitWriter, config *fieldConfig) *fieldIntWriter {
	return &fieldIntWriter{
		w: newFieldUintWriter(w, config),
	}
}

func (obj *fieldIntWriter) write() (nBit int, err error) {
	return obj.w.writeValue(uint64(obj.w.ptr.Int()))
}

// fieldUintWriter implemented fieldWriter for Unsigned Integer type
type fieldUintWriter struct {
	w      BitWriter
	ptr    reflect.Value
	bits   int
	endian int
}

func newFieldUintWriter(w BitWriter, config *fieldConfig) *fieldUintWriter {
	return &fieldUintWriter{
		w:      w,
		ptr:    config.ptr,
		bits:   config.bits,
		endian: config.endian,
	}
}

func (obj *fieldUintWriter) writeValue(value uint64) (nBit int, err error) {
	if obj.bits > 64 {
		err = fmt.Errorf("bit-field size needs <= 64bit")
		return
	}
	buf := make([]byte, 8)

	if obj.endian == endianLittle {
		binary.LittleEndian.PutUint64(buf, value)

		// little endian shift
		// 12bit: 0x1203 -> 0x1230 -> 0x0123
		if obj.bits%8 > 0 {
			buf[obj.bits/8] <<= uint(8 - obj.bits%8)
			rightShift(buf, uint(8-obj.bits%8))
		}
	} else {
		binary.BigEndian.PutUint64(buf, value)

		// big endian shift
		// 12bit: 0x****0123 -> 0x0123****
		nByte := (obj.bits + 7) / 8
		leftShift(buf, uint(8*(8-nByte)))
	}

	if nBit, err = obj.w.WriteBits(buf, obj.bits); err != nil {
		return
	}

	return
}

func (obj *fieldUintWriter) write() (nBit int, err error) {
	return obj.writeValue(obj.ptr.Uint())
}

// fieldStringWriter implemented fieldWriter for String type
type fieldStringWriter struct {
	w    BitWriter
	ptr  reflect.Value
	bits int
}

func newFieldStringWriter(w BitWriter, config *fieldConfig) *fieldStringWriter {
	return &fieldStringWriter{
		w:    w,
		ptr:  config.ptr,
		bits: config.bits,
	}
}

func (obj *fieldStringWriter) write() (nBit int, err error) {
	if obj.bits%8 != 0 {
		err = fmt.Errorf("String type size needs to 8*n bits")
		return
	}

	buf := []byte(obj.ptr.String())
	if len(buf)*8 < obj.bits {
		// padding
		buf = append(buf, make([]byte, obj.bits/8-len(buf))...)
	}
	if len(buf)*8 > obj.bits {
		err = fmt.Errorf("String size [%d] is over write size [%d]", len(buf)*8, obj.bits)
		return
	}

	if nBit, err = obj.w.Write(buf); err != nil {
		nBit *= 8
		return
	}
	nBit *= 8
	return
}

// fieldSliceWriter implemented fieldWriter for Slice type
type fieldSliceWriter struct {
	w      BitWriter
	ptr    reflect.Value
	bits   int
	len    int
	endian int
}

func newFieldSliceWriter(w BitWriter, config *fieldConfig) *fieldSliceWriter {
	return &fieldSliceWriter{
		w:      w,
		ptr:    config.ptr,
		bits:   config.bits,
		len:    config.len,
		endian: config.endian,
	}
}

func (obj *fieldSliceWriter) write() (nBit int, err error) {
	if obj.len < 1 {
		err = fmt.Errorf("Slice type needs positive length")
		return
	}

	// write slice bit-fields
	for i := 0; i < obj.len; i++ {
		var (
			w fieldWriter
			n int
		)

		ptr := obj.ptr.Index(i)
		if w, err = newFieldWriter(obj.w, &fieldConfig{ptr, obj.bits, 0, obj.endian}); err != nil {
			return
		}

		if n, err = w.write(); err != nil {
			return
		}
		nBit += n
	}

	return
}
