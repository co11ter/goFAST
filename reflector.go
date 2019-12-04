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

var regCache = make(map[string]*register)

type register struct {
	prefer bool // true for map by id
	byName map[string]int
	byID   map[int]int
}

type reflector struct {
	current *register
	values  []reflect.Value
	index   int
}

func makeMsg(msg interface{}) (m *reflector) {
	rv := reflect.ValueOf(msg)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		panic(errors.New("message is not pointer or nil"))
	}

	m = &reflector{values: []reflect.Value{rv}}
	rt := reflect.TypeOf(msg).Elem()

	var ok bool
	name := rt.PkgPath() + "." + rt.Name()
	if m.current, ok = regCache[name]; !ok {
		m.current = &register{byName: make(map[string]int), byID: make(map[int]int)}
		countID, countName := parseType(rt, m.current)
		if countID >= countName {
			m.current.prefer = true
		}
		regCache[name] = m.current
	}
	return
}

func (m *reflector) Lock(field *Field) bool {
	v, ok := m.lookUpRField(field, true)
	if !ok {
		return false
	}

	if v.Kind() == reflect.Slice {
		v = extractValue(v.Index(field.Value.(int)), true)
		m.values = append(m.values, v.Addr())
	} else {
		v = extractValue(v, true)
		m.values = append(m.values, v.Addr())
	}
	m.index++
	return true
}

func (m *reflector) Unlock() {
	m.values = m.values[:m.index]
	m.index--
}

// find value in message and assign to field
func (m *reflector) GetValue(field *Field) {
	if rField, ok := m.lookUpRField(field, false); ok {
		if rField.Kind() == reflect.Ptr {
			if !rField.IsNil() {
				field.Value = rField.Elem().Interface()
			}
		} else {
			field.Value = rField.Interface()
		}
	}
}

// find slice len in message and assign to field
func (m *reflector) GetLength(field *Field) {
	if rField, ok := m.lookUpRField(field, false); ok {
		field.Value = rField.Len()
	}
}

func (m *reflector) SetLength(field *Field) {
	if rField, ok := m.lookUpRField(field, true); ok {
		length := field.Value.(int)
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
func (m *reflector) GetTemplateID() uint {
	index, ok := m.current.byName["*"]
	if !ok {
		return 0
	}
	return uint(m.values[m.index].Elem().Field(index).Uint())
}

// set template id to message
func (m *reflector) SetTemplateID(tid uint) {
	index, ok := m.current.byName["*"]
	if !ok {
		return
	}

	rField := m.values[m.index].Elem().Field(index)
	m.set(rField, reflect.ValueOf(tid))
}

// set field value to message
func (m *reflector) SetValue(field *Field) {
	if rField, ok := m.lookUpRField(field, true); ok {
		m.set(rField, reflect.ValueOf(field.Value))
	}
}

func (m *reflector) set(field reflect.Value, value reflect.Value) {
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

func (m *reflector) lookUpRField(field *Field, initializeNils bool) (v reflect.Value, ok bool) {
	if field.index == nil {
		m.lookUpIndex(field)
	}
	if field.index == nil {
		return
	}

	v = extractValue(m.values[m.index], initializeNils)
	v = extractValue(v.Field(*field.index), initializeNils)
	ok = true
	return
}

func (m *reflector) lookUpIndex(field *Field) {
	var v int
	var ok bool
	if m.current.prefer {
		if v, ok = m.current.byID[int(field.ID)]; ok {
			field.index = &v
			return
		}
		if v, ok = m.current.byName[field.Name]; ok {
			field.index = &v
			return
		}
	}
	if v, ok = m.current.byName[field.Name]; ok {
		field.index = &v
		return
	}
	if v, ok = m.current.byID[int(field.ID)]; ok {
		field.index = &v
	}
}

func parseType(rt reflect.Type, current *register) (countID, countName int) {
	var (
		field reflect.StructField
		tmp   reflect.Type
		name  string
		id    int
		err   error
		ok    bool
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
			if _, ok = current.byID[id]; ok {
				panic(errors.New("found duplicate struct field"))
			}
			current.byID[id] = i
		} else {
			countName++
			if _, ok = current.byName[name]; ok {
				panic(errors.New("found duplicate struct field"))
			}
			current.byName[name] = i
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

func extractValue(rv reflect.Value, initializeNils bool) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() && initializeNils {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		if !rv.IsNil() {
			return rv.Elem()
		}
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
