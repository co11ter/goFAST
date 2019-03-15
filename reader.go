// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"io"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data
)

type reader struct {
	reader io.Reader
	strBuf bytes.Buffer
	bytes []byte
}

func newReader(r io.Reader) *reader {
	return &reader{r, bytes.Buffer{}, make([]byte, 1)}
}

func (r *reader) ReadPMap() (m *pMap, err error) {
	m = new(pMap)
	m.mask = 1;
	for i:=0; i < maxLoadBytes; i++ {
		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return
		}

		m.bitmap <<= 7
		m.bitmap |= uint(r.bytes[0]) & 0x7F
		m.mask <<= 7

		if 0x80 == (r.bytes[0] & 0x80) {
			return
		}
	}

	for {
		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return
		}

		if (r.bytes[0] & 0x80) == 0 {
			return
		}
	}

	// TODO what have to do here?
	return
}

func (r *reader) ReadInt(nullable bool) (*int64, error) {
	var err error
	_, err = r.reader.Read(r.bytes)
	if err != nil {
		return nil, err
	}

	var result int64
	var decrement int64 = 1

	if (r.bytes[0] & 0x40) > 0 {
		result = int64((-1 ^ int8(0x7F)) | int8((r.bytes[0] & 0x7F)))
		decrement = 0
	} else {
		result = int64(r.bytes[0] & 0x3F)
	}

	for (r.bytes[0] & 0x80) == 0 {
		result <<= 7;
		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return nil, err
		}
		result |= int64(r.bytes[0] & 0x7F);
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

func (r *reader) ReadUint(nullable bool) (*uint64, error) {
	var err error
	_, err = r.reader.Read(r.bytes)
	if err != nil {
		return nil, err
	}

	result := uint64(r.bytes[0] & 0x7F)

	for (r.bytes[0] & 0x80) == 0 {
		result <<= 7;
		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return nil, err
		}
		result |= uint64(r.bytes[0] & 0x7F);
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
	length, err := r.ReadUint(nullable)
	if err != nil {
		return nil, err
	}

	result := make([]byte, uint32(*length))
	_, err = io.ReadFull(r.reader, result)

	return &result, nil
}

// read ascii string
func (r *reader) ReadString(nullable bool) (*string, error) {
	var err error
	_, err = r.reader.Read(r.bytes)
	if err != nil {
		return nil, err
	}

	var result string

	if (r.bytes[0] & 0x7F) == 0 {
		if r.bytes[0] == 0x80 {
			if nullable {
				return nil, nil
			}
			return &result, nil
		}

		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return nil, err
		}

		if r.bytes[0] == 0x80 {
			return &result, nil
		} else if nullable && r.bytes[0] == 0x00 {
			_, err = r.reader.Read(r.bytes)
			if err != nil {
				return nil, err
			}

			if r.bytes[0] == 0x80 {
				return &result, nil
			}
		}
		err = ErrR9
		return nil, err
	}

	for {
		if (r.bytes[0] & 0x80) > 0 {
			r.strBuf.WriteByte(r.bytes[0] & 0x7F)
			break
		}
		r.strBuf.WriteByte(r.bytes[0])
		_, err = r.reader.Read(r.bytes)
		if err != nil {
			return nil, err
		}
	}

	result = r.strBuf.String()
	r.strBuf.Reset()
	return &result, nil
}
