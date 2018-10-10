package fast

import (
	"bytes"
	"io"
)

type Acceptor struct {
	prev *PMap
	current *PMap
	storage storage

	tmp *bytes.Buffer
	chunk *bytes.Buffer
	buf *Writer
	writer io.Writer
}

func newAcceptor(writer io.Writer) *Acceptor {
	return &Acceptor{
		storage: make(map[string]interface{}),
		chunk: &bytes.Buffer{},
		tmp: &bytes.Buffer{},
		writer: writer,
	}
}

func (a *Acceptor) setBuffer(buf buffer) {
	a.buf = NewWriter(buf)
}

func (a *Acceptor) writePMap() {
	a.buf.WritePMap(a.current)
}

func (a *Acceptor) writeTmp() {
	a.tmp.Write(a.buf.Bytes())
}

func (a *Acceptor) writeChunk() {
	a.chunk.Write(a.buf.Bytes())
	if a.prev != nil {
		a.current = a.prev
		a.prev = nil
	}
}

func (a *Acceptor) commit() error {
	a.current = nil
	a.prev = nil
	tmp := append(a.buf.Bytes(), a.tmp.Bytes()...)
	tmp = append(tmp, a.chunk.Bytes()...)
	_, err := a.writer.Write(tmp)
	a.chunk.Reset()
	return err
}

func (a *Acceptor) acceptPMap() {
	if a.current == nil {
		a.current = &PMap{mask: 128}
	} else {
		tmp := *a.current
		a.current = &PMap{mask: 128}
		a.prev = &tmp
	}
}

func (a *Acceptor) acceptTemplateID(id uint32) {
	a.current.SetNextBit(true)
	a.buf.WriteUint32(false, id)
}

func (a *Acceptor) accept(instruction *Instruction, value interface{}) error {
	return instruction.inject(a.buf, a.storage, a.current, value)
}
