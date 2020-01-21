package fast_test

import (
	"github.com/co11ter/goFAST"
	"strings"
	"testing"
)

var (
	xmlErrS2 = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Test" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<string name="Type" id="15">
			<delta/>
		</string>
	</template>
</templates>`

	xmlErrS3 = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Test" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<int32 name="Type" id="15">
			<constant value="abc"/>
		</int32>
	</template>
</templates>`

	xmlErrS4 = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Test" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<string name="Type" id="15">
			<constant/>
		</string>
	</template>
</templates>`

	xmlErrS5 = `
<?xml version="1.0" encoding="UTF-8"?>
<templates xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
	<template name="Test" id="1" xmlns="http://www.fixprotocol.org/ns/fast/td/1.1">
		<string name="Type" id="15">
			<default/>
		</string>
	</template>
</templates>`
)

func TestParseXMLTemplate(t *testing.T) {
	checkErr(t, xmlErrS2, fast.ErrS2)
	checkErr(t, xmlErrS3, fast.ErrS3)
	checkErr(t, xmlErrS4, fast.ErrS4)
	checkErr(t, xmlErrS5, fast.ErrS5)
}

func checkErr(t *testing.T, data string, err error) {
	_, got := fast.ParseXMLTemplate(strings.NewReader(data))
	if got != err {
		t.Fatal("not found err: '", err, "' got '", got, "'")
	}
}
