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

import (
	"bytes"
	"fmt"
	"sort"
)

// Tag is a key value pair that can be propagated on wire.
type Tag struct {
	Key   Key
	Value []byte
}

// ErrKeyNotFound is returned when a key is not found in a tag set.
type ErrKeyNotFound struct {
	Key string
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %q not found", e.Key)
}

// TagSet contains a set of tags. Use NewTagSet to build tag sets.
type TagSet struct {
	m map[Key][]byte
}

// StringValue returns value associated with the specified key
// encoded as a string. If key is not found, it returns ErrValueNotFound.
func (ts *TagSet) StringValue(k Key) (string, error) {
	sk, ok := k.(StringKey)
	if !ok {
		return "", fmt.Errorf("key %q is not a *KeyString", k.Name())
	}
	b, ok := ts.m[k]
	if !ok {
		return "", ErrKeyNotFound{Key: k.Name()}
	}
	return sk.StringValue(b), nil
	// TODO(jbd): Do all key types can provide string values?
}

func (ts *TagSet) String() string {
	var keys []Key
	for k := range ts.m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Name() < keys[j].Name() })

	var buffer bytes.Buffer
	buffer.WriteString("{ ")
	for _, k := range keys {
		buffer.WriteString(fmt.Sprintf("{%v %v}", k.Name(), k.StringValue(ts.m[k])))
	}
	buffer.WriteString(" }")
	return buffer.String()
}

func (ts *TagSet) insert(k Key, v []byte) {
	if _, ok := ts.m[k]; ok {
		return
	}
	ts.m[k] = v
}

func (ts *TagSet) update(k Key, v []byte) {
	if _, ok := ts.m[k]; ok {
		ts.m[k] = v
	}
}

func (ts *TagSet) upsert(k Key, v []byte) {
	ts.m[k] = v
}

func (ts *TagSet) delete(k Key) {
	delete(ts.m, k)
}

func newTagSet(sizeHint int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, sizeHint),
	}
}
