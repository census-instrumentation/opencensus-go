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
	"errors"
	"fmt"
	"sort"

	"golang.org/x/net/context"
)

// Tag is a key value pair that can be propagated on wire.
type Tag struct {
	K Key
	V []byte
}

// TagSet contains a set of tags. Use TagSetBuilder to build tag sets.
type TagSet struct {
	m map[Key][]byte
}

// ErrValueNotFound is returned when value is not found in a tag set.
var ErrValueNotFound = errors.New("no value found")

// ValueAsString returns value associated with the specified key
// encoded as a string. If key is not found, it returns ErrValueNotFound.
func (ts *TagSet) ValueAsString(k Key) (string, error) {
	if _, ok := k.(*KeyString); !ok {
		return "", errors.New("key is not a *KeyString")
	}
	b, ok := ts.m[k]
	if !ok {
		return "", ErrValueNotFound
	}
	return k.ValueAsString(b), nil
}

func (ts *TagSet) setKeyValue(k Key, v []byte) {
	ts.m[k] = v
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

// FromContext returns the TagSet stored in the context. The returned TagSet
// shouldn't be mutated.
func FromContext(ctx context.Context) *TagSet {
	ts := ctx.Value(tagSetCtxKey)
	if ts == nil {
		return newTagSet(0)
	}
	return ts.(*TagSet)
}

// TODO(jbd): It says "The returned TagSet shouldn't be mutated.",
// but tag set cannot be mutated. Remove the comment.

// NewContext creates a new context from the old one replacing any existing
// TagSet with the new parameter TagSet ts.
func NewContext(ctx context.Context, ts *TagSet) context.Context {
	return context.WithValue(ctx, tagSetCtxKey, ts)
}

func newTagSet(sizeHint int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, sizeHint),
	}
}

type ctxKey struct{}

var tagSetCtxKey = ctxKey{}
