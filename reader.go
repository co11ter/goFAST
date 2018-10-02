package fast

import (
	"bytes"
	"errors"
	"io"
)

type Reader struct {
	reader io.ByteReader
	buf bytes.Buffer
	last byte
}

func NewReader(reader io.ByteReader) *Reader {
	return &Reader{reader, bytes.Buffer{}, 0x00}
}

func (r *Reader) ReadPMap() (m *PMap, err error) {
	m = new(PMap)
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

func (r *Reader) ReadInt32(nullable bool) (*int32, error) {
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
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
		result |= int32(r.last & 0x7F);
	}

	if nullable {
		result -= decrement
	}

	return &result, nil
}

func (r *Reader) ReadInt64(nullable bool) (*int64, error) {
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
		result -= decrement
	}

	return &result, nil
}

func (r *Reader) ReadUint32(nullable bool) (*uint32, error) {
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

	if nullable && result > 0 {
		result--
	}

	return &result, nil
}

func (r *Reader) ReadUint64(nullable bool) (*uint64, error) {
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

	if nullable && result > 0 {
		result--
	}

	return &result, nil
}

// TODO
func (r *Reader) ReadUtfString(nullable bool) (*string, error) {
	return nil, nil
}

func (r *Reader) ReadAsciiString(nullable bool) (*string, error) {
	var err error
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return nil, err
	}

	var result string

	if (r.last & 0x7F) == 0 {
		if r.last == 0x80 {
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
			r.buf.WriteByte(r.last & 0x7F)
			break
		}
		r.buf.WriteByte(r.last)
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return nil, err
		}
	}

	result = r.buf.String()
	r.buf.Reset()
	return &result, nil
}
