package fast

type InstructionType int
type InstructionOpt int
type InstructionPresence int

const (
	TypeNull InstructionType = iota
	TypeUint32
	TypeInt32
	TypeUint64
	TypeInt64
	TypeString
	TypeDecimal
	TypeSequence
	TypeLength

	OptDefault InstructionOpt = iota
	OptConstant
	OptIncrement
	OptCopy
	OptDelta

	PresenceMandatory InstructionPresence = iota
	PresenceOptional
)

type Instruction struct {
	ID uint
	Name string
	Presence InstructionPresence
	Type InstructionType
	Opt InstructionOpt
	Instructions []*Instruction
	Value interface{}
}

// TODO
func (ins *Instruction) Visit(buf *buffer) interface{} {
	if ins.Opt == OptConstant {
		return ins.Value
	}

	if ins.Type == TypeLength {
		return buf.decodeUint32(ins.Presence == PresenceOptional)
	}

	if ins.Type == TypeUint32 {
		return buf.decodeUint32(ins.Presence == PresenceOptional)
	}

	if ins.Type == TypeUint64 {
		return buf.decodeUint64(ins.Presence == PresenceOptional)
	}

	if ins.Type == TypeString {
		buf.data = buf.data[1:] // TODO
	}

	return nil
}
