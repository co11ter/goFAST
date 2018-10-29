// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"math"
	"strconv"
)

// Instruction contains rules for encoding/decoding field.
type Instruction struct {
	ID           uint
	Name         string
	Presence     InstructionPresence
	Type         InstructionType
	Operator     InstructionOperator
	Instructions []*Instruction
	Value        interface{}
}

func (i *Instruction) key() string {
	return strconv.Itoa(int(i.ID)) + ":" + i.Name + ":" + strconv.Itoa(int(i.Type))
}

func (i *Instruction) isOptional() bool {
	return i.Presence == PresenceOptional
}

func (i *Instruction) isNullable() bool {
	return i.isOptional() && (i.Operator != OperatorConstant)
}

func (i *Instruction) inject(writer *writer, s storage, pmap *pMap, value interface{}) (err error) {
	switch i.Operator {
	case OperatorNone:
		err = i.write(writer, value)
		if err != nil {
			return
		}
		s.save(i.key(), value)
	case OperatorConstant:
		if i.isOptional() {
			pmap.SetNextBit(value != nil)
		}
		s.save(i.key(), value)
	case OperatorDefault:
		if i.Value == value {
			pmap.SetNextBit(false)
			s.save(i.key(), value)
			return
		}
		pmap.SetNextBit(true)
		err = i.write(writer, value)
		if err != nil {
			return
		}
		if value != nil {
			s.save(i.key(), value)
		}
	case OperatorDelta:
		if previous := s.load(i.key()); previous != nil {
			value = delta(value, previous)
		}
		err = i.write(writer, value)
		if err != nil {
			return
		}
		s.save(i.key(), value)
	case OperatorTail:
		// TODO
	case OperatorCopy, OperatorIncrement:
		previous := s.load(i.key())
		s.save(i.key(), value)
		if previous == nil {
			if i.Value == value {
				pmap.SetNextBit(false)
				return
			}
		} else if isEmpty(previous) {
			if value == nil {
				pmap.SetNextBit(false)
				return
			}
		}

		pmap.SetNextBit(true)
		err = i.write(writer, value)
	}
	return err
}

func (i *Instruction) write(writer *writer, value interface{}) (err error) {
	if value == nil {
		err = writer.WriteNil()
		return
	}

	switch i.Type {
	case TypeUint32, TypeLength:
		err = writer.WriteUint32(i.isNullable(), value.(uint32))
	case TypeUint64:
		err = writer.WriteUint64(i.isNullable(), value.(uint64))
	case TypeString:
		err = writer.WriteASCIIString(i.isNullable(), value.(string))
	case TypeInt64, TypeMantissa:
		err = writer.WriteInt64(i.isNullable(), value.(int64))
	case TypeInt32, TypeExponent:
		err = writer.WriteInt32(i.isNullable(), value.(int32))
	}
	return
}

func (i *Instruction) extract(reader *reader, s storage, pmap *pMap) (result interface{}, err error) {

	if i.Type == TypeDecimal && len(i.Instructions) > 0 {
		return i.extractDecimal(reader, s, pmap)
	}

	switch i.Operator {
	case OperatorNone:
		result, err = i.read(reader)
		if err != nil {
			return nil, err
		}
		s.save(i.key(), result)
	case OperatorConstant:
		if i.isOptional() {
			if pmap.IsNextBitSet() {
				result = i.Value
			}
		} else {
			result = i.Value
		}
		s.save(i.key(), result)
	case OperatorDefault:
		if pmap.IsNextBitSet() {
			result, err = i.read(reader)
		} else {
			result = i.Value
			s.save(i.key(), result)
		}
	case OperatorDelta:
		result, err = i.read(reader)
		if err != nil {
			return nil, err
		}
		if previous := s.load(i.key()); previous != nil {
			result = sum(result, previous)
		}
		s.save(i.key(), result)
	case OperatorTail:
		// TODO
	case OperatorCopy, OperatorIncrement:
		if pmap.IsNextBitSet() {
			result, err = i.read(reader)
			if err != nil {
				return nil, err
			}
			s.save(i.key(), result)
		} else {
			if s.load(i.key()) == nil {
				result = i.Value
				s.save(i.key(), result)
			} else {
				// TODO what have to do on empty value

				result = s.load(i.key())
				if i.Operator == OperatorIncrement {
					result = increment(result)
					s.save(i.key(), result)
				}
			}
		}
	}

	return
}

func (i *Instruction) read(reader *reader) (result interface{}, err error) {
	switch i.Type {
	case TypeUint32, TypeLength:
		tmp, err := reader.ReadUint32(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeUint64:
		tmp, err := reader.ReadUint64(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeString:
		tmp, err := reader.ReadASCIIString(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeInt64, TypeMantissa:
		tmp, err := reader.ReadInt64(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeInt32, TypeExponent:
		tmp, err := reader.ReadInt32(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeDecimal:
		exponent, err := reader.ReadInt32(i.isNullable())
		if err != nil {
			return result, err
		}
		if exponent != nil {
			mantissa, err := reader.ReadInt64(false)
			if err != nil {
				return result, err
			}
			result = math.Pow10(int(*exponent)) * float64(*mantissa)
		}
	}

	return result, err
}

func (i *Instruction) extractDecimal(reader *reader, s storage, pmap *pMap) (interface{}, error) {
	var mantissa int64
	var exponent int32
	for _, in := range i.Instructions {
		if in.Type == TypeMantissa {
			mField, err := i.extract(reader, s, pmap)
			if err != nil {
				return nil, err
			}
			mantissa = mField.(int64)
		}
		if in.Type == TypeExponent {
			eField, err := i.extract(reader, s, pmap)
			if err != nil {
				return nil, err
			}
			exponent = eField.(int32)
		}
	}

	return math.Pow10(int(exponent)) * float64(mantissa), nil
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

// TODO need implements for string
func sum(values ...interface{}) (res interface{}) {
	switch values[0].(type) {
	case int64:
		res = values[0].(int64) + int64(toInt(values[1]))
	case int32:
		res = values[0].(int32) + int32(toInt(values[1]))
	case uint64:
		res = values[0].(uint64) + uint64(toInt(values[1]))
	case uint32:
		res = values[0].(uint32) + uint32(toInt(values[1]))
	}
	return
}

// TODO need implements for string
func delta(values ...interface{}) (res interface{}) {
	switch values[0].(type) {
	case int64:
		res = values[0].(int64) - int64(toInt(values[1]))
	case int32:
		res = values[0].(int32) - int32(toInt(values[1]))
	case uint64:
		res = values[0].(uint64) - uint64(toInt(values[1]))
	case uint32:
		res = values[0].(uint32) - uint32(toInt(values[1]))
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
