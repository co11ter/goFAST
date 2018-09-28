package fast

import (
	"bytes"
	"errors"
	"io"
)

// TODO ok - still don't use
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

func (r *Reader) ReadInt32(nullable bool) (result int32, ok bool, err error) {
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return
	}

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
			return
		}
		result |= int32(r.last & 0x7F);
	}

	if nullable {
		result -= decrement
	}

	return
}

func (r *Reader) ReadInt64(nullable bool) (result int64, ok bool, err error) {
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return
	}

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
			return
		}
		result |= int64(r.last & 0x7F);
	}

	if nullable {
		result -= decrement
	}

	return
}

func (r *Reader) ReadUint32(nullable bool) (result uint32, ok bool, err error) {
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return
	}

	result = uint32(r.last & 0x7F)

	for (r.last & 0x80) == 0 {
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}
		result |= uint32(r.last & 0x7F);
	}

	if nullable && result > 0 {
		result--
	}

	return
}

func (r *Reader) ReadUint64(nullable bool) (result uint64, ok bool, err error) {
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return
	}

	result = uint64(r.last & 0x7F)

	for (r.last & 0x80) == 0 {
		result <<= 7;
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}
		result |= uint64(r.last & 0x7F);
	}

	if nullable && result > 0 {
		result--
	}

	return
}

func (r *Reader) ReadAsciiString(nullable bool) (result string, ok bool, err error) {
	r.last, err = r.reader.ReadByte()
	if err != nil {
		return
	}

	if (r.last & 0x7F) == 0 {
		if r.last == 0x80 {
			return
		}

		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}

		if r.last == 0x80 {
			return
		} else if nullable && r.last == 0x00 {
			r.last, err = r.reader.ReadByte()
			if err != nil {
				return
			}

			if r.last == 0x80 {
				return
			}
		}
		err = errors.New("d9")
		return
	}

	for {
		if (r.last & 0x80) > 0 {
			r.buf.WriteByte(r.last & 0x7F)
			break
		}
		r.buf.WriteByte(r.last)
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}
	}

	result = r.buf.String()
	r.buf.Reset()
	return
}
