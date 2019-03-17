// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

const defaultMask = 128

type pMap struct {
	bitmap uint
	mask uint
}

func (p *pMap) IsNextBitSet() bool {
	p.mask >>= 1
	return (p.bitmap & p.mask) != 0;
}

func (p *pMap) SetNextBit(v bool) {
	p.mask >>= 1
	if v {
		p.bitmap |= p.mask;
	}
}

func (p *pMap) String() (res string) {
	mask := p.mask
	for mask > 0 {
		mask >>= 1
		if (p.bitmap & mask) > 0 {
			res += "1"
		} else {
			res += "0"
		}
	}
	return
}

type pMapCollector struct {
	data []*pMap
	index int // index for current presence map
}

func newPMapCollector() *pMapCollector {
	return &pMapCollector{index: -1}
}

func (c *pMapCollector) reset() {
	c.data = c.data[:0]
	c.index = -1
}

func (c *pMapCollector) append(m *pMap) {
	c.data = append(c.data, m)
	if m != nil {
		c.index = len(c.data)-1
	}
}

func (c *pMapCollector) restore() {
	c.data = c.data[:len(c.data)-1]
	if c.index >= len(c.data) && c.index > 0 {
		c.index--
	}
	for c.data[c.index] == nil && c.index > 0 {
		c.index--
	}
}

func (c *pMapCollector) active() *pMap {
	return c.data[c.index]
}

func (c *pMapCollector) current() *pMap {
	return c.data[len(c.data)-1]
}