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

package tag

// Mutator modifies a tag map.
type Mutator interface {
	Mutate(t *Map) *Map
}

// Key represents a tag key. Keys with the same name will return
// true when compared with the == operator.
type Key struct {
	id   uint16
	name string
}

// NewKey creates or retrieves a string key identified by name.
// Calling NewKey consequently with the same name returns the same key.
func NewKey(name string) (Key, error) {
	return km.newStringKey(name)
}

// Name returns the name of the key.
func (k Key) Name() string {
	return k.name
}

type mutator struct {
	fn func(t *Map) *Map
}

func (m *mutator) Mutate(t *Map) *Map {
	return m.fn(t)
}

// Insert returns a mutator that inserts a
// value associated with k. If k already exists in the tag map,
// mutator doesn't update the value.
func Insert(k Key, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.insert(k, v)
			return m
		},
	}
}

// Update returns a mutator that updates the
// value of the tag associated with k with v. If k doesn't
// exists in the tag map, the mutator doesn't insert the value.
func Update(k Key, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.update(k, v)
			return m
		},
	}
}

// Upsert returns a mutator that upserts the
// value of the tag associated with k with v. It inserts the
// value if k doesn't exist already. It mutates the value
// if k already exists.
func Upsert(k Key, v string) Mutator {
	return &mutator{
		fn: func(m *Map) *Map {
			m.upsert(k, v)
			return m
		},
	}
}

// Delete returns a mutator that deletes
// the value associated with k.
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
