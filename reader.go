// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"errors"
	"io"
)

type reader struct {
	reader io.ByteReader
	strBuf bytes.Buffer
	last byte
}

func newReader(r io.ByteReader) *reader {
	return &reader{r, bytes.Buffer{}, 0x00}
}

func (r *reader) ReadPMap() (m *pMap, err error) {
	m = new(pMap)
	m.mask = 1;
	for i:=0; i < maxLoadBytes; i++ {
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}

		m.bitmap <<= 7
		m.bitmap |= uint(r.last) & 0x7F
		m.mask <<= 7

		if 0x80 == (r.last & 0x80) {
			return
		}
	}

	for {
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}

		if (r.last & 0x80) == 0 {
			return
		}
	}

	// TODO what have to do here?
	return
}

func (r *reader) ReadInt32(nullable bool) (*int32, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	var result int32
	var decrement int32 = 1

	if (r.last & 0x40) > 0 {
		result = int32((-1 ^ int8(0x7F)) | int8((r.last & 0x7F)))
		decrement = 0
	} else {
		result = int32(r.last & 0x3F)
	}

	for (r.last & 0x80) == 0 {
		result <<= 7
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		result |= int32(r.last & 0x7F);
	}

	if nullable {
		if result == 0 {
			return nil, err
		} else {
			result -= decrement
		}
	}

	return &result, nil
}

func (r *reader) ReadInt64(nullable bool) (*int64, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	var result int64
	var decrement int64 = 1

	if (r.last & 0x40) > 0 {
		result = int64((-1 ^ int8(0x7F)) | int8((r.last & 0x7F)))
		decrement = 0
	} else {
		result = int64(r.last & 0x3F)
	}

	for (r.last & 0x80) == 0 {
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		result |= int64(r.last & 0x7F);
	}

	if nullable {
		if result == 0 {
			return nil, err
		} else {
			result -= decrement
		}
	}

	return &result, nil
}

func (r *reader) ReadUint32(nullable bool) (*uint32, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	result := uint32(r.last & 0x7F)

	for (r.last & 0x80) == 0 {
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		result |= uint32(r.last & 0x7F);
	}

	if nullable {
		if result == 0 {
			return nil, err
		} else {
			result--
		}
	}

	return &result, nil
}

func (r *reader) ReadUint64(nullable bool) (*uint64, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	result := uint64(r.last & 0x7F)

	for (r.last & 0x80) == 0 {
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		result |= uint64(r.last & 0x7F);
	}

	if nullable {
		if result == 0 {
			return nil, err
		} else {
			result--
		}
	}

	return &result, nil
}

func (r *reader) ReadByteVector(nullable bool) (*[]byte, error) {
	length, err := r.ReadUint32(nullable)
	if err != nil {
		return nil, err
	}

	var result []byte
	for *length > 0 {
		r.last, err = r.reader.ReadByte()
		result = append(result, r.last)
		*length--
	}

	return &result, nil
}

func (r *reader) ReadASCIIString(nullable bool) (*string, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	var result string

	if (r.last & 0x7F) == 0 {
		if r.last == 0x80 {
			if nullable {
				return nil, nil
			}
			return &result, nil
		}

		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}

		if r.last == 0x80 {
			return &result, nil
		} else if nullable && r.last == 0x00 {
			r.last, err = r.reader.ReadByte()
			if err != nil {
				return nil, err
			}

			if r.last == 0x80 {
				return &result, nil
			}
		}
		err = errors.New("d9")
		return nil, err
	}

	for {
		if (r.last & 0x80) > 0 {
			r.strBuf.WriteByte(r.last & 0x7F)
			break
		}
		r.strBuf.WriteByte(r.last)
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	result = r.strBuf.String()
	r.strBuf.Reset()
	return &result, nil
}
