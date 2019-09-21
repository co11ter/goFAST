// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

// Sender is interface for getting data avoid reflection.
type Sender interface {
	// GetTemplateID must return template id for message.
	GetTemplateID() uint

	// GetValue must set actual value to Field.Value for Field.Name or Field.ID.
	GetValue(*Field)

	// GetLength must set actual sequence length to Field.Value for Field.Name or Field.ID.
	GetLength(*Field)

	// Lock indicates a group or sequence. Field.Value will contain index of sequence
	Lock(*Field) bool
	Unlock()
}

// Receiver is interface for setting data avoid reflection.
type Receiver interface {
	// SetTemplateID indicates template id for message.
	SetTemplateID(uint)

	// SetValue indicates actual Field.Value for Field.Name or Field.ID.
	SetValue(*Field)

	// SetLength indicates length of sequence.
	SetLength(*Field)

	// Lock indicates a group or sequence. Field.Value will contain index of sequence
	Lock(*Field) bool
	Unlock()
}
