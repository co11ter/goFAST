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

// reader reads type data from io.Reader. No thread safe!
type reader struct {
	reader io.Reader
	strBuf bytes.Buffer
	bytes  []byte

	tmpErr  error
	tmpUint uint64
	tmpInt  int64
	tmpDcrm int64
	tmpLen  *uint64
	tmpByte []byte
	tmpStr  string
}

func newReader(r io.Reader) *reader {
	return &reader{reader: r, strBuf: bytes.Buffer{}, bytes: make([]byte, 1)}
}

func (r *reader) ReadPMap() (m *pMap, err error) {
	m = new(pMap)
	m.mask = 1
	for i := 0; i < maxLoadBytes; i++ {
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
	_, r.tmpErr = r.reader.Read(r.bytes)
	if r.tmpErr != nil {
		return nil, r.tmpErr
	}

	r.tmpDcrm = 1

	if (r.bytes[0] & 0x40) > 0 {
		r.tmpInt = int64((-1 ^ int8(0x7F)) | int8((r.bytes[0] & 0x7F)))
		r.tmpDcrm = 0
	} else {
		r.tmpInt = int64(r.bytes[0] & 0x3F)
	}

	for (r.bytes[0] & 0x80) == 0 {
		r.tmpInt <<= 7
		_, r.tmpErr = r.reader.Read(r.bytes)
		if r.tmpErr != nil {
			return nil, r.tmpErr
		}
		r.tmpInt |= int64(r.bytes[0] & 0x7F)
	}

	if nullable {
		if r.tmpInt == 0 {
			return nil, r.tmpErr
		} else {
			r.tmpInt -= r.tmpDcrm
		}
	}

	return &r.tmpInt, nil
}

func (r *reader) ReadUint(nullable bool) (*uint64, error) {
	_, r.tmpErr = r.reader.Read(r.bytes)
	if r.tmpErr != nil {
		return nil, r.tmpErr
	}

	r.tmpUint = uint64(r.bytes[0] & 0x7F)

	for (r.bytes[0] & 0x80) == 0 {
		r.tmpUint <<= 7
		_, r.tmpErr = r.reader.Read(r.bytes)
		if r.tmpErr != nil {
			return nil, r.tmpErr
		}
		r.tmpUint |= uint64(r.bytes[0] & 0x7F)
	}

	if nullable {
		if r.tmpUint == 0 {
			return nil, r.tmpErr
		} else {
			r.tmpUint--
		}
	}

	return &r.tmpUint, nil
}

func (r *reader) ReadByteVector(nullable bool) (*[]byte, error) {
	r.tmpLen, r.tmpErr = r.ReadUint(nullable)
	if r.tmpErr != nil {
		return nil, r.tmpErr
	}

	if uint32(*r.tmpLen) > uint32(len(r.tmpByte)) {
		r.tmpByte = make([]byte, uint32(*r.tmpLen))
	} else {
		r.tmpByte = r.tmpByte[:uint32(*r.tmpLen)]
	}

	_, r.tmpErr = io.ReadFull(r.reader, r.tmpByte)
	return &r.tmpByte, nil
}

// read ascii string
func (r *reader) ReadString(nullable bool) (*string, error) {
	_, r.tmpErr = r.reader.Read(r.bytes)
	if r.tmpErr != nil {
		return nil, r.tmpErr
	}

	if (r.bytes[0] & 0x7F) == 0 {
		if r.bytes[0] == 0x80 {
			if nullable {
				return nil, nil
			}
			return &r.tmpStr, nil
		}

		_, r.tmpErr = r.reader.Read(r.bytes)
		if r.tmpErr != nil {
			return nil, r.tmpErr
		}

		if r.bytes[0] == 0x80 {
			return &r.tmpStr, nil
		} else if nullable && r.bytes[0] == 0x00 {
			_, r.tmpErr = r.reader.Read(r.bytes)
			if r.tmpErr != nil {
				return nil, r.tmpErr
			}

			if r.bytes[0] == 0x80 {
				return &r.tmpStr, nil
			}
		}
		r.tmpErr = ErrR9
		return nil, r.tmpErr
	}

	for {
		if (r.bytes[0] & 0x80) > 0 {
			r.strBuf.WriteByte(r.bytes[0] & 0x7F)
			break
		}
		r.strBuf.WriteByte(r.bytes[0])
		_, r.tmpErr = r.reader.Read(r.bytes)
		if r.tmpErr != nil {
			return nil, r.tmpErr
		}
	}

	r.tmpStr = r.strBuf.String()
	r.strBuf.Reset()
	return &r.tmpStr, nil
}
