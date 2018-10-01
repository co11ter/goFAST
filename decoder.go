package fast

import (
	"fmt"
	"io"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data
)

type Decoder struct {
	repo map[uint]*Template
	visitor *Visitor

	writer io.Writer
}

func NewDecoder(reader io.ByteReader, tmps ...*Template) *Decoder {
	decoder := &Decoder{
		repo: make(map[uint]*Template),
		visitor: newVisitor(reader),
	}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

func (d *Decoder) SetLog(writer io.Writer) {
	d.writer = writer
}

func (d *Decoder) Decode(msg interface{}) error {
	d.log("// ----- new message start ----- //\n")
	d.log("pmap parsing: ")
	d.visitor.visitPMap()
	d.log("\n  pmap = ", *d.visitor.current, "\n")

	d.log("template parsing: ")
	templateID := d.visitor.visitTemplateID()
	d.log("\n  template = ", templateID, "\n")

	tpl, ok := d.repo[uint(templateID)]
	if !ok {
		return nil
	}

	m := newMsg(msg)
	d.decodeSegment(tpl.Instructions, m)

	return nil
}

func (d *Decoder) decodeSequence(instructions []*Instruction, msg *message) {
	d.log("sequence start: ")

	length := int(d.visitor.visit(instructions[0]).Value.(uint32))
	d.log("\n  length = ", length, "\n")

	for i:=0; i<length; i++ {
		d.log("sequence elem[", i, "] start: \n")
		d.log("pmap parsing: ")
		d.visitor.visitPMap()
		d.log("\n  pmap = ", *d.visitor.current, "\n")
		for _, instruction := range instructions[1:] {

			d.log("  parsing: ", instruction.Name, " ")
			field := d.visitor.visit(instruction)
			d.log("\n    ", field.Name, " = ", field.Value, "\n")
			msg.AssignSlice(field, i)
		}
	}
}

func (d *Decoder) decodeSegment(instructions []*Instruction, msg *message) {
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			d.decodeSequence(instruction.Instructions, msg)
		} else {
			d.log("parsing: ", instruction.Name, " ")
			field := d.visitor.visit(instruction)
			d.log("\n  ", field.Name, " = ", field.Value, "\n")
			msg.Assign(field)
		}
	}
}

func (d *Decoder) log(param ...interface{}) {
	if d.writer == nil {
		return
	}

	d.writer.Write([]byte(fmt.Sprint(param...)))
}