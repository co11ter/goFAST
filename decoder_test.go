// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast_test

import (
	"bytes"
	"github.com/co11ter/goFAST"
	"io"
	"io/ioutil"
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
	tpls, _ := fast.ParseXMLTemplate(ftpl)

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
		t.Fatal("messages is not equal, got: ", msg, ", expect: ", expect)
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

func TestIntegerDecode(t *testing.T) {
	var msg integerType
	decode(integerData1, &msg, &integerMessage1, t)
}

func TestGroupDecode(t *testing.T) {
	var msg groupType
	decode(groupData1, &msg, &groupMessage1, t)
}

// write profile command: go test -bench=. -cpuprofile=cpu.out -memprofile=mem.out
// convert to cpuprof.pdf command: go tool pprof -pdf -output=cpuprof.pdf goFAST.test cpu.out
// convert to memprof.pdf command: go tool pprof -pdf -output=memprof.pdf goFAST.test mem.out
func BenchmarkDecoder_Decode(b *testing.B) {
	file, err := os.Open("testdata/data.dat")
	if err != nil {
		b.Fatal(err)
	}

	data, _ := ioutil.ReadAll(file)
	file.Close()
	reader.Write(data)

	var msg benchmarkMessage
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader.Next(4) // skip sequence data
		err = decoder.Decode(&msg)
		if err == io.EOF {
			b.StopTimer()
			reader.Write(data)
			b.StartTimer()
			continue
		}
		if err != nil {
			b.Fatal(err)
		}
	}
}