package fast

import (
	"io"
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

func (v *Visitor) visit(instruction *Instruction) (interface{}, error) {
	return instruction.extract(v.reader, v.storage, v.current)
}
