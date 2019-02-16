// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast_test

import (
	"bytes"
	"github.com/co11ter/goFAST"
	"os"
	"reflect"
	"testing"
)

var (
	decoder *fast.Decoder
	buffer *bytes.Buffer
)

func init() {
	ftpl, err := os.Open("testdata/test.xml")
	if err != nil {
		panic(err)
	}
	defer ftpl.Close()
	tpls := fast.ParseXMLTemplate(ftpl)

	buffer = &bytes.Buffer{}
	decoder = fast.NewDecoder(buffer, tpls...)
}

func decode(data []byte, msg interface{}, expect interface{}, t *testing.T) {
	buffer.Write(data)
	err := decoder.Decode(msg)
	if err != nil {
		t.Fatal("can not decode", err)
	}

	if buffer.Len() > 0 {
		t.Fatal("buffer is not empty")
	}

	if !reflect.DeepEqual(msg, expect) {
		t.Fatal("messages is not equal", msg, expect)
	}
}

func TestDecimal(t *testing.T) {
	data := []byte{0xf8, 0x81, 0xfe, 0x4, 0x83, 0xff, 0xc, 0x8a, 0xfc, 0xa0, 0xff, 0x0, 0xef}
	type Msg struct {
		CopyDecimal          float64
		MandatoryDecimal     float64
		IndividualDecimal    float64
		IndividualDecimalOpt float64
	}

	var msg Msg
	var expect = Msg{
		CopyDecimal: 5.15,
		MandatoryDecimal: 154.6,
		IndividualDecimal: 0.0032,
		IndividualDecimalOpt: 11.1,
	}

	decode(data, &msg, &expect, t)
}
