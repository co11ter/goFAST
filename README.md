goFAST
======

[![GoDoc](https://godoc.org/github.com/co11ter/goFAST?status.svg)](https://godoc.org/github.com/co11ter/goFAST)

goFAST is a Go implementation of the FAST Protocol (FIX Adapted for STreaming).
Work-in-progress, expect bugs and missing features.

Installation
------------

Install the FAST library using the "go get" command:

    go get github.com/co11ter/goFAST

Usage
-----


```
var xmlData = `
   <?xml version="1.0" encoding="UTF-8"?>
   <templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
   	<template name="Done" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
   		<string name="Type" id="15">
   			<constant value="99"/>
   		</string>
   		<string name="Test" id="131" presence="optional"/>
   		<uInt64 name="Time" id="20" presence="optional"/>
   		<int32 name="Equal" id="271"/>
   		<sequence name="Sequence">
   			<length name="SeqLength" id="146"/>
   			<uInt64 name="SomeField" id="38"/>
   		</sequence>
   	</template>
   </templates>`
   
type Seq struct {
    SomeField uint64
}

type Msg struct {
    TemplateID  uint    `json:"*"`    // template id
    FieldByID   string  `json:"15"`   // assign value by instruction id
    FieldByName string  `json:"Test"` // assign value by instruction name
    Equal       int32                 // name of field is default value for assign
    Nullable    *uint64 `json:"20"`   // nullable - will skip, if field data is absent
    Skip        int     `json:"-"`    // skip
    Sequence    []Seq
}
```

### Decoding

```
var msg Msg
reader := bytes.NewReader(
    []byte{0xc0, 0x81, 0x74, 0x65, 0x73, 0xf4, 0x80, 0x80, 0x81, 0x80, 0x82},
)

tpls := fast.ParseXMLTemplate(strings.NewReader(xmlData))
decoder := fast.NewDecoder(
    reader,
    tpls...,
)

if err := decoder.Decode(&msg); err != nil {
    panic(err)
}
fmt.Print(msg)

```

### Encoding

```
var buf bytes.Buffer
var msg = Msg{
    TemplateID: 1,
    FieldByName: "test",
    Sequence: []Seq{
        {SomeField: 2},
    },
}

tpls := fast.ParseXMLTemplate(strings.NewReader(xmlData))
encoder := fast.NewEncoder(&buf, tpls...)

if err := encoder.Encode(&msg); err != nil {
    panic(err)
}
fmt.Printf("%x", buf.Bytes())
```