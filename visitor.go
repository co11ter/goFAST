package fast

import (
	"io"
	"math/big"
)

type Visitor struct {
	prev *PMap
	current *PMap
	storage storage

	reader *Reader
}

func newVisitor(reader io.ByteReader) *Visitor {
	return &Visitor{
		storage: newStorage(),
		reader: NewReader(reader),
	}
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
		tmp, err := v.reader.ReadUint32(false)
		if err != nil {
			panic(err)
		}
		return uint(*tmp)
	}
	return 0
}

// TODO need refactor
func (v *Visitor) visitDecimal(instruction *Instruction) interface{} {
	var mantissa int64
	var exponent int32
	for _, in := range instruction.Instructions {
		if in.Type == TypeMantissa {
			mField := v.visit(in)
			mantissa = mField.(int64)
		}
		if in.Type == TypeExponent {
			eField := v.visit(in)
			exponent = eField.(int32)
		}
	}

	result, _ := (&big.Float{}).SetMantExp(
		(&big.Float{}).SetInt64(mantissa),
		int(exponent),
	).Float64()
	return result
}

func (v *Visitor) visit(instruction *Instruction) (result interface{}) {

	// TODO
	if instruction.Type == TypeDecimal {
		return v.visitDecimal(instruction)
	}

	switch instruction.Opt {
	case OptNone:
		result = v.decode(instruction)
		v.storage.save(instruction.key(), result)
	case OptConstant:
		if instruction.IsOptional() {
			if v.current.IsNextBitSet() {
				result = instruction.Value
			}
		} else {
			result = instruction.Value
		}
		v.storage.save(instruction.key(), result)
	case OptDefault:
		if v.current.IsNextBitSet() {
			result = v.decode(instruction)
		} else{
			result = instruction.Value
			v.storage.save(instruction.key(), result)
		}
	case OptDelta:
		result = v.decode(instruction)
		if previous := v.storage.load(instruction.key()); previous != nil {
			result = sum(result, previous)
		}
		v.storage.save(instruction.key(), result)
	case OptTail:
		// TODO
	case OptCopy, OptIncrement:
		if v.current.IsNextBitSet() {
			result = v.decode(instruction)
			v.storage.save(instruction.key(), result)
		} else {
			if v.storage.load(instruction.key()) == nil {
				result = instruction.Value
				v.storage.save(instruction.key(), result)
			} else {
				// TODO what have to do on empty value

				result = v.storage.load(instruction.key())
				if instruction.Opt == OptIncrement {
					result = increment(result)
					v.storage.save(instruction.key(), result)
				}
			}
		}
	}

	return result
}

func (v *Visitor) decode(instruction *Instruction) interface{} {
	switch instruction.Type {
	case TypeUint32, TypeLength:
		tmp, err := v.reader.ReadUint32(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		if tmp != nil {
			return *tmp
		}
	case TypeUint64:
		tmp, err := v.reader.ReadUint64(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		if tmp != nil {
			return *tmp
		}
	case TypeString:
		tmp, err := v.reader.ReadAsciiString(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		if tmp != nil {
			return *tmp
		}
	case TypeInt64, TypeMantissa:
		tmp, err := v.reader.ReadInt64(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		if tmp != nil {
			return *tmp
		}
	case TypeInt32, TypeExponent:
		tmp, err := v.reader.ReadInt32(instruction.IsNullable())
		if err != nil {
			panic(err)
		}
		if tmp != nil {
			return *tmp
		}
	}
	return nil
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
	case int:
		return value.(int)
	case uint:
		return int(value.(uint))
	}
	return 0
}

func increment(value interface{}) (res interface{}) {
	return sum(value, 1)
}
