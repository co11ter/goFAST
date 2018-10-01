package fast

import (
	"bytes"
	"io"
)

type Writer struct {
	writer io.Writer
	buf bytes.Buffer
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer, bytes.Buffer{}}
}
