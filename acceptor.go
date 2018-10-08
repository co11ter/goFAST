package fast

import (
	"io"
)

type Acceptor struct {
	prev *PMap
	current *PMap
	storage storage

	writer *Writer
}

func newAcceptor(writer io.Writer) *Acceptor {
	return &Acceptor{
		storage: make(map[string]interface{}),
		writer: NewWriter(writer),
	}
}

func (a *Acceptor) commit() error {
	return a.writer.commit()
}

func (a *Acceptor) acceptPMap() {
	a.current = &PMap{mask: 128}
}

func (a *Acceptor) acceptTemplateID(id uint32) {
	a.current.SetNextBit(true)
	a.writer.WriteUint32(false, &id)
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
		err := a.writer.WriteNil()
		if err != nil {
			panic(err)
		}
		return
	}

	switch instruction.Type {
	case TypeUint32, TypeLength:
		tmp := value.(uint32)
		err := a.writer.WriteUint32(instruction.IsNullable(), &tmp)
		if err != nil {
			panic(err)
		}
	case TypeUint64:
		tmp := value.(uint64)
		err := a.writer.WriteUint64(instruction.IsNullable(), &tmp)
		if err != nil {
			panic(err)
		}
	case TypeString:
		tmp := value.(string)
		err := a.writer.WriteAsciiString(instruction.IsNullable(), &tmp)
		if err != nil {
			panic(err)
		}
	case TypeInt64, TypeMantissa:
		tmp := value.(int64)
		err := a.writer.WriteInt64(instruction.IsNullable(), &tmp)
		if err != nil {
			panic(err)
		}
	case TypeInt32, TypeExponent:
		tmp := value.(int32)
		err := a.writer.WriteInt32(instruction.IsNullable(), &tmp)
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
