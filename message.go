package fast

import (
	"errors"
	"reflect"
	"strconv"
)

const structTag = "fast"

type message struct {
	tagMap map[string][]int
	msg    interface{}
}

func (m *message) Assign(field *Field) {
	if v, ok := m.tagMap[strconv.Itoa(int(field.ID))]; ok {
		_ = v
		//m.rValue.Field(v).Set(reflect.ValueOf(field.Value))
	}
	if v, ok := m.tagMap[field.Name]; ok {
		_ = v
		//m.rValue.Field(v).Set(reflect.ValueOf(field.Value))
	}
}

func newMsg(msg interface{}) *message {

	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	rt := reflect.TypeOf(msg).Elem()

	m := &message{tagMap: make(map[string][]int)}

	parseType(rt, m.tagMap, nil)

	return m
}

func parseType(rt reflect.Type, tagMap map[string][]int, index *int) {

	for i := 0; i < rt.NumField(); i++ {

		field := rt.Field(i)

		name := lookUp(field)
		if name == "" {
			continue
		}

		if _, ok := tagMap[name]; !ok {
			tagMap[name] = []int{}
		}

		if index != nil {
			tagMap[name] = append(tagMap[name], *index)
		}

		tagMap[name] = append(tagMap[name], i)

		if field.Type.Kind() == reflect.Slice {
			parseType(field.Type.Elem(), tagMap, &i)
		}
	}
}

func lookUp(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup(structTag); ok && tag != "" {
		if tag == "-" {
			return ""
		}
		return tag
	}
	return field.Name
}
