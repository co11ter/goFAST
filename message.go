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
	tagMap map[string]int
	msg    interface{}
}

func newMsg(msg interface{}) *message {

	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	rt := reflect.TypeOf(msg).Elem()

	m := &message{tagMap: make(map[string]int), msg: msg}

	parseType(rt, m.tagMap)

	return m
}

func (m *message) lookUpRField(field *field) (v reflect.Value, ok bool) {
	index := m.lookUpIndex(field)
	if index == nil {
		return
	}

	v = reflect.ValueOf(m.msg).Elem().Field(*index)
	ok = true
	return
}

// find value in message and assign to field
func (m *message) Get(field *field) {
	if rField, ok := m.lookUpRField(field); ok {
		if rField.Kind() == reflect.Ptr {
			if !rField.IsNil() {
				field.value = rField.Elem().Interface()
			}
		} else {
			field.value = rField.Interface()
		}
	}
}

// find slice len in message and assign to field
func (m *message) GetLen(field *field) {
	if rField, ok := m.lookUpRField(field); ok {
		field.value = rField.Len()
	}
}

// find value in message and assign to field
func (m *message) GetSlice(field *field) {
	rField, ok := m.lookUpRField(field.parent)
	if !ok {
		return
	}

	index := m.lookUpIndex(field)
	if index == nil {
		return
	}

	rField = rField.Index(field.num).Field(*index)

	if rField.Kind() == reflect.Ptr {
		if !rField.IsNil() {
			field.value = rField.Elem().Interface()
		}
	} else {
		field.value = rField.Interface()
	}
}

// find template id in message and return
func (m *message) GetTID() uint {
	index, ok := m.tagMap["*"]
	if !ok {
		return 0
	}
	return uint(reflect.ValueOf(m.msg).Elem().Field(index).Uint())
}

// set template id to message
func (m *message) SetTID(tid uint) {
	index, ok := m.tagMap["*"]
	if !ok {
		return
	}

	rField := reflect.ValueOf(m.msg).Elem().Field(index)
	m.set(rField, reflect.ValueOf(tid))
}

// set field value to message
func (m *message) Set(field *field) {
	if field.value == nil {
		return
	}

	if rField, ok := m.lookUpRField(field); ok {
		m.set(rField, reflect.ValueOf(field.value))
	}
}

func (m *message) SetSlice(field *field) {
	if field.value == nil {
		return
	}

	rField, ok := m.lookUpRField(field.parent)
	if !ok {
		return
	}

	if field.num >= rField.Cap() {
		newCap := rField.Cap() + rField.Cap()/2
		if newCap < 4 {
			newCap = 4
		}
		newValue := reflect.MakeSlice(rField.Type(), rField.Len(), newCap)
		reflect.Copy(newValue, rField)
		rField.Set(newValue)
	}

	if field.num >= rField.Len() {
		rField.SetLen(field.num + 1)
	}

	index := m.lookUpIndex(field)
	if index == nil {
		return
	}

	rField = rField.Index(field.num).Field(*index)
	m.set(rField, reflect.ValueOf(field.value))
}

func (m *message) set(field reflect.Value, value reflect.Value) {
	if field.Kind() == reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		field.Elem().Set(reflect.ValueOf(value))
		return
	}
	field.Set(reflect.ValueOf(value))
}

func (m *message) lookUpIndex(field *field) *int {
	id := strconv.Itoa(int(field.id))

	if v, ok := m.tagMap[id]; ok {
		return &v
	}

	if v, ok := m.tagMap[field.name]; ok {
		return &v
	}

	return nil
}

func parseType(rt reflect.Type, tagMap map[string]int) {

	for i := 0; i < rt.NumField(); i++ {

		field := rt.Field(i)

		name := lookUpTag(field)
		if name == "" {
			continue
		}

		tagMap[name] = i

		if field.Type.Kind() == reflect.Slice {
			parseType(field.Type.Elem(), tagMap)
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
