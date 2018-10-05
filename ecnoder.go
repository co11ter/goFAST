package fast

import "io"

type Encoder struct {
	repo map[uint]*Template
	acceptor *Acceptor

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

// TODO
func (e *Encoder) Encode(msg interface{}) error {
	return nil
}
