// Copyright 2017 Google Inc.
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
)

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

func newTagSet(size int) *TagSet {
	return &TagSet{
		m: make(map[Key][]byte, size),
	}
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
