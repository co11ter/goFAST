package fast

import (
	"io"
	"math/big"
	"strconv"
)

type Field struct {
	ID uint // instruction id
	Name string
	Type InstructionType
	Value interface{}
}

func (f *Field) key() string {
	return strconv.Itoa(int(f.ID)) + ":" + f.Name + ":" + strconv.Itoa(int(f.Type))
}

type Visitor struct {
	prev *PMap
	current *PMap
	storage map[string]interface{} // TODO prev values

	reader *Reader
}

func newVisitor(reader io.ByteReader) *Visitor {
	return &Visitor{
		storage: make(map[string]interface{}),
		reader: NewReader(reader),
	}
}

func (v *Visitor) save(key string, value interface{}) {
	v.storage[key] = &value
}

func (v *Visitor) load(key string) interface{} {
	if value, ok := v.storage[key]; ok {
		return value
	}
	return nil
}

func (v *Visitor) visitPMap() {
	var err error
	if v.current == nil {
		v.current, err = v.reader.ReadPMap()
		if err != nil {
			panic(err)
		}
	} else {
		tmp := *v.current
		v.current, err = v.reader.ReadPMap()
		if err != nil {
			panic(err)
		}
		v.prev = &tmp
	}
}

func (v *Visitor) visitTemplateID() uint {
	if v.current.IsNextBitSet() {
		tmp, _, err := v.reader.ReadUint32(false)
		if err != nil {
			panic(err)
		}
		return uint(tmp)
	}
	return 0
}

// TODO need refactor
func (v *Visitor) visitDecimal(instruction *Instruction, field *Field) {
	var mantissa int64
	var exponent int32
	for _, in := range instruction.Instructions {
		if in.Type == TypeMantissa {
			mField := v.visit(in)
			mantissa = mField.Value.(int64)
		}
		if in.Type == TypeExponent {
			eField := v.visit(in)
			exponent = eField.Value.(int32)
		}
	}

	field.Value, _ = (&big.Float{}).SetMantExp(
		(&big.Float{}).SetInt64(mantissa),
		int(exponent),
	).Float64()
}

func (v *Visitor) visit(instruction *Instruction) *Field {
	field := &Field{
		ID: instruction.ID,
		Name: instruction.Name,
		Type: instruction.Type,
	}

	// TODO
	if instruction.Type == TypeDecimal {
		v.visitDecimal(instruction, field)
		return field
	}

	switch instruction.Opt {
	case OptNone:
		field.Value = v.decode(instruction)
		v.save(field.key(), field.Value)
	case OptConstant:
		if instruction.IsOptional() {
			if v.current.IsNextBitSet() {
				field.Value = instruction.Value
			}
		} else {
			field.Value = instruction.Value
		}
		v.save(field.key(), field.Value)
	case OptDefault:
		if v.current.IsNextBitSet() {
			field.Value = v.decode(instruction)
		} else{
			field.Value = instruction.Value
			v.save(field.key(), field.Value)
		}
	case OptDelta:
		field.Value = v.decode(instruction)
		if previous := v.load(field.key()); previous != nil {
			field.Value = sum(field.Value, previous)
		}
	case OptTail:
		// TODO
	case OptCopy, OptIncrement:
		if v.current.IsNextBitSet() {
			field.Value = v.decode(instruction)
			v.save(field.key(), field.Value)
		} else {
			if v.load(field.key()) == nil {
				field.Value = instruction.Value
				v.save(field.key(), field.Value)
			} else {
				// TODO what have to do on empty value

				field.Value = v.load(field.key())
			}
		}
		if instruction.Opt == OptIncrement {
			increment(field.Value)
		}
	}

	return field
}

func (v *Visitor) decode(instruction *Instruction) interface{} {
	switch instruction.Type {
	case TypeLength:
		tmp, _, err := v.reader.ReadUint32(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	case TypeUint32:
		tmp, _, err := v.reader.ReadUint32(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	case TypeUint64:
		tmp, _, err := v.reader.ReadUint64(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	case TypeString:
		tmp, _, err := v.reader.ReadAsciiString(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	case TypeInt64, TypeMantissa:
		tmp, _, err := v.reader.ReadInt64(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	case TypeInt32, TypeExponent:
		tmp, _, err := v.reader.ReadInt32(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		return tmp
	default:
		return nil
	}
}

// TODO need implements for string
func sum(values ...interface{}) (res interface{}) {
	switch values[0].(type) {
	case int64:
		res = values[0].(int64)+int64(toInt(values[1]))
	case int32:
		res = values[0].(int32)+int32(toInt(values[1]))
	case uint64:
		res = values[0].(uint64)+uint64(toInt(values[1]))
	case uint32:
		res = values[0].(uint32)+uint32(toInt(values[1]))
	}
	return
}

func toInt(value interface{}) int {
	switch value.(type) {
	case int64:
		return int(value.(int64))
	case int32:
		return int(value.(int32))
	case uint64:
		return int(value.(uint64))
	case uint32:
		return int(value.(uint32))
	}
	return 0
}

func increment(value interface{}) (res interface{}) {
	return sum(value, 1)
}
