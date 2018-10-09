package fast

import (
	"bytes"
	"fmt"
	"io"
)

type Encoder struct {
	repo map[uint]*Template
	acceptor *Acceptor
	tid uint

	logWriter io.Writer
}

func NewEncoder(writer io.Writer, tmps ...*Template) *Encoder {
	encoder := &Encoder{
		repo: make(map[uint]*Template),
		acceptor: newAcceptor(writer),
	}
	for _, t := range tmps {
		encoder.repo[t.ID] = t
	}
	encoder.acceptor.setBuffer(&bytes.Buffer{})
	return encoder
}

func (e *Encoder) SetLog(writer io.Writer) {
	e.logWriter = writer
	e.acceptor.setBuffer(newLogger(writer))
}

func (e *Encoder) Encode(msg interface{}) error {
	e.log("// ----- new message start ----- //\n")
	m := newMsg(msg)
	e.tid = m.LookUpTID()

	tpl, ok := e.repo[e.tid]
	if !ok {
		return nil
	}

	e.acceptor.acceptPMap()
	e.log("template = ", e.tid)
	e.log("\n  encoding -> ")
	e.acceptor.acceptTemplateID(uint32(e.tid))

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
			e.acceptor.accept(instruction, field.Value)
		}
	}
	e.log("\n")
	e.acceptor.commit()
}

func (e *Encoder) encodeSequence(instructions []*Instruction, msg *message, length int) {
	e.log("\nsequence start: ")
	e.log("\n  length = ", length, "\n")
	e.log("    encoding -> ")
	e.acceptor.accept(instructions[0], uint32(length))
	for i:=0; i<length; i++ {
		e.log("\n  sequence elem[", i, "] start: ")
		e.acceptor.acceptPMap()
		for _, instruction := range instructions[1:] {
			field := &Field{
				ID: instruction.ID,
				Name: instruction.Name,
				TemplateID: e.tid,
			}

			msg.LookUpSlice(field, i)
			e.log("\n    ", instruction.Name, " = ", field.Value, "\n")
			e.log("      encoding -> ")
			e.acceptor.accept(instruction, field.Value)
		}
	}
}

func (e *Encoder) log(param ...interface{}) {
	if e.logWriter == nil {
		return
	}

	e.logWriter.Write([]byte(fmt.Sprint(param...)))
}
