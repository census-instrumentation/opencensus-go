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
	"fmt"

	"golang.org/x/net/context"
)

// Tag is a key value pair that can be propagated on wire.
type Tag struct {
	K Key
	V []byte
}

// ErrKeyNotFound is returned when a key is not found in a tag set.
type ErrKeyNotFound struct {
	Key string
}

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %q not found", e.Key)
}

// TagSet contains a set of tags. Use TagSetBuilder to build tag sets.
type TagSet struct {
	m map[Key][]byte
}

// StringValue returns value associated with the specified key
// encoded as a string. If key is not found, it returns ErrValueNotFound.
func (ts *TagSet) StringValue(k Key) (string, error) {
	ks, ok := k.(*KeyString)
	if !ok {
		return "", fmt.Errorf("key %q is not a *KeyString", k.Name())
	}
	b, ok := ts.m[k]
	if !ok {
		return "", ErrKeyNotFound{Key: k.Name()}
	}
	return ks.StringValue(b), nil
	// TODO(jbd): Do all key types can provide string values?
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

// FromContext returns the TagSet stored in the context.
func FromContext(ctx context.Context) *TagSet {
	// The returned TagSet shouldn't be mutated.
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
