package fast

type PMap struct {
	bitmap uint
	mask uint
	value uint
}

func (p *PMap) IsNextBitSet() bool {
	p.mask >>= 1
	return (p.bitmap & p.mask) != 0;
}

func (p *PMap) SetNextBit(v bool) {
	p.mask >>= 1
	if v {
		p.value |= p.mask;
	}
}
