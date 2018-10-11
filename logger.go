// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"fmt"
	"io"
)

type logger struct {
	*bytes.Buffer
	log io.Writer
}

func newLogger(logWriter io.Writer) *logger {
	return &logger{&bytes.Buffer{}, logWriter}
}

func (l *logger) Write(b []byte) (n int, err error) {
	l.Buffer.Write(b)
	return l.log.Write([]byte(fmt.Sprintf("%x", b)))
}
