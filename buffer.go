package fast

import "encoding/hex"

type buffer struct {
	data []byte
}

func newBuffer(data []byte) *buffer {
	return &buffer{data: data}
}

func (b *buffer) cutEmpty() {
	for i, c := range b.data {
		if 0 == (c & 0x80) {
			b.data = b.data[i+1:]
			return
		}
	}
}

func (b *buffer) decodeUint32(optional bool) (result uint32) {
	i := 0
	result = uint32(b.data[i] & 0x7F)

	for (b.data[i] & 0x80) == 0 {
		result <<= 7;
		i++
		result |= uint32(b.data[i] & 0x7F);
	}

	b.data = b.data[i+1:]

	if optional && result > 0 {
		result--
	}
	return result
}

func (b *buffer) decodeUint64(optional bool) (result uint64) {
	i := 0
	result = uint64(b.data[i] & 0x7F)

	for (b.data[i] & 0x80) == 0 {
		result <<= 7;
		i++
		result |= uint64(b.data[i] & 0x7F);
	}

	b.data = b.data[i+1:]

	if optional && result > 0 {
		result--
	}
	return result
}

func  decodeString() (result string) {
	/*i := 0
	result = string(b.data)
	if (b.data[i] & 0x80) == 0 {
		if b.data[i] == 0x80 {
			return result
		}
	}*/
	return ""
}

func (b *buffer) Int() (r []int8) {
	for _, x := range b.data {
		r = append(r, int8(x))
	}
	return
}

func (b *buffer) Hex() string {
	var result string
	str := hex.EncodeToString(b.data)
	for i:=0; i<len(str); i++ {
		if i%2==0 {
			result += " "
		}
		result += string(str[i])
	}
	return result
}