package fast

type Field struct {
	ID uint // instruction id
	Name string
	Value interface{}
}

type Visitor struct {
	prev *pmap
	current *pmap
	storage map[uint]interface{} // prev values

	buf *buffer
}

func newVisitor() *Visitor {
	return &Visitor{storage: make(map[uint]interface{})}
}

func (v *Visitor) visitPmap() {
	if v.current == nil {
		v.current = newPmap(v.buf)
	} else {
		tmp := *v.current
		v.current = newPmap(v.buf)
		v.prev = &tmp
	}
}

func (v *Visitor) visitTemplateID() uint {
	if v.current.isNextBitSet() {
		return uint(v.buf.decodeUint32(false))
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
			if v.current.isNextBitSet() {
				field.Value = instruction.Value
			}
		} else {
			field.Value = instruction.Value
		}
		v.storage[instruction.ID] = field.Value
	case OptDefault:
		if v.current.isNextBitSet() {
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
		return v.buf.decodeUint32(instruction.IsOptional())
	case TypeUint32:
		return v.buf.decodeUint32(instruction.IsOptional())
	case TypeUint64:
		return v.buf.decodeUint64(instruction.IsOptional())
	case TypeString:
		v.buf.data = v.buf.data[1:] // TODO
		return nil
	default:
		return nil
	}
}
