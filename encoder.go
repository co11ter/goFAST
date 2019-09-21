// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"io"
	"sync"
)

// A Encoder encodes and writes data to io.Writer.
type Encoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id
	pmc *pMapCollector

	writers []*writer
	writerIndex int // index for current writer

	msg Sender

	target io.Writer

	logger *writerLog
	mu sync.Mutex
}

// Reset resets dictionary
func (e *Encoder) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.storage = newStorage()
}

// NewEncoder returns a new encoder that writes FAST-encoded message to writer.
func NewEncoder(writer io.Writer, tmps ...*Template) *Encoder {
	encoder := &Encoder{
		repo: make(map[uint]*Template),
		storage: make(map[string]interface{}),
		target: writer,
		pmc: newPMapCollector(),
	}
	for _, t := range tmps {
		encoder.repo[t.ID] = t
	}
	return encoder
}

// SetLog sets writer for logging
func (e *Encoder) SetLog(writer io.Writer) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if writer != nil {
		e.logger = wrapWriterLog(writer)
		return
	}

	if e.logger != nil {
		e.logger = nil
	}
}

// Encode encodes msg struct to writer. If an encountered value implements the Sender interface
// and is not a nil pointer, Encode calls method of Sender to produce encoded message.
func (e *Encoder) Encode(msg interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.pmc.reset()
	e.writers = []*writer{}
	e.writerIndex = 0

	if e.logger != nil {
		e.logger.prefix = "\n"
	}

	e.log("// ----- new message start ----- //")

	var ok bool
	if e.msg, ok = msg.(Sender); !ok {
		e.msg = makeMsg(msg)
	}
	e.tid = e.msg.GetTemplateID()

	tpl, ok := e.repo[e.tid]
	if !ok {
		return ErrD9
	}

	e.pmc.append(&pMap{mask: defaultMask})
	e.addWriter()
	e.log("template = ", e.tid)
	e.log("  encoding -> ")
	e.acceptTemplateID(uint32(e.tid))

	err := e.encodeSegment(tpl.Instructions)
	if err != nil {
		return err
	}
	return e.commit()
}

func (e *Encoder) addWriter() {
	if e.logger != nil {
		e.writers = append(e.writers, newWriter(e.logger, wrapWriterLog(e.logger.log)))
	} else {
		e.writers = append(e.writers, newWriter(&bytes.Buffer{}, &bytes.Buffer{}))
	}
	e.writerIndex = len(e.writers) -1
}

func (e *Encoder) commit() error {
	tmp := &bytes.Buffer{}
	for _, writer := range e.writers {
		writer.WriteTo(tmp)
	}
	_, err := tmp.WriteTo(e.target)
	return err
}

func (e *Encoder) acceptTemplateID(id uint32) {
	e.pmc.active().SetNextBit(true)
	_ = e.writers[e.writerIndex].WriteUint(false, uint64(id), maxSize32)
}

func (e *Encoder) encodeSegment(instructions []*Instruction) error {
	if e.logger != nil {
		e.logger.Shift()
		defer e.logger.Unshift()
	}

	var err error
	for _, instruction := range instructions {
		switch instruction.Type {
		case TypeSequence:
			err = e.encodeSequence(instruction)
		case TypeGroup:
			err = e.encodeGroup(instruction)
		default:
			field := acquireField()
			field.ID = instruction.ID
			field.Name = instruction.Name

			e.msg.GetValue(field)
			e.log(instruction.Name, " = ", field.Value)
			e.log("  encoding -> ")
			err = instruction.inject(
				e.writers[e.writerIndex],
				e.storage,
				e.pmc.active(),
				field.Value,
			)
			releaseField(field)
		}

		if err != nil {
			return err
		}
	}
	e.log("pmap = ", e.pmc.current())
	e.log("  encoding -> ")

	if m := e.pmc.current(); m != nil {
		_ = e.writers[e.writerIndex].WritePMap(m)
	}

	return nil
}

func (e *Encoder) encodeGroup(instruction *Instruction) error {
	e.log("group start: ")
	parent := acquireField()
	parent.ID = instruction.ID
	parent.Name = instruction.Name

	if instruction.isOptional() {
		e.pmc.active().SetNextBit(true)
	}

	current := e.writerIndex // remember current writer index

	var pmap *pMap
	if instruction.pMapSize > 0 {
		pmap = &pMap{mask: defaultMask}
	}

	e.pmc.append(pmap)
	e.addWriter()

	e.msg.Lock(parent)
	err := e.encodeSegment(instruction.Instructions)
	if err != nil {
		return err
	}
	e.msg.Unlock()
	releaseField(parent)

	e.pmc.restore()
	e.writerIndex = current // restore index
	return nil
}

func (e *Encoder) encodeSequence(instruction *Instruction) error {
	parent := acquireField()
	parent.ID = instruction.ID
	parent.Name = instruction.Name

	e.msg.GetLength(parent)
	length := parent.Value.(int)

	e.log("sequence start: ")
	e.log("  length = ", length)
	e.log("    encoding -> ")
	err := instruction.Instructions[0].inject(
		e.writers[e.writerIndex],
		e.storage,
		e.pmc.active(),
		uint32(length),
	)
	if err != nil {
		return err
	}

	current := e.writerIndex // remember current writer index
	for i:=0; i<length; i++ {
		parent.Value = i
		e.log("sequence elem[", i, "] start: ")

		var pmap *pMap
		if instruction.pMapSize > 0 {
			pmap = &pMap{mask: defaultMask}
		}

		e.pmc.append(pmap)
		e.addWriter()

		e.msg.Lock(parent)
		err = e.encodeSegment(instruction.Instructions[1:])
		if err != nil {
			return err
		}
		e.msg.Unlock()
		e.pmc.restore()
	}
	releaseField(parent)
	e.writerIndex = current // restore index
	return nil
}

func (e *Encoder) log(param ...interface{}) {
	if e.logger == nil {
		return
	}

	e.logger.Log(param...)
}
