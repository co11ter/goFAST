// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast_test

type decimal struct {
	TemplateID           uint `fast:"*"`
	CopyDecimal          float64
	MandatoryDecimal     float64
	IndividualDecimal    float64
	IndividualDecimalOpt float64
}

var (
	decimalData1    = []byte{0xf8, 0x81, 0xfe, 0x4, 0x83, 0xff, 0xc, 0x8a, 0xfc, 0xa0, 0xff, 0x0, 0xef}
	decimalMessage1 = decimal{
		TemplateID:           1,
		CopyDecimal:          5.15,
		MandatoryDecimal:     154.6,
		IndividualDecimal:    0.0032,
		IndividualDecimalOpt: 11.1,
	}
)
