package fast

import "io"

type Acceptor struct {
	prev *PMap
	current *PMap
	storage map[string]interface{} // TODO prev values

	writer *Writer
}

func newAcceptor(writer io.Writer) *Acceptor {
	return &Acceptor{
		storage: make(map[string]interface{}),
		writer: NewWriter(writer),
	}
}

func (a *Acceptor) acceptPMap() {
	a.current = &PMap{mask: 128}
}

func (a *Acceptor) acceptTemplateID(id uint32) {
	a.current.SetNextBit(true)
	a.writer.WriteUint32(false, &id)
}

func (a *Acceptor) accept(instruction *Instruction, field *Field) {

}
