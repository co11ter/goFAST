// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"io"
	"sync"
)

// A Decoder reads and decodes FAST-encoded message from an io.ByteReader.
// You may need buffered reader since decoder reads data byte by byte.
type Decoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id

	pMaps []*pMap
	index int // index for current presence map

	reader *reader
	msg *message

	logger *readerLog
	mu sync.Mutex
}

// NewDecoder returns a new decoder that reads from reader.
func NewDecoder(reader io.Reader, tmps ...*Template) *Decoder {
	decoder := &Decoder{
		repo: make(map[uint]*Template),
		storage: newStorage(),
		reader: newReader(reader),
		msg: newMsg(),
	}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

// SetLog sets writer for logging
func (d *Decoder) SetLog(writer io.Writer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if writer != nil {
		d.logger = wrapReaderLog(d.reader.reader, writer)
		d.reader = newReader(d.logger)
		return
	}

	if d.logger != nil {
		d.reader = newReader(d.logger.Reader)
		d.logger = nil
	}
}

// Decode reads the next FAST-encoded message from reader
// and stores it in the value pointed to by msg.
func (d *Decoder) Decode(msg interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.tid = 0
	d.pMaps = []*pMap{}

	if d.logger != nil {
		d.logger.prefix = "\n"
	}

	d.log("// ----- new message start ----- //")
	d.log("pmap decoding: ")
	d.visitPMap()
	d.log("  pmap = ", *d.pMaps[d.index], "\ntemplate decoding: ")

	d.tid = d.visitTemplateID()
	d.log("  template = ", d.tid)

	tpl, ok := d.repo[d.tid]
	if !ok {
		return ErrD9
	}

	d.msg.Reset(msg)
	d.msg.SetTID(d.tid)
	d.decodeSegment(tpl.Instructions)
	d.log("")

	return nil
}

func (d *Decoder) visitPMap() {
	m, err := d.reader.ReadPMap()
	if err != nil {
		panic(err)
	}

	if len(d.pMaps) > 0 {
		d.index++
	}

	d.pMaps = append(d.pMaps, m)
}

func (d *Decoder) restorePMap() {
	d.pMaps = d.pMaps[:d.index]
	d.index--
}

func (d *Decoder) visitTemplateID() uint {
	if d.pMaps[d.index].IsNextBitSet() {
		tmp, err := d.reader.ReadUint(false)
		if err != nil {
			panic(err)
		}
		return uint(*tmp)
	}
	return 0
}

func (d *Decoder) decodeGroup(instruction *Instruction) {
	d.log("group start: ")

	if instruction.isOptional() && !d.pMaps[d.index].IsNextBitSet() {
		d.log("group is empty")
		return
	}

	parent := &field{
		id: instruction.ID,
		name: instruction.Name,
		templateID: d.tid,
	}

	if instruction.pMapSize > 0 {
		d.log("pmap decoding: ")
		d.visitPMap()
		d.log("  pmap = ", *d.pMaps[d.index])
	}

	d.msg.Lock(parent)
	d.decodeSegment(instruction.Instructions)
	d.msg.Unlock()

	if instruction.pMapSize > 0 {
		d.restorePMap()
	}
}

func (d *Decoder) decodeSequence(instruction *Instruction) {
	d.log("sequence start: ")

	tmp, err := instruction.Instructions[0].extract(d.reader, d.storage, d.pMaps[d.index])
	if err != nil {
		panic(err)
	}
	length := int(tmp.(uint32))
	d.log("  length = ", length)

	parent := &field{
		id: instruction.ID,
		name: instruction.Name,
		templateID: d.tid,
		value: length,
	}

	d.msg.SetLen(parent)

	for i:=0; i<length; i++ {
		parent.num = i
		d.log("sequence elem[", i, "] start: ")

		if instruction.pMapSize > 0 {
			d.log("pmap decoding: ")
			d.visitPMap()
			d.log("  pmap = ", *d.pMaps[d.index])
		}

		d.msg.Lock(parent)
		d.decodeSegment(instruction.Instructions[1:])
		d.msg.Unlock()

		if instruction.pMapSize > 0 {
			d.restorePMap()
		}
	}
}

func (d *Decoder) decodeSegment(instructions []*Instruction) {
	if d.logger != nil {
		d.logger.Shift()
		defer d.logger.Unshift()
	}

	var err error
	for _, instruction := range instructions {
		switch instruction.Type {
		case TypeSequence:
			d.decodeSequence(instruction)
		case TypeGroup:
			d.decodeGroup(instruction)
		default:
			d.log("decoding: ", instruction.Name)
			d.log("  pmap -> ", d.pMaps[d.index])
			d.log("  reader -> ")
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: d.tid,
			}
			field.value, err = instruction.extract(d.reader, d.storage, d.pMaps[d.index])
			if err != nil {
				panic(err)
			}
			d.log("  ", field.name, " = ", field.value)
			d.msg.Set(field)
		}
	}
}

func (d *Decoder) log(param ...interface{}) {
	if d.logger == nil {
		return
	}

	d.logger.Log(param...)
}