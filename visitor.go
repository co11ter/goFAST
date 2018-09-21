package fast

type Field struct {
	ID uint // instruction id
	Name string
	Value interface{}
}

type Visitor struct {
	prev *PMap
	current *PMap
	storage map[uint]interface{} // TODO prev values

	reader *Reader
}

func newVisitor(reader *Reader) *Visitor {
	return &Visitor{storage: make(map[uint]interface{}), reader: reader}
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

func (v *Visitor) visit(instruction *Instruction) *Field {
	field := &Field{
		ID: instruction.ID,
		Name: instruction.Name,
	}

	switch instruction.Opt {
	case OptNone:
		field.Value = v.decode(instruction)
		v.storage[instruction.ID] = field.Value
	case OptConstant:
		if instruction.IsOptional() {
			if v.current.IsNextBitSet() {
				field.Value = instruction.Value
			}
		} else {
			field.Value = instruction.Value
		}
		v.storage[instruction.ID] = field.Value
	case OptDefault:
		if v.current.IsNextBitSet() {
			field.Value = v.decode(instruction)
		} else{
			field.Value = instruction.Value
			v.storage[instruction.ID] = field.Value
		}
	case OptDelta:
		// TODO
	case OptTail:
		// TODO
	case OptCopy, OptIncrement: // TODO
		/*if v.current.isNextBitSet() {
			field.Value = v.decode(instruction)
			v.storage[instruction.ID] = field.Value
		}*/
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
	default:
		return nil
	}
}
