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
	reader *bytes.Buffer
)

func init() {
	ftpl, err := os.Open("testdata/test.xml")
	if err != nil {
		panic(err)
	}
	defer ftpl.Close()
	tpls := fast.ParseXMLTemplate(ftpl)

	reader = &bytes.Buffer{}
	decoder = fast.NewDecoder(reader, tpls...)
}

func decode(data []byte, msg interface{}, expect interface{}, t *testing.T) {
	reader.Write(data)
	err := decoder.Decode(msg)
	if err != nil {
		t.Fatal("can not decode", err)
	}

	if reader.Len() > 0 {
		t.Fatal("buffer is not empty")
	}

	if !reflect.DeepEqual(msg, expect) {
		t.Fatal("messages is not equal", msg, expect)
	}
}

func TestDecimalDecode(t *testing.T) {
	var msg decimalType
	decode(decimalData1, &msg, &decimalMessage1, t)
}

func TestSequenceDecode(t *testing.T) {
	var msg sequenceType
	decode(sequenceData1, &msg, &sequenceMessage1, t)
}

func TestByteVectorDecode(t *testing.T) {
	var msg byteVectorType
	decode(byteVectorData1, &msg, &byteVectorMessage1, t)
}

func TestStringDecode(t *testing.T) {
	var msg stringType
	decode(stringData1, &msg, &stringMessage1, t)
}