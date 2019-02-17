// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

type field struct {
	id         uint   // instruction id
	name       string // instruction name
	templateID uint   // template id
	num        int    // slice index
	multiple   bool   // has internal fields

	value interface{}
}
