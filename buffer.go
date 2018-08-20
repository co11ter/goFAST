package fast

import "reflect"

type buffer struct {
	data []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{data: data}
}

// value should be pointer
func (b *buffer) decode(value interface{}) bool {
	i := 0
	tmp := uint32(b.data[i] & 0x7F)

	for (b.data[i] & 0x80) == 0 {
		tmp <<= 7;
		i++
		tmp |= uint32(b.data[i] & 0x7F);
	}

	b.data = b.data[i+1:]
	reflect.ValueOf(value).Elem().Set(reflect.ValueOf(tmp))
	return true
}
