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
