// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"fmt"
	"io"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data
)

type Decoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id

	prev *pMap
	current *pMap

	reader *reader

	logWriter io.Writer
}

func (d *Decoder) SetLog(writer io.Writer) {
	d.logWriter = writer
}

func (d *Decoder) Decode(msg interface{}) error {
	d.log("// ----- new message start ----- //\n")
	d.log("pmap parsing: ")
	d.visitPMap()
	d.log("\n  pmap = ", *d.current, "\n")

	d.log("template parsing: ")
	d.tid = d.visitTemplateID()
	d.log("\n  template = ", d.tid, "\n")

	tpl, ok := d.repo[d.tid]
	if !ok {
		return nil
	}

	m := newMsg(msg)
	m.assignTID(d.tid)
	d.decodeSegment(tpl.Instructions, m)
	d.tid = 0

	return nil
}

func NewDecoder(reader io.ByteReader, tmps ...*Template) *Decoder {
	decoder := &Decoder{
		repo: make(map[uint]*Template),
		storage: newStorage(),
		reader: newReader(reader),
	}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

func (d *Decoder) visitPMap() {
	var err error
	if d.current == nil {
		d.current, err = d.reader.ReadPMap()
		if err != nil {
			panic(err)
		}
	} else {
		tmp := *d.current
		d.current, err = d.reader.ReadPMap()
		if err != nil {
			panic(err)
		}
		d.prev = &tmp
	}
}

func (d *Decoder) visitTemplateID() uint {
	if d.current.IsNextBitSet() {
		tmp, err := d.reader.ReadUint32(false)
		if err != nil {
			panic(err)
		}
		return uint(*tmp)
	}
	return 0
}

func (d *Decoder) decodeSequence(instructions []*Instruction, msg *message) {
	d.log("sequence start: ")

	tmp, err := instructions[0].extract(d.reader, d.storage, d.current)
	if err != nil {
		panic(err)
	}
	length := int(tmp.(uint32))
	d.log("\n  length = ", length, "\n")

	for i:=0; i<length; i++ {
		d.log("sequence elem[", i, "] start: \n")
		d.log("pmap parsing: ")
		d.visitPMap()
		d.log("\n  pmap = ", *d.current, "\n")
		for _, instruction := range instructions[1:] {

			d.log("  parsing: ", instruction.Name, " ")
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: d.tid,
			}
			field.value, err = instruction.extract(d.reader, d.storage, d.current)
			if err != nil {
				panic(err)
			}
			d.log("\n    ", field.name, " = ", field.value, "\n")
			msg.AssignSlice(field, i)
		}
	}
}

func (d *Decoder) decodeSegment(instructions []*Instruction, msg *message) {
	var err error
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			d.decodeSequence(instruction.Instructions, msg)
		} else {
			d.log("parsing: ", instruction.Name, " ")
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: d.tid,
			}
			field.value, err = instruction.extract(d.reader, d.storage, d.current)
			if err != nil {
				panic(err)
			}
			d.log("\n  ", field.name, " = ", field.value, "\n")
			msg.Assign(field)
		}
	}
}

func (d *Decoder) log(param ...interface{}) {
	if d.logWriter == nil {
		return
	}

	d.logWriter.Write([]byte(fmt.Sprint(param...)))
}