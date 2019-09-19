// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

type Sender interface {
	GetTemplateID() uint
	GetValue(*Field)
	GetLength(*Field)
	Lock(*Field) bool
	Unlock()
}

type Receiver interface {
	SetTemplateID(uint)
	SetValue(*Field)
	SetLength(*Field)
	Lock(*Field) bool
	Unlock()
}
