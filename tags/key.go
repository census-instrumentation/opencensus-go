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

// Mutator modifies a tag set.
type Mutator interface {
	Mutate(t *TagSet) *TagSet
}

// Key represents a tag key. Keys with the same name will return
// true when compared with the == operator.
type Key interface {
	// Name returns the name of the key.
	Name() string

	// StringValue encodes the given value represented in binary to string.
	StringValue(b []byte) string
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

// StringValue encodes the given value represented in binary to string.
func (k StringKey) StringValue(v []byte) string {
	return string(v)
}

type mutator struct {
	fn func(t *TagSet) *TagSet
}

func (m *mutator) Mutate(t *TagSet) *TagSet {
	return m.fn(t)
}

// InsertString returns a mutator that inserts a
// value assiciated with k. If k already exists in the tag set,
// mutator doesn't update the value.
func InsertString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(ts *TagSet) *TagSet {
			ts.insert(k, []byte(v))
			return ts
		},
	}
}

// UpdateString returns a mutator that updates the
// value of the tag assiciated with k with v. If k doesn't
// exists in the tag set, the mutator doesn't insert the value.
func UpdateString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(ts *TagSet) *TagSet {
			ts.update(k, []byte(v))
			return ts
		},
	}
}

// UpsertString returns a mutator that upserts the
// value of the tag assiciated with k with v. It inserts the
// value if k doesn't exist already. It mutates the value
// if k already exists.
func UpsertString(k StringKey, v string) Mutator {
	return &mutator{
		fn: func(ts *TagSet) *TagSet {
			ts.upsert(k, []byte(v))
			return ts
		},
	}
}

// Delete returns a mutator that deletes
// the value assiciated with k.
func Delete(k Key) Mutator {
	return &mutator{
		fn: func(ts *TagSet) *TagSet {
			ts.delete(k)
			return ts
		},
	}
}

// NewTagSet returns a new tag set originated from orig,
// modified with the provided mutators.
func NewTagSet(orig *TagSet, m ...Mutator) *TagSet {
	var ts *TagSet
	if orig == nil {
		ts = newTagSet(0)
	} else {
		ts = newTagSet(len(orig.m))
		for k, v := range orig.m {
			ts.insert(k, v)
		}
	}
	for _, mod := range m {
		ts = mod.Mutate(ts)
	}
	return ts
}
