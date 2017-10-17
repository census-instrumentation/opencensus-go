// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package tags

// Mutator modifies a tag map.
type Mutator interface {
	Mutate(t *Map) *Map
}

// Key represents a tag key. Keys with the same name will return
// true when compared with the == operator.
type Key interface {
	// Name returns the name of the key.
	Name() string

	// ValueToString encodes the given value represented in binary to string.
	ValueToString(b []byte) string
}

// StringKey is a Key and represents string keys.
type StringKey struct {
	id   uint16
	name string
}

// NewStringKey creates or retrieves a string key identified by name.
// Calling NewStringKey consequently with the same name returns the same key.
func NewStringKey(name string) (StringKey, error) {
	return km.newStringKey(name)
}

// Name returns the name of the key.
func (k StringKey) Name() string {
	return k.name
}

// ValueToString represents the []byte as string.
func (k StringKey) ValueToString(v []byte) string {
	return string(v)
}

type mutator struct {
	fn func(t *Map) *Map
}

func (m *mutator) Mutate(t *Map) *Map {
	return m.fn(t)
}

// InsertString returns a mutator that inserts a
// value assiciated with k. If k already exists in the tag map,
// mutator doesn't update the value.
func InsertString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.insert(k, []byte(v))
			return m
		},
	}
}

// UpdateString returns a mutator that updates the
// value of the tag assiciated with k with v. If k doesn't
// exists in the tag map, the mutator doesn't insert the value.
func UpdateString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.update(k, []byte(v))
			return m
		},
	}
}

// UpsertString returns a mutator that upserts the
// value of the tag assiciated with k with v. It inserts the
// value if k doesn't exist already. It mutates the value
// if k already exists.
func UpsertString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.upsert(k, []byte(v))
			return m
		},
	}
}

// Delete returns a mutator that deletes
// the value assiciated with k.
func Delete(k Key) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.delete(k)
			return m
		},
	}
}

// NewMap returns a new tag map originated from orig,
// modified with the provided mutators.
func NewMap(orig *Map, mutator ...Mutator) *Map {
	var m *Map
	if orig == nil {
		m = newMap(0)
	} else {
		m = newMap(len(orig.m))
		for k, v := range orig.m {
			m.insert(k, v)
		}
	}
	for _, mod := range mutator {
		m = mod.Mutate(m)
	}
	return m
}
