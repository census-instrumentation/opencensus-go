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

import "encoding/binary"

// TagSetBuilder is the data structure used to build new TagSet. Its purpose
// to ensure TagSet can be built from multiple pieces over time but that it is
// immutable once built.
type TagSetBuilder struct {
	ts *TagSet
}

// StartFromEmpty starts building a new TagSet.
func (tb *TagSetBuilder) StartFromEmpty() *TagSetBuilder {
	tb.ts = newTagSet(0)
	return tb
}

// StartFromTagSet starts building a new TagSet from an existing TagSet.
func (tb *TagSetBuilder) StartFromTagSet(ts *TagSet) *TagSetBuilder {
	if ts == nil {
		tb.ts = newTagSet(0)
		return tb
	}

	tb.ts = newTagSet(len(ts.m))
	for k, b := range ts.m {
		tb.ts.upsertBytes(k, b)
	}
	return tb
}

/*
// StartFromEncoded starts building a new TagSet from an encoded []byte.
func (tb *TagSetBuilder) StartFromEncoded(encoded []byte) error {
	var err error
	tb.ts, err = DecodeFromFullSignatureToTagSet(encoded)
	if err != nil {
		return fmt.Errorf("NewContextWithWireFormat(_) failed. %v", err)
	}
	return nil
}
*/

func (tb *TagSetBuilder) InsertString(k *KeyString, s string) *TagSetBuilder {
	tb.InsertBytes(k, []byte(s))
	return tb
}

func (tb *TagSetBuilder) InsertInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.InsertBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) InsertBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.InsertBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) UpdateString(k *KeyString, s string) *TagSetBuilder {
	tb.UpdateBytes(k, []byte(s))
	return tb
}

func (tb *TagSetBuilder) UpdateInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.UpdateBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) UpdateBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.UpdateBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) UpsertString(k *KeyString, s string) *TagSetBuilder {
	tb.UpsertBytes(k, []byte(s))
	return tb
}

func (tb *TagSetBuilder) UpsertInt64(k *KeyInt64, i int64) *TagSetBuilder {
	v := make([]byte, 8)
	binary.LittleEndian.PutUint64(v, uint64(i))
	tb.UpsertBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) UpsertBool(k *KeyBool, b bool) *TagSetBuilder {
	v := make([]byte, 1)
	if b {
		v[0] = 1
	} else {
		v[1] = 0
	}
	tb.UpsertBytes(k, v)
	return tb
}

func (tb *TagSetBuilder) InsertBytes(k Key, bs []byte) *TagSetBuilder {
	tb.ts.insertBytes(k, bs)
	return tb
}

func (tb *TagSetBuilder) UpdateBytes(k Key, bs []byte) *TagSetBuilder {
	tb.ts.updateBytes(k, bs)
	return tb
}

func (tb *TagSetBuilder) UpsertBytes(k Key, bs []byte) *TagSetBuilder {
	tb.ts.upsertBytes(k, bs)
	return tb
}

func (tb *TagSetBuilder) Delete(k Key) *TagSetBuilder {
	tb.ts.delete(k)
	return tb
}

func (tb *TagSetBuilder) Apply(tcs ...TagChange) *TagSetBuilder {
	for _, tc := range tcs {
		switch tc.Op() {
		case TagOpInsert:
			tb.ts.insertBytes(tc.Key(), tc.Value())
		case TagOpUpdate:
			tb.ts.updateBytes(tc.Key(), tc.Value())
		case TagOpUpsert:
			tb.ts.upsertBytes(tc.Key(), tc.Value())
		case TagOpDelete:
			tb.ts.delete(tc.Key())
		default:
			continue
		}
	}
	return tb
}

// Build returns the built TagSet and clears the builder.
func (tb *TagSetBuilder) Build() *TagSet {
	ts := tb.ts
	tb.ts = nil
	return ts
}
