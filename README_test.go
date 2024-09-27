package bitio_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/hidez8891/bitio"
)

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

////////////////////////////////////////////////////////////////////////////////

func Example_ReadContainer() {
	r := bytes.NewReader([]byte{
		0x12, 0x34, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66,
		0x67, 0x68, 0xc1, 0xc2, 0xc3, 0xc4, 0xf1, 0xf2,
		0xf3, 0xf4,
	})

	c, _ := ReadContainer(r)

	fmt.Printf("Sign = %s\n", hex.EncodeToString(c.Sign))
	fmt.Printf("Size = %d\n", c.Size)
	fmt.Printf("Name = %s\n", c.Name)
	fmt.Printf("Data = %s\n", hex.EncodeToString(c.Data))
	fmt.Printf("CRC  = %x\n", c.CRC)

	// Output:
	// Sign = 010203
	// Size = 4
	// Name = abcdefgh
	// Data = c1c2c3c4
	// CRC  = f1f2f3f4
}

func Example_WriteContainer() {
	c := &Container{
		Sign: []byte{0x01, 0x02, 0x03},
		Size: 4,
		Name: "abcdefgh",
		Data: []byte{0xc1, 0xc2, 0xc3, 0xc4},
		CRC:  0xf1f2f3f4,
	}
	w := new(bytes.Buffer)

	WriteContainer(w, c)

	b := w.Bytes()
	fmt.Printf("b = %s\n", hex.EncodeToString(b[0:8]))
	fmt.Printf("  = %s\n", hex.EncodeToString(b[8:16]))
	fmt.Printf("  = %s\n", hex.EncodeToString(b[16:]))

	// Output:
	// b = 1234616263646566
	//   = 6768c1c2c3c4f1f2
	//   = f3f4
}

func Example_ReadBit() {
	r := bytes.NewReader([]byte{0x12, 0x34, 0x56, 0x78})

	b1, b2, b3 := ReadBit(r)

	fmt.Printf("b1 = %02x\n", b1)
	fmt.Printf("b2 = %s\n", hex.EncodeToString(b2))
	fmt.Printf("b3 = %s\n", hex.EncodeToString(b3))

	// Output:
	// b1 = 01
	// b2 = 0234
	// b3 = 5678
}

func Example_WriteBit() {
	w := new(bytes.Buffer)

	WriteBit(w)

	fmt.Printf("w = %s\n", hex.EncodeToString(w.Bytes()))

	// Output:
	// w = 12345678
}
