// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

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
