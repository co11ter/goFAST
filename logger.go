// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

import (
	"bytes"
	"fmt"
	"io"
)

type writerLog struct {
	*bytes.Buffer
	log io.Writer
}

func wrapWriterLog(writer io.Writer) *writerLog {
	return &writerLog{&bytes.Buffer{}, writer}
}

func (l *writerLog) Write(b []byte) (n int, err error) {
	l.log.Write([]byte(fmt.Sprintf("%x", b)))
	return 	l.Buffer.Write(b)
}

func (l *writerLog) Log(param ...interface{}) {
	l.log.Write([]byte(fmt.Sprint(param...)))
}

type readerLog struct {
	io.ByteReader
	log io.Writer
}

func wrapReaderLog(reader io.ByteReader, writer io.Writer) *readerLog {
	return &readerLog{reader,writer}
}

func (l *readerLog) ReadByte() (byte, error) {
	b, err := l.ByteReader.ReadByte()
	if err == nil {
		if _, err := l.log.Write([]byte(fmt.Sprintf("%x", b))); err != nil {
			return b, err
		}
	}
	return b, err
}

func (l *readerLog) Log(param ...interface{}) {
	l.log.Write([]byte(fmt.Sprint(param...)))
}
