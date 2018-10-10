package fast

import (
	"bytes"
	"io"
)

type Acceptor struct {
	prev *PMap
	current *PMap
	storage storage

	tmp *bytes.Buffer
	chunk *bytes.Buffer
	buf *Writer
	writer io.Writer
}

func newAcceptor(writer io.Writer) *Acceptor {
	return &Acceptor{
		storage: make(map[string]interface{}),
		chunk: &bytes.Buffer{},
		tmp: &bytes.Buffer{},
		writer: writer,
	}
}

func (a *Acceptor) setBuffer(buf buffer) {
	a.buf = NewWriter(buf)
}

func (a *Acceptor) writePMap() {
	a.buf.WritePMap(a.current)
}

func (a *Acceptor) writeTmp() {
	a.tmp.Write(a.buf.Bytes())
}

func (a *Acceptor) writeChunk() {
	a.chunk.Write(a.buf.Bytes())
	if a.prev != nil {
		a.current = a.prev
		a.prev = nil
	}
}

func (a *Acceptor) commit() error {
	a.current = nil
	a.prev = nil
	tmp := append(a.buf.Bytes(), a.tmp.Bytes()...)
	tmp = append(tmp, a.chunk.Bytes()...)
	_, err := a.writer.Write(tmp)
	a.chunk.Reset()
	return err
}

func (a *Acceptor) acceptPMap() {
	if a.current == nil {
		a.current = &PMap{mask: 128}
	} else {
		tmp := *a.current
		a.current = &PMap{mask: 128}
		a.prev = &tmp
	}
}

func (a *Acceptor) acceptTemplateID(id uint32) {
	a.current.SetNextBit(true)
	a.buf.WriteUint32(false, id)
}

func (a *Acceptor) accept(instruction *Instruction, value interface{}) {
	switch instruction.Opt {
	case OptNone:
		a.encode(instruction, value)
		a.storage.save(instruction.key(), value)
	case OptConstant:
		if instruction.IsOptional() {
			a.current.SetNextBit(value != nil)
		}
		a.storage.save(instruction.key(), value)
	case OptDefault:
		if instruction.Value == value {
			a.current.SetNextBit(false)
			a.storage.save(instruction.key(), value)
			return
		}
		a.current.SetNextBit(true)
		a.encode(instruction, value)
		if value != nil {
			a.storage.save(instruction.key(), value)
		}
	case OptDelta:
		if previous := a.storage.load(instruction.key()); previous != nil {
			value = delta(value, previous)
		}
		a.encode(instruction, value)
		a.storage.save(instruction.key(), value)
	case OptTail:
		// TODO
	case OptCopy, OptIncrement:
		previous := a.storage.load(instruction.key())
		a.storage.save(instruction.key(), value)
		if previous == nil {
			if instruction.Value == value {
				a.current.SetNextBit(false)
				return
			}
		} else if isEmpty(previous) {
			if value == nil {
				a.current.SetNextBit(false)
				return
			}
		}

		a.current.SetNextBit(true)
		a.encode(instruction, value)
	}
}

func (a *Acceptor) encode(instruction *Instruction, value interface{}) {
	if value == nil {
		err := a.buf.WriteNil()
		if err != nil {
			panic(err)
		}
		return
	}

	switch instruction.Type {
	case TypeUint32, TypeLength:
		err := a.buf.WriteUint32(instruction.IsNullable(), value.(uint32))
		if err != nil {
			panic(err)
		}
	case TypeUint64:
		err := a.buf.WriteUint64(instruction.IsNullable(), value.(uint64))
		if err != nil {
			panic(err)
		}
	case TypeString:
		err := a.buf.WriteAsciiString(instruction.IsNullable(), value.(string))
		if err != nil {
			panic(err)
		}
	case TypeInt64, TypeMantissa:
		err := a.buf.WriteInt64(instruction.IsNullable(), value.(int64))
		if err != nil {
			panic(err)
		}
	case TypeInt32, TypeExponent:
		err := a.buf.WriteInt32(instruction.IsNullable(), value.(int32))
		if err != nil {
			panic(err)
		}
	}
}

func isEmpty(value interface{}) bool {
	switch value.(type) {
	case int64:
		return value.(int64) == 0
	case int32:
		return value.(int32) == 0
	case uint64:
		return value.(uint64) == 0
	case uint32:
		return value.(uint32) == 0
	case int:
		return value.(int) == 0
	case uint:
		return value.(uint) == 0
	case string:
		return value.(string) == ""
	}
	return true
}
