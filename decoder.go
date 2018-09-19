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
	d.parsePmap()
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

	d.parseFields(tpl, msg)

	return
}

func (d *Decoder) parseFields(tpl *Template, msg interface{}) {
	m := newMsg(msg)

	for _, instruction := range tpl.Instructions {
		if instruction.Type == TypeSequence {

			d.log("sequence start: ")

			length := int(d.visit(instruction.Instructions[0]).Value.(uint32))
			d.log("  length: ", length)

			if length > 0 {
				tmp := *d.current
				d.current = newPmap(d.buf)
				d.prev = &tmp
				d.log("  pmap: ", *d.current)
			}

			for i:=0; i<length; i++ {
				for _, internal := range instruction.Instructions[1:] {

					d.log("  parsing: ", internal.Name)
					field := d.visit(internal)
					d.log("    parsed: ", field.Name, field.Value)
					m.AssignSlice(field, i)
				}
			}
		} else {
			d.log("parsing: ", instruction.Name)
			field := d.visit(instruction)
			d.log("  parsed: ", field.Name, field.Value)
			m.Assign(field)
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

func (d *Decoder) parsePmap() {
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