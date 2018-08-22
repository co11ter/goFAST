package fast

type buffer struct {
	data []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{data: data}
}

func (b *buffer) decodeUint32() (result uint32) {
	i := 0
	result = uint32(b.data[i] & 0x7F)

	for (b.data[i] & 0x80) == 0 {
		result <<= 7;
		i++
		result |= uint32(b.data[i] & 0x7F);
	}

	b.data = b.data[i+1:]
	return result
}

func (b *buffer) decodeString() (result string) {
	/*i := 0
	result = string(b.data)
	if (b.data[i] & 0x80) == 0 {
		if b.data[i] == 0x80 {
			return result
		}
	}*/
	return ""
}
