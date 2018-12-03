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

	prev *pMap
	current *pMap

	tmp *bytes.Buffer
	chunk *bytes.Buffer
	writer *writer

	target io.Writer

	logger *writerLog
	mu sync.Mutex
}

// NewEncoder returns a new encoder that writes FAST-encoded message to writer.
func NewEncoder(writer io.Writer, tmps ...*Template) *Encoder {
	encoder := &Encoder{
		repo: make(map[uint]*Template),
		storage: make(map[string]interface{}),
		chunk: &bytes.Buffer{},
		tmp: &bytes.Buffer{},
		writer: newWriter(&bytes.Buffer{}),
		target: writer,
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
		e.writer = newWriter(e.logger)
		return
	}

	if e.logger != nil {
		e.writer = newWriter(e.logger.Buffer)
		e.logger = nil
	}
}

// Encode encodes msg struct to writer
func (e *Encoder) Encode(msg interface{}) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.log("// ----- new message start ----- //\n")
	m := newMsg(msg)
	e.tid = m.GetTID()

	tpl, ok := e.repo[e.tid]
	if !ok {
		return ErrD9
	}

	e.acceptPMap()
	e.log("template = ", e.tid)
	e.log("\n  encoding -> ")
	e.acceptTemplateID(uint32(e.tid))

	e.encodeSegment(tpl.Instructions, m)

	return nil
}

func (e *Encoder) writePMap() {
	e.writer.WritePMap(e.current)
}

func (e *Encoder) writeTmp() {
	e.tmp.Write(e.writer.Bytes())
}

func (e *Encoder) writeChunk() {
	e.chunk.Write(e.writer.Bytes())
	if e.prev != nil {
		e.current = e.prev
		e.prev = nil
	}
}

func (e *Encoder) commit() error {
	e.current = nil
	e.prev = nil
	tmp := append(e.writer.Bytes(), e.tmp.Bytes()...)
	tmp = append(tmp, e.chunk.Bytes()...)
	_, err := e.target.Write(tmp)
	e.chunk.Reset()
	return err
}

func (e *Encoder) acceptPMap() {
	if e.current == nil {
		e.current = &pMap{mask: 128}
	} else {
		tmp := *e.current
		e.current = &pMap{mask: 128}
		e.prev = &tmp
	}
}

func (e *Encoder) acceptTemplateID(id uint32) {
	e.current.SetNextBit(true)
	e.writer.WriteUint32(false, id)
}

func (e *Encoder) encodeSegment(instructions []*Instruction, msg *message) {
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: e.tid,
			}
			msg.GetLen(field)
			e.encodeSequence(instruction, msg, field.value.(int))
		} else {
			field := &field{
				id: instruction.ID,
				name: instruction.Name,
				templateID: e.tid,
			}

			msg.Get(field)
			e.log("\n", instruction.Name, " = ", field.value, "\n")
			e.log("  encoding -> ")
			instruction.inject(e.writer, e.storage, e.current, field.value)
		}
	}
	e.log("\npmap = ", e.current, "\n")
	e.log("  encoding -> ")
	e.writePMap()
	e.log("\n")
	e.commit()
	e.log("\n")
}

func (e *Encoder) encodeSequence(instruction *Instruction, msg *message, length int) {
	e.log("\nsequence start: ")
	e.log("\n  length = ", length, "\n")
	e.log("    encoding -> ")
	instruction.Instructions[0].inject(e.writer, e.storage, e.current, uint32(length))

	parent := &field{
		id: instruction.ID,
		name: instruction.Name,
		templateID: e.tid,
	}

	e.writeTmp()
	for i:=0; i<length; i++ {
		e.log("\n  sequence elem[", i, "] start: ")
		e.acceptPMap()
		for _, internal := range instruction.Instructions[1:] {
			field := &field{
				id: internal.ID,
				name: internal.Name,
				templateID: e.tid,
				num: i,
				parent: parent,
			}

			msg.GetSlice(field)
			e.log("\n    ", internal.Name, " = ", field.value, "\n")
			e.log("      encoding -> ")
			internal.inject(e.writer, e.storage, e.current, field.value)
		}

		e.log("\n  pmap = ", e.current, "\n")
		e.log("    encoding -> ")
		e.writePMap()
		e.writeChunk()
	}
}

func (e *Encoder) log(param ...interface{}) {
	if e.logger == nil {
		return
	}

	e.logger.Log(param...)
}
