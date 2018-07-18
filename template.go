package fast

import "io"

type Template struct {

}

func NewTemplate(r io.Reader) *Template {
	return &Template{}
}
