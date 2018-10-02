package fast

import (
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

