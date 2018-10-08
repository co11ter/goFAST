package fast

import (
	"io"
)

type Encoder struct {
	repo map[uint]*Template
	acceptor *Acceptor
	tid uint

	logWriter io.Writer
}

func NewEncoder(writer io.Writer, tmps ...*Template) *Encoder {
	encoder := &Encoder{
		repo: make(map[uint]*Template),
		acceptor: newAcceptor(writer),
	}
	for _, t := range tmps {
		encoder.repo[t.ID] = t
	}
	return encoder
}

func (e *Encoder) SetLog(writer io.Writer) {
	e.logWriter = writer
}

func (e *Encoder) Encode(msg interface{}) error {
	m := newMsg(msg)
	e.tid = m.LookUpTID()

	tpl, ok := e.repo[e.tid]
	if !ok {
		return nil
	}

	e.acceptor.acceptPMap()
	e.acceptor.acceptTemplateID(uint32(e.tid))

	e.encodeSegment(tpl.Instructions, m)

	return nil
}

func (e *Encoder) encodeSegment(instructions []*Instruction, msg *message) {
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			e.encodeSequence(instruction.Instructions, msg)
		} else {
			field := &Field{
				ID: instruction.ID,
				Name: instruction.Name,
				TemplateID: e.tid,
			}

			msg.LookUp(field)
			e.acceptor.accept(instruction, field.Value)
		}
	}
	e.acceptor.commit()
}

func (e *Encoder) encodeSequence(instructions []*Instruction, msg *message) {

}
