package fast

type Template struct {
	ID uint
	Name string
	Instructions []*Instruction
}

type Field struct {
	ID uint // instruction id
	Name string
	Index int
	Value interface{}
}
