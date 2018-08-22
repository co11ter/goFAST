package fast

import (
	"testing"
)

func Test_newMsg(t *testing.T) {
	type Sequense struct {
		Test uint32 `fast:"33"`
	}
	type Msg struct {
		Test uint32 `fast:"11"`
		Seq []Sequense `fast:"22"`
	}

	var msg Msg
	m := newMsg(&msg)

	if v, ok := m.tagMap["11"]; !ok || len(v) != 1 {
		t.Fatal("invalid parsing tag '11'")
	}

	if v, ok := m.tagMap["22"]; !ok || len(v) != 1 {
		t.Fatal("invalid parsing tag '22'")
	}

	if v, ok := m.tagMap["33"]; !ok || len(v) != 2 {
		t.Fatal("invalid parsing tag '33'")
	}
}
