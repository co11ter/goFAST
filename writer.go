// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"io"
)

const (
	maxSize32 = 4*8/7
	maxSize64 = 8*8/7
)

type buffer interface {
	io.Writer
	io.WriterTo
	Bytes() []byte
	Reset()
}

type writer struct {
	dataBuf buffer
	pMapBuf buffer

	strBuf bytes.Buffer
}

func newWriter(dataBuf, pMapBuf buffer) *writer {
	return &writer{dataBuf: dataBuf, pMapBuf: pMapBuf}
}

func (w *writer) Write(p []byte) (int, error) {
	return w.dataBuf.Write(p)
}

func (w *writer) WriteTo(writer io.Writer) {
	_, _ = w.pMapBuf.WriteTo(writer)
	_, _ = w.dataBuf.WriteTo(writer)
}

func (w *writer) Reset() {
	w.pMapBuf.Reset()
	w.dataBuf.Reset()
}

func (w *writer) WritePMap(m *pMap) error {
	var err error
	if m.bitmap == 0 {
		_, err = w.pMapBuf.Write([]byte{0x80})
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

	_, err = w.pMapBuf.Write(b[i+1:])
	return err
}

func (w *writer) WriteUint(nullable bool, value uint64, size int) (err error) {
	if !nullable && value == 0 {
		_, err = w.dataBuf.Write([]byte{0x80})
		return
	}

	if nullable && value > 0 {
		value++
	}

	b := make([]byte, size+1)
	i := size
	for i >= 0 && value != 0 {
		b[i] = byte(value & 0x7F)
		value >>= 7
		i--
	}

	b[size] |= 0x80
	_, err = w.dataBuf.Write(b[i+1:])

	return
}

func (w *writer) WriteInt(nullable bool, value int64, size int) (err error) {
	if !nullable && value == 0 {
		_, err = w.dataBuf.Write([]byte{0x80})
		return
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
	_, err = w.dataBuf.Write(b[i:size+1])
	return
}

func (w *writer) WriteByteVector(nullable bool, value []byte) (err error) {
	err = w.WriteUint(nullable, uint64(len(value)), maxSize32)
	if err != nil {
		return
	}
	_, err = w.dataBuf.Write(value)
	return
}

func (w *writer) WriteString(nullable bool, value string) (err error) {
	if len(value) == 0 {
		if nullable {
			_, err = w.dataBuf.Write([]byte{0x00})
			return
		}
		_, err = w.dataBuf.Write([]byte{0x00})
		return
	}

	if len(value) == 1 && value[0] == 0x00 {
		if nullable {
			_, err = w.dataBuf.Write([]byte{0x00, 0x00, 0x80})
		} else {
			_, err = w.dataBuf.Write([]byte{0x00, 0x80})
		}
		return
	}

	w.strBuf.WriteString(value[:len(value)-1])
	_, err = w.dataBuf.Write(w.strBuf.Bytes())
	if err != nil {
		return
	}

	w.strBuf.Reset()
	_, err = w.dataBuf.Write([]byte{value[len(value)-1] | 0x80})
	return
}

func (w *writer) WriteNil() error {
	_, err := w.dataBuf.Write([]byte{0x80})
	return err
}
