# bitio

bitio is Golang library for bit reader/writer.

## Support Type

- bool
- int (int8 - int64)
- uint (uint8 - uint64)
- string (fixed size)
- array (fixed length)
- slice (variable length)

## Syntax

| Type              | Syntax         | Description                                   |
| ----------------- | -------------- | --------------------------------------------- |
| field size (bit)  | `bit:"1"`      | value size is 1 bit.                          |
| field size (byte) | `byte:"2"`     | value size is 2 bytes.                        |
| array length      | `len:"3"`      | array is composed of 3 values.                |
| slice length      | `len:"Len"`    | slice is composed of `Len` values.            |
| endianness        | `endian:"big"` | value is big-endian. (default: little-endian) |

## Example

### BitField Reader/Writer

```go
type Container struct {
	Sign []byte `bit:"4" len:"3"`       // 4bit x 3
	Size int    `bit:"4"`               // 4bit
	Name string `byte:"8"`              // 8byte (8chars)
	Data []byte `byte:"1" len:"Size"`   // 1byte x Size
	CRC  uint   `bit:"32" endian:"big"` // 32bit (4byte), big endian
}

func ReadContainer(r io.Reader) (*Container, error) {
	c := &Container{}

	br := bitio.NewBitFieldReader(r)
	if _, err := br.ReadStruct(c); err != nil {
		return nil, err
	}

	return c, nil
}

func WriteContainer(w io.Writer, c *Container) (nBit int, err error) {
	bw := bitio.NewBitFieldWriter(w)
	if nBit, err = bw.WriteStruct(c); err != nil {
		return
	}

	return
}
```

### Bit Reader/Writer

```go
func ReadBit(r io.Reader) (byte, []byte, []byte) {
	br := bitio.NewBitReadBuffer(r)

	//read 4bit
	var b1 byte
	br.ReadBit(&b1, 4)

	//read 12bit
	b2 := make([]byte, 2)
	br.ReadBits(b2, 12)

	//read 2byte(16bit)
	b3 := make([]byte, 2)
	br.Read(b3)

	return b1, b2, b3
}

func WriteBit(w io.Writer) {
	bw := bitio.NewBitWriteBuffer(w)

	//write 4bit
	bw.WriteBit(0x01, 4)

	//write 12bit
	b2 := []byte{0x02, 0x34}
	bw.WriteBits(b2, 12)

	//write 2byte(16bit)
	b3 := []byte{0x56, 0x78}
	bw.Write(b3)

	bw.Flush()
}
```
