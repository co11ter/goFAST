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

var xmlEncodeTemplate = `
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

    <template name="Skip" id="999" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
        <string name="Str" id="15">
            <constant value="99"/>
        </string>
    </template>
</templates>`

type SenderSeq struct {
	SomeField uint64
}

type SenderMsg struct {
	TemplateID  uint
	Type   string
	Test   string
	Time   uint64
	Equal        int32
	Sequence    []SenderSeq

	seqLocked bool
	seqIndex int
}

func (br *SenderMsg) GetTemplateID() uint {
	return br.TemplateID
}

func (br *SenderMsg) GetLength(field *fast.Field) {
	if field.Name == "Sequence" {
		field.Value = len(br.Sequence)
	}
}

func (br *SenderMsg) Lock(field *fast.Field) bool {
	br.seqLocked = field.Name == "Sequence"
	if br.seqLocked {
		br.seqIndex = field.Value.(int)
	}
	return br.seqLocked
}

func (br *SenderMsg) Unlock() {
	br.seqLocked = false
	br.seqIndex = 0
}

func (br *SenderMsg) GetValue(field *fast.Field) {
	switch field.ID {
	case 131:
		field.Value = br.Test
	case 38:
		field.Value = br.Sequence[br.seqIndex].SomeField
	}
}

func Example_senderEncode() {
	var buf bytes.Buffer
	var msg = SenderMsg{
		TemplateID: 1,
		Test: "test",
		Sequence: []SenderSeq{
			{SomeField: 2},
		},
	}

	tpls, err := fast.ParseXMLTemplate(strings.NewReader(xmlEncodeTemplate))
	if err != nil {
		panic(err)
	}
	encoder := fast.NewEncoder(&buf, tpls...)

	if err := encoder.Encode(&msg); err != nil {
		panic(err)
	}
	fmt.Printf("%x", buf.Bytes())

	// Output: c081746573f480808182
}
