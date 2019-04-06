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

type register struct {
	prefer bool // true for map by id
	byName map[string]int
	byID   map[int]int
}

type message struct {
	current *register
	cache map[string]*register
	values []reflect.Value
	index int
}

func newMsg() *message {
	return &message{
		cache: make(map[string]*register),
		current: &register{byName: make(map[string]int), byID: make(map[int]int)},
	}
}

func (m *message) Reset(msg interface{}) {
	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	m.values = []reflect.Value{rv}
	rt := reflect.TypeOf(msg).Elem()

	var ok bool
	name := rt.PkgPath() + "." + rt.Name()
	if m.current, ok = m.cache[name]; !ok {
		m.current = &register{byName: make(map[string]int), byID: make(map[int]int)}
		countID, countName := parseType(rt, m.current)
		if countID >= countName {
			m.current.prefer = true
		}
		m.cache[name] = m.current
	}
}

func (m *message) Lock(field *field) bool {
	v, ok := m.lookUpRField(field)
	if !ok {
		return false
	}

	if v.Kind() == reflect.Slice {
		v = extractValue(v.Index(field.num))
		m.values = append(m.values, v.Addr())
	} else {
		v = extractValue(v)
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
	if field.index == nil {
		m.lookUpIndex(field)
	}
	if field.index == nil {
		return
	}

	v = extractValue(m.values[m.index])
	v = extractValue(v.Field(*field.index))
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
	index, ok := m.current.byName["*"]
	if !ok {
		return 0
	}
	return uint(m.values[m.index].Elem().Field(index).Uint())
}

// set template id to message
func (m *message) SetTID(tid uint) {
	index, ok := m.current.byName["*"]
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
		field = field.Elem()
	}
	if field.Kind() == reflect.Slice {
		newValue := reflect.MakeSlice(field.Type(), value.Len(), value.Len())
		reflect.Copy(newValue, value)
		field.Set(newValue)
	} else {
		field.Set(value)
	}
}

func (m *message) lookUpIndex(field *field) {
	var v int
	var ok bool
	if m.current.prefer {
		if v, ok = m.current.byID[int(field.id)]; ok {
			field.index = &v
			return
		}
		if v, ok = m.current.byName[field.name]; ok {
			field.index = &v
			return
		}
	}
	if v, ok = m.current.byName[field.name]; ok {
		field.index = &v
		return
	}
	if v, ok = m.current.byID[int(field.id)]; ok {
		field.index = &v
	}
}

func parseType(rt reflect.Type, current *register) (countID, countName int) {
	var (
		field reflect.StructField
		tmp reflect.Type
		name string
		id int
		err error
	)
	for i := 0; i < rt.NumField(); i++ {

		field = rt.Field(i)

		name = lookUpTag(field)
		if name == "" {
			continue
		}

		id, err = strconv.Atoi(name)
		if err == nil {
			countID++
			current.byID[id] = i
		} else {
			current.byName[name] = i
			countName++
		}

		tmp = extractType(field.Type)

		// extract first element of slice
		if tmp.Kind() == reflect.Slice {
			tmp = extractType(tmp.Elem())
		}

		if tmp.Kind() == reflect.Struct {
			d, n := parseType(tmp, current)
			countID += d
			countID += n
		}
	}
	return
}

func extractValue(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return rv.Elem()
	}
	return rv
}

func extractType(rt reflect.Type) reflect.Type {
	if rt.Kind() == reflect.Ptr {
		return rt.Elem()
	}
	return rt
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
