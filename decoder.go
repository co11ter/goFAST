package fast

import (
	"fmt"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data

	DebugHex = "hex"
	DebugInt = "int"
)

type Decoder struct {
	repo map[uint]*Template
	visitor *Visitor

	debug string
}

func NewDecoder(tmps ...*Template) *Decoder {
	decoder := &Decoder{repo: make(map[uint]*Template), visitor: newVisitor()}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

func (d *Decoder) Debug(typ string) {
	d.debug = typ
}

func (d *Decoder) Decode(segment []byte, msg interface{}) {
	d.visitor.buf = newBuffer(segment)

	d.log("data: ")

	d.log("pmap parsing: ")
	d.visitor.visitPmap()
	d.log("  pmap parsed: ", *d.visitor.current)

	templateID := d.visitor.visitTemplateID()
	d.log("template: ", templateID)

	tpl, ok := d.repo[uint(templateID)]
	if !ok {
		return
	}

	m := newMsg(msg)
	d.decodeSegment(tpl.Instructions, m)

	return
}

func (d *Decoder) decodeSequence(instructions []*Instruction, msg *message) {
	d.log("sequence start: ")

	length := int(d.visitor.visit(instructions[0]).Value.(uint32))
	d.log("  length: ", length)

	if length > 0 {
		d.visitor.visitPmap()
		d.log("  pmap: ", *d.visitor.current)
	}

	for i:=0; i<length; i++ {
		for _, instruction := range instructions[1:] {

			d.log("  parsing: ", instruction.Name)
			field := d.visitor.visit(instruction)
			d.log("    parsed: ", field.Name, field.Value)
			msg.AssignSlice(field, i)
		}
	}
}

func (d *Decoder) decodeSegment(instructions []*Instruction, msg *message) {
	for _, instruction := range instructions {
		if instruction.Type == TypeSequence {
			d.decodeSequence(instruction.Instructions, msg)
		} else {
			d.log("parsing: ", instruction.Name)
			field := d.visitor.visit(instruction)
			d.log("  parsed: ", field.Name, field.Value)
			msg.Assign(field)
		}
	}
}

func (d *Decoder) log(label string, param ...interface{}) {
	if d.debug == "" {
		return
	}
	if d.debug == DebugHex {
		param = append([]interface{}{label, d.visitor.buf.Hex()}, param...)
		fmt.Println(param...)
	} else {
		param = append([]interface{}{label, d.visitor.buf.Int()}, param...)
		fmt.Println(param...)
	}
}