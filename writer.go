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

func (w *Writer) WriteUint32(nullable bool, value uint32) error {
	if !nullable && value == 0 {
		return w.buf.WriteByte(0x80)
	}

	b := make([]byte, 5)
	i := 4
	for i >= 0 && value != 0 {
		b[i] = byte(value & 0x7F)
		value >>= 7
		i--
	}

	b[4] |= 0x80
	w.buf.Write(b[i+1:])

	return nil
}

func (w *Writer) WriteUint64(nullable bool, value uint64) error {
	if !nullable && value == 0 {
		return w.buf.WriteByte(0x80)
	}

	b := make([]byte, 10)
	i := 9
	for i >= 0 && value != 0 {
		b[i] = byte(value & 0x7F)
		value >>= 7
		i--
	}

	b[9] |= 0x80
	w.buf.Write(b[i+1:])

	return nil
}

func (w *Writer) WriteInt32(nullable bool, value int32) error {
	if !nullable && value == 0 {
		return w.buf.WriteByte(0x80)
	}

	positive := value > 0

	if nullable && positive {
		value++
	}

	var sign int32
	if value <= 0 {
		sign = -1
	}

	b := make([]byte, 5)
	i := 4
	for i >= 0 && value != sign {
		b[i] = byte(value & 0x7F)
		value >>= 7
		i--
	}

	i++
	if positive {
		if (b[i] & 0x40) > 0 {
			i--
			b[i] = 0x00
		}
	} else if (b[i] & 0x40) == 0 {
		i--
		b[i] = 0x7F
	}

	b[4] |= 0x80
	w.buf.Write(b[i+1:])

	return nil
}

func (w *Writer) WriteInt64(nullable bool, value int64) error {
	if !nullable && value == 0 {
		return w.buf.WriteByte(0x80)
	}

	positive := value > 0

	if nullable && positive {
		value++
	}

	var sign int64
	if value <= 0 {
		sign = -1
	}

	b := make([]byte, 10)
	i := 9
	for i >= 0 && value != sign {
		b[i] = byte(value & 0x7F)
		value >>= 7
		i--
	}

	i++
	if positive {
		if (b[i] & 0x40) > 0 {
			i--
			b[i] = 0x00
		}
	} else if (b[i] & 0x40) == 0 {
		i--
		b[i] = 0x7F
	}

	b[9] |= 0x80
	w.buf.Write(b[i+1:])

	return nil
}

// TODO
func (w *Writer) WriteUtfString(nullable bool, value string) error {
	return nil
}

// TODO
func (w *Writer) WriteAsciiString(nullable bool, value string) error {
	if len(value) == 0 {
		if nullable {
			w.buf.WriteByte(0x00)
		}
		return w.buf.WriteByte(0x80)
	}

	if len(value) == 1 && value[0] == 0x00 {
		if nullable {
			w.buf.Write([]byte{0x00, 0x00, 0x80})
		} else {
			w.buf.Write([]byte{0x00, 0x80})
		}
		return nil
	}

	w.buf.WriteString(value[:len(value)-1])
	return w.buf.WriteByte(value[len(value)-1] | 0x80)
}

func (w *Writer) WriteNil() error {
	return w.buf.WriteByte(0x80)
}
