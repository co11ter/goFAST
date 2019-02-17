// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"io"
	"sync"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data
)

// A Decoder reads and decodes FAST-encoded message from an io.ByteReader.
// You may need buffered reader since decoder reads data byte by byte.
type Decoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id

	prev *pMap
	current *pMap

	reader *reader
	msg *message

	logger *readerLog
	prefix string // prefix for logger
	mu sync.Mutex
}

// NewDecoder returns a new decoder that reads from reader.
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
		d.reader = newReader(d.logger.ByteReader)
		d.logger = nil
	}
}

// Decode reads the next FAST-encoded message from reader
// and stores it in the value pointed to by msg.
func (d *Decoder) Decode(msg interface{}) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.prefix = "\n"
	d.tid = 0

	d.log("// ----- new message start ----- //")
	d.log("pmap decoding: ")
	d.visitPMap()
	d.log("  pmap = ", *d.current, "\ntemplate decoding: ")

	d.tid = d.visitTemplateID()
	d.log("  template = ", d.tid)

	tpl, ok := d.repo[d.tid]
	if !ok {
		return ErrD9
	}

	d.msg = newMsg(msg)
	d.msg.SetTID(d.tid)
	d.decodeSegment(tpl.Instructions, 0)

	return nil
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

func (d *Decoder) decodeSequence(instruction *Instruction) {
	d.log("sequence start: ")

	tmp, err := instruction.Instructions[0].extract(d.reader, d.storage, d.current)
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

	d.msg.Lock(parent)
	defer d.msg.Unlock()

	for i:=0; i<length; i++ {
		d.log("sequence elem[", i, "] start: ")
		d.log("pmap decoding: ")
		d.visitPMap()
		d.log("  pmap = ", *d.current)
		d.decodeSegment(instruction.Instructions[1:], i)
	}
}

func (d *Decoder) decodeSegment(instructions []*Instruction, index int) {
	d.shift()
	defer d.unshift()

	var err error
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			d.decodeSequence(instruction)
		} else {
			d.log("decoding: ", instruction.Name)
			d.log("  pmap -> ", d.current)
			d.log("  reader -> ")
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: d.tid,
				num: index,
			}
			field.value, err = instruction.extract(d.reader, d.storage, d.current)
			if err != nil {
				panic(err)
			}
			d.log("  ", field.name, " = ", field.value)
			d.msg.Set(field)
		}
	}
}

func (d *Decoder) shift() {
	if d.logger == nil {
		return
	}
	d.prefix += "  "
}

func (d *Decoder) unshift() {
	if d.logger == nil {
		return
	}
	d.prefix = d.prefix[:len(d.prefix)-2]
}

func (d *Decoder) log(param ...interface{}) {
	if d.logger == nil {
		return
	}

	d.logger.Log(append([]interface{}{d.prefix}, param...)...)
}