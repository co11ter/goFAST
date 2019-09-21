// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"sync"
)

// Field contains value for decoding/encoding
type Field struct {
	ID    uint
	Name  string
	Value interface{}

	index *int // message field index for reflection
}

var fieldPool = sync.Pool{
	New: func() interface{} {
		return &Field{}
	},
}

func acquireField() *Field {
	return fieldPool.Get().(*Field)
}

func releaseField(field *Field) {
	field.ID = 0
	field.Name = ""
	field.Value = nil
	field.index = nil
	fieldPool.Put(field)
}