// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"io"
	"sync"
)

// A Decoder reads and decodes FAST-encoded message from an io.Reader.
// You may need buffered reader since decoder reads data byte by byte.
type Decoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id
	pmc *pMapCollector

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
		pmc: newPMapCollector(),
	}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

// Reset resets dictionary
func (d *Decoder) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.storage = newStorage()
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
	d.pmc.reset()

	if d.logger != nil {
		d.logger.prefix = "\n"
		d.logger.Log("// ----- new message start ----- //")
		d.logger.Log("pmap decoding: ")
	}

	err := d.visitPMap()
	if err != nil {
		return err
	}

	if d.logger != nil {
		d.logger.Log("  pmap = ", *d.pmc.active(), "\ntemplate decoding: ")
	}

	d.tid, err = d.visitTemplateID()
	if err != nil {
		return err
	}

	if d.logger != nil {
		d.logger.Log("  template = ", d.tid)
	}

	tpl, ok := d.repo[d.tid]
	if !ok {
		return ErrD9
	}

	d.msg.Reset(msg)
	d.msg.SetTID(d.tid)
	return d.decodeSegment(tpl.Instructions)
}

func (d *Decoder) visitPMap() error {
	m, err := d.reader.ReadPMap()
	if err != nil {
		return err
	}

	d.pmc.append(m)
	return nil
}

func (d *Decoder) visitTemplateID() (uint, error) {
	if d.pmc.active().IsNextBitSet() {
		tmp, err := d.reader.ReadUint(false)
		if err != nil {
			return 0, err
		}
		return uint(*tmp), nil
	}
	return 0, nil
}

func (d *Decoder) decodeGroup(instruction *Instruction) error {
	if d.logger != nil {
		d.logger.Log("group start: ")
	}

	if instruction.isOptional() && !d.pmc.active().IsNextBitSet() {
		if d.logger != nil {
			d.logger.Log("group is empty")
		}
		return nil
	}

	parent := &field{
		id: instruction.ID,
		name: instruction.Name,
		templateID: d.tid,
	}

	if instruction.pMapSize > 0 {
		if d.logger != nil {
			d.logger.Log("pmap decoding: ")
		}
		err := d.visitPMap()
		if err != nil {
			return err
		}
		if d.logger != nil {
			d.logger.Log("  pmap = ", *d.pmc.active())
		}
	}

	locked := d.msg.Lock(parent)
	err := d.decodeSegment(instruction.Instructions)
	if err != nil {
		return err
	}

	if locked {
		d.msg.Unlock()
	}

	if instruction.pMapSize > 0 {
		d.pmc.restore()
	}

	return nil
}

func (d *Decoder) decodeSequence(instruction *Instruction) error {
	if d.logger != nil {
		d.logger.Log("sequence start: ")
	}

	tmp, err := instruction.Instructions[0].extract(d.reader, d.storage, d.pmc.active())
	if err != nil {
		return err
	}

	if tmp == nil {
		return nil
	}

	length := int(tmp.(uint32))
	if d.logger != nil {
		d.logger.Log("  length = ", length)
	}

	parent := &field{
		id: instruction.ID,
		name: instruction.Name,
		templateID: d.tid,
		value: length,
	}

	d.msg.SetLen(parent)

	for i:=0; i<length; i++ {
		parent.num = i
		if d.logger != nil {
			d.logger.Log("sequence elem[", i, "] start: ")
		}

		if instruction.pMapSize > 0 {
			if d.logger != nil {
				d.logger.Log("pmap decoding: ")
			}
			err = d.visitPMap()
			if err != nil {
				return err
			}
			if d.logger != nil {
				d.logger.Log("  pmap = ", *d.pmc.active())
			}
		}

		locked := d.msg.Lock(parent)
		err = d.decodeSegment(instruction.Instructions[1:])
		if err != nil {
			return err
		}

		if locked {
			d.msg.Unlock()
		}

		if instruction.pMapSize > 0 {
			d.pmc.restore()
		}
	}

	return nil
}

func (d *Decoder) decodeSegment(instructions []*Instruction) error {
	if d.logger != nil {
		d.logger.Shift()
		defer d.logger.Unshift()
	}

	var err error
	for _, instruction := range instructions {
		switch instruction.Type {
		case TypeSequence:
			err = d.decodeSequence(instruction)
		case TypeGroup:
			err = d.decodeGroup(instruction)
		default:
			if d.logger != nil {
				d.logger.Log("decoding: ", instruction.Name)
				d.logger.Log("  pmap -> ", d.pmc.active())
				d.logger.Log("  reader -> ")
			}

			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: d.tid,
			}
			field.value, err = instruction.extract(d.reader, d.storage, d.pmc.active())
			if err != nil {
				return err
			}

			if d.logger != nil {
				d.logger.Log("  ", field.name, " = ", field.value)
			}

			d.msg.Set(field)
		}

		if err != nil {
			return err
		}
	}

	return err
}
