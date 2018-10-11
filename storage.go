// Copyright 2018 Alexander Poltoratskiy. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package fast

type storage map[string]interface{}

func newStorage() storage {
	return make(map[string]interface{})
}

func (s storage) save(key string, value interface{}) {
	s[key] = value
}

func (s storage) load(key string) interface{} {
	if value, ok := s[key]; ok {
		return value
	}
	return nil
}

