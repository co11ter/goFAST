package fast

type PMap struct {
	bitmap uint
	mask uint
}

func (p *PMap) IsNextBitSet() bool {
	p.mask >>= 1
	return (p.bitmap & p.mask) != 0;
}
