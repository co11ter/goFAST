// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/co11ter/goFAST"
)

var xmlDecodeTemplate = `
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

type ReceiverSeq struct {
	SomeField uint64
}

type ReceiverMsg struct {
	TemplateID  uint
	Type   string
	Test   string
	Time   uint64
	Equal        int32
	Sequence    []ReceiverSeq

	seqLocked bool
	seqIndex int
}

func (br *ReceiverMsg) SetTemplateID(tid uint) {
	br.TemplateID = tid
}

func (br *ReceiverMsg) SetLength(field *fast.Field) {
	if field.Name == "Sequence" && len(br.Sequence) < field.Value.(int) {
		br.Sequence = make([]ReceiverSeq, field.Value.(int))
	}
}

func (br *ReceiverMsg) Lock(field *fast.Field) bool {
	br.seqLocked = field.Name == "Sequence"
	if br.seqLocked {
		br.seqIndex = field.Value.(int)
	}
	return br.seqLocked
}

func (br *ReceiverMsg) Unlock() {
	br.seqLocked = false
	br.seqIndex = 0
}

func (br *ReceiverMsg) SetValue(field *fast.Field) {
	switch field.ID {
	case 15:
		br.Type = field.Value.(string)
	case 131:
		br.Test = field.Value.(string)
	case 20:
		br.Time = field.Value.(uint64)
	case 271:
		br.Equal = field.Value.(int32)
	case 38:
		br.Sequence[br.seqIndex].SomeField = field.Value.(uint64)
	}
}

func Example_receiverDecode() {
	var msg ReceiverMsg
	reader := bytes.NewReader(
		[]byte{0xc0, 0x81, 0x74, 0x65, 0x73, 0xf4, 0x80, 0x80, 0x81, 0x80, 0x82},
	)

	tpls, err := fast.ParseXMLTemplate(strings.NewReader(xmlDecodeTemplate))
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

	// Output: {1 99 test 0 0 [{0}] false 0}
}