// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

// Instruction contains rules for encoding/decoding field.
type Instruction struct {
	ID           uint
	Name         string
	Presence     InstructionPresence
	Type         InstructionType
	Operator     InstructionOperator
	Instructions []*Instruction
	Value        interface{}

	pMapSize int
	key   string
}

func (i *Instruction) isValid() bool {
	if i.Operator == OperatorDelta && (i.Type < TypeUint32 || i.Type > TypeMantissa) {
		return false
	}

	if i.Operator == OperatorTail && (i.Type < TypeAsciiString || i.Type > TypeByteVector) {
		return false
	}

	return true
}

func (i *Instruction) isOptional() bool {
	return i.Presence == PresenceOptional
}

func (i *Instruction) isNullable() bool {
	return i.isOptional() && (i.Operator != OperatorConstant)
}

func (i *Instruction) hasPmapBit() bool {
	return i.Operator > OperatorDelta || (i.Operator == OperatorConstant && i.isOptional())
}

func (i *Instruction) inject(writer *writer, s storage, pmap *pMap, value interface{}) (err error) {

	if i.Type == TypeDecimal && len(i.Instructions) > 0 {
		return i.injectDecimal(writer, s, pmap, value)
	}

	switch i.Operator {
	case OperatorNone:
		err = i.write(writer, value)
		if err != nil {
			return
		}
		s.save(i.key, value)
	case OperatorConstant:
		if i.isOptional() {
			pmap.SetNextBit(value != nil)
		}
		s.save(i.key, value)
	case OperatorDefault:
		if i.Value == value {
			pmap.SetNextBit(false)
			s.save(i.key, value)
			return
		}
		pmap.SetNextBit(true)
		err = i.write(writer, value)
		if err != nil {
			return
		}
		if value != nil {
			s.save(i.key, value)
		}
	case OperatorDelta:
		if previous := s.load(i.key); previous != nil {
			value = delta(value, previous)
		}
		err = i.write(writer, value)
		if err != nil {
			return
		}
		s.save(i.key, value)
	case OperatorTail:
		// TODO
	case OperatorCopy, OperatorIncrement:
		previous := s.load(i.key)
		s.save(i.key, value)
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
	case TypeByteVector:
		err = writer.WriteByteVector(i.isNullable(), value.([]byte))
	case TypeUint32, TypeLength:
		err = writer.WriteUint(i.isNullable(), uint64(value.(uint32)), maxSize32)
	case TypeUint64:
		err = writer.WriteUint(i.isNullable(), value.(uint64), maxSize64)
	case TypeAsciiString:
		err = writer.WriteString(i.isNullable(), value.(string))
	case TypeUnicodeString:
		err = writer.WriteByteVector(i.isNullable(), []byte(value.(string)))
	case TypeInt64, TypeMantissa:
		err = writer.WriteInt(i.isNullable(), value.(int64), maxSize64)
	case TypeInt32, TypeExponent:
		err = writer.WriteInt(i.isNullable(), int64(value.(int32)), maxSize32)
	case TypeDecimal:
		mantissa, exponent := newMantExp(value.(float64))
		err = writer.WriteInt(i.isNullable(), int64(exponent), maxSize32)
		if err != nil {
			return
		}
		err = writer.WriteInt(false, mantissa, maxSize64)
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
		s.save(i.key, result)
	case OperatorConstant:
		if i.isOptional() {
			if pmap.IsNextBitSet() {
				result = i.Value
			}
		} else {
			result = i.Value
		}
		s.save(i.key, result)
	case OperatorDefault:
		if pmap.IsNextBitSet() {
			result, err = i.read(reader)
		} else {
			result = i.Value
			s.save(i.key, result)
		}
	case OperatorDelta:
		result, err = i.read(reader)
		if err != nil {
			return nil, err
		}
		if previous := s.load(i.key); previous != nil {
			result = sum(result, previous)
		}
		s.save(i.key, result)
	case OperatorTail:
		// TODO
	case OperatorCopy, OperatorIncrement:
		if pmap.IsNextBitSet() {
			result, err = i.read(reader)
			if err != nil {
				return nil, err
			}
			s.save(i.key, result)
		} else {
			if s.load(i.key) == nil {
				result = i.Value
				s.save(i.key, result)
			} else {
				// TODO what have to do on empty value

				result = s.load(i.key)
				if i.Operator == OperatorIncrement {
					result = increment(result)
					s.save(i.key, result)
				}
			}
		}
	}

	return
}

func (i *Instruction) read(reader *reader) (result interface{}, err error) {
	switch i.Type {
	case TypeByteVector:
		tmp, err := reader.ReadByteVector(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeUint32, TypeLength:
		tmp, err := reader.ReadUint(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = uint32(*tmp)
		}
	case TypeUint64:
		tmp, err := reader.ReadUint(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeAsciiString:
		tmp, err := reader.ReadString(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeUnicodeString:
		tmp, err := reader.ReadByteVector(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = string(*tmp)
		}
	case TypeInt64, TypeMantissa:
		tmp, err := reader.ReadInt(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = *tmp
		}
	case TypeInt32, TypeExponent:
		tmp, err := reader.ReadInt(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			result = int32(*tmp)
		}
	case TypeDecimal:
		tmp, err := reader.ReadInt(i.isNullable())
		if err != nil {
			return result, err
		}
		if tmp != nil {
			exponent := int32(*tmp)
			mantissa, err := reader.ReadInt(false)
			if err != nil {
				return result, err
			}
			result = newFloat(*mantissa, exponent)
		}
	}

	return result, err
}

func (i *Instruction) injectDecimal(writer *writer, s storage, pmap *pMap, value interface{}) (err error) {
	mantissa, exponent := newMantExp(value.(float64))
	for _, in := range i.Instructions {
		if in.Type == TypeMantissa {
			err = in.inject(writer, s, pmap, mantissa)
			if err != nil {
				return
			}
		}
		if in.Type == TypeExponent {
			err = in.inject(writer, s, pmap, exponent)
			if err != nil {
				return
			}
		}
	}

	return
}

func (i *Instruction) extractDecimal(reader *reader, s storage, pmap *pMap) (interface{}, error) {
	var mantissa int64
	var exponent int32
	for _, in := range i.Instructions {
		if in.Type == TypeMantissa {
			mField, err := in.extract(reader, s, pmap)
			if err != nil {
				return nil, err
			}
			mantissa = mField.(int64)
		}
		if in.Type == TypeExponent {
			eField, err := in.extract(reader, s, pmap)
			if err != nil {
				return nil, err
			}
			exponent = eField.(int32)
		}
	}

	return newFloat(mantissa, exponent), nil
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
