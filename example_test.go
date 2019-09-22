// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast_test

import (
	"bytes"
	"fmt"
	"github.com/co11ter/goFAST"
	"strings"
)

func ExampleDecoder_Decode_reflection() {
	var xmlData = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Done" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<string name="Type" id="15">
			<constant value="99"/>
		</string>
		<string name="Test" id="131" presence="optional"/>
		<uInt64 name="Time" id="20" presence="optional"/>
		<int32 name="Equal" id="271"/>
		<sequence name="Sequence">
			<length name="SeqLength" id="146"/>
			<uInt64 name="SomeField" id="38"/>
		</sequence>
	</template>
</templates>`

	type Seq struct {
		SomeField uint64
	}

	type ReflectMsg struct {
		TemplateID  uint    `fast:"*"`    // template id
		FieldByID   string  `fast:"15"`   // assign value by instruction id
		FieldByName string  `fast:"Test"` // assign value by instruction name
		Equal       int32   			  // name of field is default value for assign
		Nullable    *uint64 `fast:"20"`   // nullable - will skip, if field data is absent
		Skip        int     `fast:"-"`    // skip
		Sequence    []Seq
	}

	var msg ReflectMsg
	reader := bytes.NewReader(
		[]byte{0xc0, 0x81, 0x74, 0x65, 0x73, 0xf4, 0x80, 0x80, 0x81, 0x80, 0x82},
	)

	tpls, err := fast.ParseXMLTemplate(strings.NewReader(xmlData))
	if err != nil {
		panic(err)
	}
	decoder := fast.NewDecoder(
		reader,
		tpls...,
	)

	if err := decoder.Decode(&msg); err != nil {
		panic(err)
	}
	fmt.Print(msg)

	// Output: {1 99 test 0 <nil> 0 [{0}]}
}

func ExampleEncoder_Encode_reflection() {
	var xmlData = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Done" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<string name="Type" id="15">
			<constant value="99"/>
		</string>
		<string name="Test" id="131" presence="optional"/>
		<uInt64 name="Time" id="20" presence="optional"/>
		<int32 name="Equal" id="271"/>
		<sequence name="Sequence">
			<length name="SeqLength" id="146"/>
			<uInt64 name="SomeField" id="38"/>
		</sequence>
	</template>
</templates>`

	type Seq struct {
		SomeField uint64
	}

	type ReflectMsg struct {
		TemplateID  uint    `fast:"*"`    // template id
		FieldByID   string  `fast:"15"`   // assign value by instruction id
		FieldByName string  `fast:"Test"` // assign value by instruction name
		Equal       int32   			  // name of field is default value for assign
		Nullable    *uint64 `fast:"20"`   // nullable - will skip, if field data is absent
		Skip        int     `fast:"-"`    // skip
		Sequence    []Seq
	}

	var buf bytes.Buffer
	var msg = ReflectMsg{
		TemplateID: 1,
		FieldByName: "test",
		Sequence: []Seq{
			{SomeField: 2},
		},
	}

	tpls, err := fast.ParseXMLTemplate(strings.NewReader(xmlData))
	if err != nil {
		panic(err)
	}
	encoder := fast.NewEncoder(&buf, tpls...)

	if err := encoder.Encode(&msg); err != nil {
		panic(err)
	}
	fmt.Printf("%x", buf.Bytes())

	// Output: c081746573f4808182
}
