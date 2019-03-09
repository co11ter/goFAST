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
	values []reflect.Value
	index int
}

func newMsg(msg interface{}) *message {

	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	rt := reflect.TypeOf(msg).Elem()

	m := &message{
		tagMap: make(map[string]int),
		values: []reflect.Value{rv},
	}

	parseType(rt, m.tagMap)

	return m
}

func (m *message) Lock(field *field) bool {
	v, ok := m.lookUpRField(field)
	if !ok {
		return false
	}

	if v.Kind() == reflect.Slice {
		m.values = append(m.values, v.Index(field.num).Addr())
	} else {
		m.values = append(m.values, v.Addr())
	}
	m.index++
	return true
}

func (m *message) Unlock() {
	m.values = m.values[:m.index]
	m.index--
}

func (m *message) lookUpRField(field *field) (v reflect.Value, ok bool) {
	index := m.lookUpIndex(field)
	if index == nil {
		return
	}

	v = m.values[m.index].Elem().Field(*index)
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

func (m *message) SetLen(field *field) {
	if rField, ok := m.lookUpRField(field); ok {
		length := field.value.(int)
		if length > rField.Cap() {
			newValue := reflect.MakeSlice(rField.Type(), length, length)
			reflect.Copy(newValue, rField)
			rField.Set(newValue)
		}

		if length > rField.Len() {
			rField.SetLen(length)
		}
	}
}

// find template id in message and return
func (m *message) GetTID() uint {
	index, ok := m.tagMap["*"]
	if !ok {
		return 0
	}
	return uint(m.values[m.index].Elem().Field(index).Uint())
}

// set template id to message
func (m *message) SetTID(tid uint) {
	index, ok := m.tagMap["*"]
	if !ok {
		return
	}

	rField := m.values[m.index].Elem().Field(index)
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

func (m *message) set(field reflect.Value, value reflect.Value) {
	if field.Kind() == reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		field.Elem().Set(reflect.ValueOf(value))
		return
	}
	field.Set(value)
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

		if field.Type.Kind() == reflect.Struct {
			parseType(field.Type, tagMap)
		}

		if field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.Struct {
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
