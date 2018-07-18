package fast

type Decoder struct {

}

func NewDecoder(t *Template) *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(data []byte) *Message {
	return &Message{}
}