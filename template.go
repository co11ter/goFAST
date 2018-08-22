package fast

type Template struct {
	ID uint
	Name string
	Instructions []*Instruction
}

type Field struct {
	ID uint // instruction id
	Name string
	Value interface{}
}

func (t *Template) Process(buf *buffer) <-chan *Field {
	ch := make(chan *Field)
	go func() {
		defer close(ch)

		var value interface{}
		for _, instruction := range t.Instructions {
			if instruction.Type == TypeSequence {
				_ = buf.decodeUint32() // length
				for _, internal := range instruction.Instructions {
					value = internal.Visit(buf)
					ch <- &Field{ID: internal.ID, Name: internal.Name, Value: value}
				}

			}
			value = instruction.Visit(buf)
			ch <- &Field{ID: instruction.ID, Name: instruction.Name, Value: value}
		}
	}()
	return ch
}
