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

// TagSetBuilder builds tag sets. It allows
// a TagSet to be be built from multiple pieces
// over time but is immutable once built.
type TagSetBuilder struct {
	ts *TagSet
}

// NewTagSetBuilder starts building a new builder from an existing TagSet.
// If the given TagSet is nil, it starts with an empty set.
func NewTagSetBuilder(ts *TagSet) *TagSetBuilder {
	tb := &TagSetBuilder{}

	if ts == nil {
		tb.ts = newTagSet(0)
		return tb
	}

	tb.ts = newTagSet(len(ts.m))
	for k, b := range ts.m {
		tb.ts.setKeyValue(k, b)
	}
	return tb
}

// InsertString inserts a string value associated with the the key.
// If there is already a value exists with the given key, it doesn't
// update the existing value.
func (tb *TagSetBuilder) InsertString(k *KeyString, s string) *TagSetBuilder {
	if _, ok := tb.ts.m[k]; !ok {
		tb.ts.setKeyValue(k, []byte(s))
	}
	return tb
}

// UpdateString updates a string value associated with the the key.
// If the given key doesn't already exist in the tag set, it does nothing.
func (tb *TagSetBuilder) UpdateString(k *KeyString, s string) *TagSetBuilder {
	if _, ok := tb.ts.m[k]; ok {
		tb.ts.setKeyValue(k, []byte(s))
	}
	return tb
}

// UpsertString updates or insert a string value associated with the key.
func (tb *TagSetBuilder) UpsertString(k *KeyString, s string) *TagSetBuilder {
	tb.ts.setKeyValue(k, []byte(s))
	return tb
}

// Delete deletes the tag associated with the the key.
func (tb *TagSetBuilder) Delete(k Key) *TagSetBuilder {
	delete(tb.ts.m, k)
	return tb
}

// Build returns the built TagSet and clears the builder.
func (tb *TagSetBuilder) Build() *TagSet {
	ts := tb.ts
	tb.ts = nil
	return ts
}
