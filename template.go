package fast

type Template struct {
	ID uint
	Name string
	Instructions []*Instruction
}

type Field struct {
	ID uint // instruction id
	Value interface{}
}

func (t *Template) Process(buf *buffer) <-chan *Field {
	ch := make(chan *Field)
	go func() {
		defer close(ch)

		var value interface{}
		for _, instruction := range t.Instructions {
			value = instruction.Visit(buf)
			ch <- &Field{ID: instruction.ID, Value: value}
		}
	}()
	return ch
}
