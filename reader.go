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
		r.last, err = r.reader.ReadByte()
		if err != nil {
			return
		}
		if (r.last & 0x80) == 0 {
			break
		}
		r.buf.WriteByte(r.last)
	}
	result = r.buf.String()
	r.buf.Reset()
	return
}

//func (r *Reader) Int() (res []int8) {
//	for _, x := range b.data {
//		res = append(res, int8(x))
//	}
//	return
//}
//
//func (r *Reader) Hex() string {
//	var result string
//	str := hex.EncodeToString(b.data)
//	for i:=0; i<len(str); i++ {
//		if i%2==0 {
//			result += " "
//		}
//		result += string(str[i])
//	}
//	return result
//}
