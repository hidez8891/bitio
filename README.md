# bitio

bitio is Golang library for bit reader/writer.

## Support Type

* int (int8 - int64)
* uint (uint8 - uint64)
* string (fixed length)
* slice (fixed length)

## Example

### BitField Reader/Writer

```go
type Container struct {
	Sign byte   `bit:"4"`           // 4bit
	Size int    `bit:"4"`           // 4bit
	Data []byte `byte:"1" len:"10"` // 1byte x 10
}

func ReadContainer(r io.Reader) (*Container, error) {
	c := &Container{}

	br := bitio.NewBitFieldReader(r)
	if _, err := br.Read(c); err != nil {
		return nil, err
	}

	return c, nil
}

func WriteContainer(w io.Writer, c *Container) (nBit int, err error) {
	bw := bitio.NewBitFieldWriter(w)
	if nBit, err = bw.Write(c); err != nil {
		return
	}

	return
}
```

### Bit Reader/Writer

```go
func ReadBit(r io.Reader) {
	br := bitio.NewBitReadBuffer(r)

	//read 1bit
	var b1 byte
	br.ReadBit(&b1, 1)

	//read 10bit
	b2 := make([]byte, 2)
	br.ReadBits(b2, 10)

	//read 2byte(16bit)
	b3 := make([]byte, 2)
	br.Read(b3)
}

func WriteBit(w io.Writer) {
	bw := bitio.NewBitWriteBuffer(w)

	//write 1bit
	bw.WriteBit(0x01, 1)

	//write 10bit
	b2 := []byte{0x01, 0x02}
	bw.WriteBits(b2, 10)

	//write 2byte(16bit)
	b3 := []byte{0x01, 0x02}
	bw.Write(b3)
}
```
