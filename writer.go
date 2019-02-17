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

type writer struct {
	buf buffer
	strBuf bytes.Buffer
	tmpBuf bytes.Buffer
}

func newWriter(buf buffer) *writer {
	return &writer{buf: buf}
}

func (w *writer) Bytes() []byte {
	b := append(w.buf.Bytes(), w.tmpBuf.Bytes()...)
	w.buf.Reset()
	w.tmpBuf.Reset()
	return b
}

func (w *writer) WritePMap(m *pMap) error {
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

func (w *writer) WriteUint32(nullable bool, value uint32) error {
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

func (w *writer) WriteUint64(nullable bool, value uint64) error {
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

func (w *writer) writeInt(nullable bool, value int64, size int) error {
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

	b := make([]byte, size+2)
	i := size
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

	b[size] |= 0x80
	w.buf.Write(b[i:size+1])
	return nil
}

func (w *writer) WriteInt32(nullable bool, value int32) error {
	return w.writeInt(nullable, int64(value), 4)
}

func (w *writer) WriteInt64(nullable bool, value int64) error {
	return w.writeInt(nullable, value, 9)
}

// TODO
func (w *writer) WriteUnicodeString(nullable bool, value string) error {
	return nil
}

func (w *writer) WriteASCIIString(nullable bool, value string) error {
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

func (w *writer) WriteNil() error {
	w.buf.Write([]byte{0x80})
	return nil
}
