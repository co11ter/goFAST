// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

type Field struct {
	ID uint // instruction id
	Name string
	TemplateID uint

	Value interface{}
}


