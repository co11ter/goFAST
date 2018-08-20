package fast

type buffer struct {
	data []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{data: data}
}

func (b *buffer) decodeUint32() uint32 {
	i := 0
	tmp := uint32(b.data[i] & 0x7F)

	for (b.data[i] & 0x80) == 0 {
		tmp <<= 7;
		i++
		tmp |= uint32(b.data[i] & 0x7F);
	}

	b.data = b.data[i+1:]
	return tmp
}
