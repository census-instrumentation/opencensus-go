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

	"golang.org/x/net/context"
)

// Tag is the tuple (key, value) used only when extracting []Tag from a TagSet.
type Tag struct {
	K Key
	V []byte
}

// TagSet is the object holding the tags stored in context. It is not meant to
// be created manually by code outside the library. It should only be created
// using the TagSetBuilder class.
type TagSet struct {
	m map[Key][]byte
}

// ValueAsString returns the string associated with a specified key.
func (ts *TagSet) ValueAsString(k Key) (string, error) {
	if _, ok := k.(*KeyString); !ok {
		return "", fmt.Errorf("values of key '%v' are not of type string", k.Name())
	}

	b, ok := ts.m[k]
	if !ok {
		return "", fmt.Errorf("no value assigned to tag key '%v'", k.Name())
	}
	return string(b), nil
}

func newTagSet(sizeHint int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, sizeHint),
	}
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
		buffer.WriteString(fmt.Sprintf("{%v %v}", k.Name(), k.ValueAsString(ts.m[k])))
	}
	buffer.WriteString(" }")
	return buffer.String()
}

func (ts *TagSet) insertBytes(k Key, b []byte) bool {
	if _, ok := ts.m[k]; ok {
		return false
	}
	ts.m[k] = b
	return true
}

func (ts *TagSet) updateBytes(k Key, b []byte) bool {
	if _, ok := ts.m[k]; !ok {
		return false
	}
	ts.m[k] = b
	return true
}

func (ts *TagSet) upsertBytes(k Key, b []byte) {
	ts.m[k] = b
}

func (ts *TagSet) delete(k Key) {
	delete(ts.m, k)
}

type ctxKey struct{}

// FromContext returns the TagSet stored in the context. The TagSet shoudln't
// be modified.
func FromContext(ctx context.Context) *TagSet {
	ts, ok := ctx.Value(ctxKey{}).(*TagSet)
	if !ok {
		ts = newTagSet(0)
	}
	return ts
}

// NewContext creates a new context from the old one replacing any existing
// TagSet with the new parameter TagSet ts.
func NewContext(ctx context.Context, ts *TagSet) context.Context {
	return context.WithValue(ctx, ctxKey{}, ts)
}
