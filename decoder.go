package fast

import (
	"fmt"
	"encoding/hex"
)

const (
	maxLoadBytes = (32 << (^uint(0) >> 63)) * 8 / 7 // max size of 7-th bits data
)

type Decoder struct {
	repo map[uint]*Template

	buf *buffer
	current *pmap
	prev *pmap

	debug bool
}

func NewDecoder(tmps ...*Template) *Decoder {
	decoder := &Decoder{repo: make(map[uint]*Template)}
	for _, t := range tmps {
		decoder.repo[t.ID] = t
	}
	return decoder
}

func (d *Decoder) Debug() {
	d.debug = true
}

func (d *Decoder) Decode(segment []byte, msg interface{}) {
	d.buf = newBuffer(segment)

	d.log("data: ", utohex(d.buf.data))

	d.parsePmap()
	d.log("pmap: ", utohex(d.buf.data), *d.current)

	var templateID uint

	if d.current.isNextBitSet() {
		templateID = uint(d.buf.decodeUint32())
		d.log("template: ", utohex(d.buf.data), templateID)
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

	var value interface{}
	var field *Field
	for _, instruction := range tpl.Instructions {
		if instruction.Type == TypeSequence {

			length := int(instruction.Instructions[0].Visit(d.buf).(uint32))
			d.log("length", utohex(d.buf.data), length)

			if length > 0 {
				tmp := *d.current
				d.current = newPmap(d.buf)
				d.prev = &tmp
				d.log("pmap: ", utohex(d.buf.data), *d.current)
			}

			for i:=0; i<length; i++ {
				for _, internal := range instruction.Instructions[1:] {
					value = internal.Visit(d.buf)

					field = &Field{
						ID: internal.ID,
						Name: internal.Name,
						Value: value,
						Index: i,
					}
					d.log("sequence", utohex(d.buf.data), field)
					m.Assign(field)
				}
			}
		} else {
			value = instruction.Visit(d.buf)
			field := &Field{ID: instruction.ID, Name: instruction.Name, Value: value}
			d.log("field", utohex(d.buf.data), field)
			m.Assign(field)
		}
	}
}

func (d *Decoder) parsePmap() {
	d.current = newPmap(d.buf)
}

func (d *Decoder) log(a ...interface{}) {
	if d.debug {
		fmt.Println(a...)
	}
}

// ------------

func utoi(d []byte) (r []int8) {
	for _, b := range d {
		r = append(r, int8(b))
	}
	return
}

func utohex(d []byte) string {
	return hex.EncodeToString(d)
}