// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"io"
)

type buffer interface {
	io.Writer
	Bytes() []byte
	Reset()
}

type Writer struct {
	buf buffer
	strBuf bytes.Buffer
	tmpBuf bytes.Buffer
}

func NewWriter(buf buffer) *Writer {
	return &Writer{buf: buf}
}

func (w *Writer) Bytes() []byte {
	b := append(w.buf.Bytes(), w.tmpBuf.Bytes()...)
	w.buf.Reset()
	w.tmpBuf.Reset()
	return b
}

func (w *Writer) WritePMap(m *pMap) error {
	_, err := w.tmpBuf.Write(w.buf.Bytes())
	w.buf.Reset()
	if err != nil {
		return err
	}

	if m.bitmap == 0 {
		_, err = w.buf.Write([]byte{0x80})
		return err
	}

	b := make([]byte, 8)
	i :=7
	for i >= 0 && m.bitmap != 0 {
		b[i] = byte(m.bitmap)
		m.bitmap >>= 7
		i--
	}
	b[7] |= 0x80

	_, err = w.buf.Write(b[i+1:])
	return err
}

func (w *Writer) WriteUint32(nullable bool, value uint32) error {
	if !nullable && value == 0 {
		w.buf.Write([]byte{0x80})
		return nil
	}

	if nullable && value > 0 {
		value++
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
		w.buf.Write([]byte{0x80})
		return nil
	}

	if nullable && value > 0 {
		value++
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
		w.buf.Write([]byte{0x80})
		return nil
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
		w.buf.Write([]byte{0x80})
		return nil
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

func (w *Writer) WriteAsciiString(nullable bool, value string) error {
	if len(value) == 0 {
		if nullable {
			w.buf.Write([]byte{0x00})
			return nil
		}
		w.buf.Write([]byte{0x00})
		return nil
	}

	if len(value) == 1 && value[0] == 0x00 {
		if nullable {
			w.buf.Write([]byte{0x00, 0x00, 0x80})
		} else {
			w.buf.Write([]byte{0x00, 0x80})
		}
		return nil
	}

	w.strBuf.WriteString(value[:len(value)-1])
	w.buf.Write(w.strBuf.Bytes())
	w.strBuf.Reset()
	w.buf.Write([]byte{value[len(value)-1] | 0x80})
	return nil
}

func (w *Writer) WriteNil() error {
	w.buf.Write([]byte{0x80})
	return nil
}
