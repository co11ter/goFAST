package fast

import (
	"bytes"
	"fmt"
	"io"
)

type Encoder struct {
	repo map[uint]*Template
	storage storage

	tid uint // template id

	prev *PMap
	current *PMap

	tmp *bytes.Buffer
	chunk *bytes.Buffer
	writer *Writer

	target io.Writer

	logWriter io.Writer
}

func NewEncoder(writer io.Writer, tmps ...*Template) *Encoder {
	encoder := &Encoder{
		repo: make(map[uint]*Template),
		storage: make(map[string]interface{}),
		chunk: &bytes.Buffer{},
		tmp: &bytes.Buffer{},
		target: writer,
	}
	for _, t := range tmps {
		encoder.repo[t.ID] = t
	}
	encoder.setBuffer(&bytes.Buffer{})
	return encoder
}

func (e *Encoder) setBuffer(buf buffer) {
	e.writer = NewWriter(buf)
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
		e.current = &PMap{mask: 128}
	} else {
		tmp := *e.current
		e.current = &PMap{mask: 128}
		e.prev = &tmp
	}
}

func (e *Encoder) acceptTemplateID(id uint32) {
	e.current.SetNextBit(true)
	e.writer.WriteUint32(false, id)
}

func (e *Encoder) SetLog(writer io.Writer) {
	e.logWriter = writer
	e.setBuffer(newLogger(writer))
}

func (e *Encoder) Encode(msg interface{}) error {
	e.log("// ----- new message start ----- //\n")
	m := newMsg(msg)
	e.tid = m.LookUpTID()

	tpl, ok := e.repo[e.tid]
	if !ok {
		return nil
	}

	e.acceptPMap()
	e.log("template = ", e.tid)
	e.log("\n  encoding -> ")
	e.acceptTemplateID(uint32(e.tid))

	e.encodeSegment(tpl.Instructions, m)

	return nil
}

func (e *Encoder) encodeSegment(instructions []*Instruction, msg *message) {
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			field := &Field{
				ID: instruction.ID,
				Name: instruction.Name,
				TemplateID: e.tid,
			}
			msg.LookUpLen(field)
			e.encodeSequence(instruction.Instructions, msg, field.Value.(int))
		} else {
			field := &Field{
				ID: instruction.ID,
				Name: instruction.Name,
				TemplateID: e.tid,
			}

			msg.LookUp(field)
			e.log("\n", instruction.Name, " = ", field.Value, "\n")
			e.log("  encoding -> ")
			instruction.inject(e.writer, e.storage, e.current, field.Value)
		}
	}
	e.log("\npmap = ", e.current, "\n")
	e.log("  encoding -> ")
	e.writePMap()
	e.log("\n")
	e.commit()
	e.log("\n")
}

func (e *Encoder) encodeSequence(instructions []*Instruction, msg *message, length int) {
	e.log("\nsequence start: ")
	e.log("\n  length = ", length, "\n")
	e.log("    encoding -> ")
	instructions[0].inject(e.writer, e.storage, e.current, uint32(length))

	e.writeTmp()
	for i:=0; i<length; i++ {
		e.log("\n  sequence elem[", i, "] start: ")
		e.acceptPMap()
		for _, instruction := range instructions[1:] {
			field := &Field{
				ID: instruction.ID,
				Name: instruction.Name,
				TemplateID: e.tid,
			}

			msg.LookUpSlice(field, i)
			e.log("\n    ", instruction.Name, " = ", field.Value, "\n")
			e.log("      encoding -> ")
			instruction.inject(e.writer, e.storage, e.current, field.Value)
		}

		e.log("\n  pmap = ", e.current, "\n")
		e.log("    encoding -> ")
		e.writePMap()
		e.writeChunk()
	}
}

func (e *Encoder) log(param ...interface{}) {
	if e.logWriter == nil {
		return
	}

	e.logWriter.Write([]byte(fmt.Sprint(param...)))
}
