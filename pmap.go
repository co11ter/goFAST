package fast

type pmap struct {
	bitmap uint
	mask uint
}

func newPmap(buf *buffer) *pmap {
	m := new(pmap)
	m.mask = 1;
	for i:=0; i < maxLoadBytes; i++ {
		m.bitmap <<= 7
		m.bitmap |= uint(buf.data[i]) & '\x7F'
		m.mask <<= 7

		if (0x80 == (buf.data[i] & 0x80)) {
			buf.data = buf.data[i+1:]
			return m
		}
	}

	buf.cutEmpty()
	return nil
}

func (p *pmap) isNextBitSet() bool {
	p.mask >>= 1
	return (p.bitmap & p.mask) != 0;
}
