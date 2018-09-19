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

	buf *buffer
	current *pmap
	prev *pmap

	debug string
}

func NewDecoder(tmps ...*Template) *Decoder {
	decoder := &Decoder{repo: make(map[uint]*Template)}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

func (d *Decoder) Debug(typ string) {
	d.debug = typ
}

func (d *Decoder) Decode(segment []byte, msg interface{}) {
	d.buf = newBuffer(segment)

	d.log("data: ")

	d.log("pmap parsing: ")
	d.decodePmap()
	d.log("  pmap parsed: ", *d.current)

	var templateID uint

	if d.current.isNextBitSet() {
		templateID = uint(d.buf.decodeUint32(false))
		d.log("template: ", templateID)
	}

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

	length := int(d.visit(instructions[0]).Value.(uint32))
	d.log("  length: ", length)

	if length > 0 {
		tmp := *d.current
		d.current = newPmap(d.buf)
		d.prev = &tmp
		d.log("  pmap: ", *d.current)
	}

	for i:=0; i<length; i++ {
		for _, internal := range instructions[1:] {

			d.log("  parsing: ", internal.Name)
			field := d.visit(internal)
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
			field := d.visit(instruction)
			d.log("  parsed: ", field.Name, field.Value)
			msg.Assign(field)
		}
	}
}

func (d *Decoder) visit(instruction *Instruction) *Field {
	if instruction.Opt == OptConstant {
		return &Field{
			ID: instruction.ID,
			Name: instruction.Name,
			Value: instruction.Value,
		}
	}

	if instruction.Type == TypeLength {
		return &Field{
			ID: instruction.ID,
			Name: instruction.Name,
			Value: d.buf.decodeUint32(instruction.IsOptional()),
		}
	}

	if instruction.Type == TypeUint32 {
		return &Field{
			ID: instruction.ID,
			Name: instruction.Name,
			Value: d.buf.decodeUint32(instruction.IsOptional()),
		}
	}

	if instruction.Type == TypeUint64 {
		return &Field{
			ID: instruction.ID,
			Name: instruction.Name,
			Value: d.buf.decodeUint64(instruction.IsOptional()),
		}
	}

	if instruction.Type == TypeString {
		d.buf.data = d.buf.data[1:] // TODO
	}

	return &Field{ID: instruction.ID, Name: instruction.Name, Value: nil}
}

func (d *Decoder) decodePmap() {
	d.current = newPmap(d.buf)
}

func (d *Decoder) log(label string, param ...interface{}) {
	if d.debug == "" {
		return
	}
	if d.debug == DebugHex {
		param = append([]interface{}{label, d.buf.Hex()}, param...)
		fmt.Println(param...)
	} else {
		param = append([]interface{}{label, d.buf.Int()}, param...)
		fmt.Println(param...)
	}
}