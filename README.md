# bitio

bitio is Golang library for reading/writing bit field.

## Support Type

* int (int8 - int64)
* uint (uint8 - uint64)
* string (fixed length)
* slice (fixed length)

## Example

```go
type Container struct {
	Sign byte   `bit:"4"`           // 4bit
	Size int    `bit:"4"`           // 4bit
	Data []byte `byte:"1" len:"10"` // 1byte x 10
}

func Read(r io.Reader) Container {
	c := Container{}
	bitio.Read(&c, r)
	return c
}

func Write(w io.Writer, c Container) {
	bitio.Write(w, &c)
}
```
