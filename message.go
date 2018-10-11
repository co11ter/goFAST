// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

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

func newMsg(msg interface{}) *message {

	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	rt := reflect.TypeOf(msg).Elem()

	m := &message{tagMap: make(map[string][]int), msg: msg}

	parseType(rt, m.tagMap, nil)

	return m
}

func (m *message) LookUp(field *field) {
	path := m.lookUpPath(field)
	if len(path) == 0 {
		return
	}

	if rField := reflect.ValueOf(m.msg).Elem().Field(path[0]); rField.Kind() == reflect.Ptr {
		if !rField.IsNil() {
			field.value = rField.Elem().Interface()
		}
	} else {
		field.value = rField.Interface()
	}
}

func (m *message) LookUpLen(field *field) {
	path := m.lookUpPath(field)
	if len(path) == 0 {
		return
	}

	field.value = reflect.ValueOf(m.msg).Elem().Field(path[0]).Len()
}

func (m *message) LookUpSlice(field *field, index int) {
	path := m.lookUpPath(field)
	if len(path) < 2 {
		return
	}

	rField := reflect.ValueOf(m.msg).Elem().Field(path[0]).Index(index).Field(path[1])
	if rField.Kind() == reflect.Ptr {
		if !rField.IsNil() {
			field.value = rField.Elem().Interface()
		}
	} else {
		field.value = rField.Interface()
	}
}

func (m *message) LookUpTID() uint {
	path, ok := m.tagMap["*"]
	if !ok {
		return 0
	}
	return uint(reflect.ValueOf(m.msg).Elem().Field(path[0]).Uint())
}

func (m *message) assignTID(tid uint) {
	path, ok := m.tagMap["*"]
	if !ok {
		return
	}

	if rField := reflect.ValueOf(m.msg).Elem().Field(path[0]); rField.Kind() == reflect.Ptr {
		rField.Set(reflect.New(rField.Type().Elem()))
		rField.Elem().Set(reflect.ValueOf(tid))
	} else {
		rField.Set(reflect.ValueOf(tid))
	}
}

func (m *message) Assign(field *field) {
	if field.value == nil {
		return
	}

	path := m.lookUpPath(field)
	if len(path) == 0 {
		return
	}

	if rField := reflect.ValueOf(m.msg).Elem().Field(path[0]); rField.Kind() == reflect.Ptr {
		rField.Set(reflect.New(rField.Type().Elem()))
		rField.Elem().Set(reflect.ValueOf(field.value))
	} else {
		rField.Set(reflect.ValueOf(field.value))
	}
}

func (m *message) AssignSlice(field *field, index int) {
	if field.value == nil {
		return
	}

	path := m.lookUpPath(field)
	if len(path) < 2 {
		return
	}

	value := reflect.ValueOf(m.msg).Elem().Field(path[0])
	if index >= value.Cap() {
		newCap := value.Cap() + value.Cap()/2
		if newCap < 4 {
			newCap = 4
		}
		newValue := reflect.MakeSlice(value.Type(), value.Len(), newCap)
		reflect.Copy(newValue, value)
		value.Set(newValue)
	}

	if index >= value.Len() {
		value.SetLen(index + 1)
	}

	if rField := value.Index(index).Field(path[1]); rField.Kind() == reflect.Ptr {
		rField.Set(reflect.New(rField.Type().Elem()))
		rField.Elem().Set(reflect.ValueOf(field.value))
	} else {
		rField.Set(reflect.ValueOf(field.value))
	}
}

func (m *message) lookUpPath(field *field) []int {
	name := strconv.Itoa(int(field.id))
	tid  := strconv.Itoa(int(field.templateID))
	if v, ok := m.tagMap[name + "," + tid]; ok {
		return v
	}

	if v, ok := m.tagMap[name]; ok {
		return v
	}

	if v, ok := m.tagMap[field.name + "," + tid]; ok {
		return v
	}

	if v, ok := m.tagMap[field.name]; ok {
		return v
	}

	return []int{}
}

func parseType(rt reflect.Type, tagMap map[string][]int, index *int) {

	for i := 0; i < rt.NumField(); i++ {

		field := rt.Field(i)

		name := lookUpTag(field)
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

func lookUpTag(field reflect.StructField) string {
	if tag, ok := field.Tag.Lookup(structTag); ok && tag != "" {
		if tag == "-" {
			return ""
		}
		return tag
	}
	return field.Name
}
