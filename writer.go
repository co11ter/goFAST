package fast

import (
	"bytes"
	"github.com/kr/pretty"
	"io"
)

type Writer struct {
	writer io.Writer
	buf bytes.Buffer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer, bytes.Buffer{}}
}

// TODO
func (w *Writer) commit() error {
	pretty.Println("bytes: ", w.buf.Bytes())
	//_, err := w.buf.WriteTo(w.writer)
	return nil
}

func (w *Writer) WriteUint32(nullable bool, value *uint32) error {
	if nullable {
		if value == nil {
			return w.buf.WriteByte(0x80)
		}
	} else if *value == 0 {
		return w.buf.WriteByte(0x80)
	}

	b := make([]byte, 5)
	i := 4
	for i >= 0 && *value != 0 {
		b[i] = byte(*value & 0x7F)
		*value >>= 7
		i--
	}

	b[4] |= 0x80
	w.buf.Write(b[i+1:])

	return nil
}

// TODO
func (w *Writer) WriteUint64(nullable bool, value *uint64) error {
	return nil
}

// TODO
func (w *Writer) WriteInt32(nullable bool, value *int32) error {
	return nil
}

// TODO
func (w *Writer) WriteInt64(nullable bool, value *int64) error {
	return nil
}

// TODO
func (w *Writer) WriteUtfString(nullable bool, value *string) error {
	return nil
}

// TODO
func (w *Writer) WriteAsciiString(nullable bool, value *string) error {
	return nil
}

func (w *Writer) WriteNil() error {
	return w.buf.WriteByte(0x80)
}
