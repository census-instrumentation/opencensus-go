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

// TagSet2 is the object holding the tags stored in context.
type TagSet2 struct {
	keysIDs []int32
	vIndxs  []idxs
	values  *buffer
}

// idxs is a convenience data structure to hold start/end indexes.
type idxs struct {
	s, e int
}

func (ts *TagSet2) Add(t Tag) bool {
	for _, id := range ts.keysIDs {
		if id == t.Key().ID() {
			return false
		}
	}

	s := ts.values.writeIdx
	t.encodeValueToBuffer(ts.values)
	e := ts.values.writeIdx
	ts.keysIDs = append(ts.keysIDs, t.Key().ID())
	ts.vIndxs = append(ts.vIndxs, idxs{s, e})
	return true
}

func (ts *TagSet2) Replace(t Tag) bool {
	for i, id := range ts.keysIDs {
		if id == t.Key().ID() {
			s := ts.values.writeIdx
			t.encodeValueToBuffer(ts.values)
			e := ts.values.writeIdx
			ts.vIndxs[i] = idxs{s, e}
			return true
		}
	}
	return false
}

func (ts *TagSet2) AddOrReplace(t Tag) {
	for i, id := range ts.keysIDs {
		if id == t.Key().ID() {
			s := ts.values.writeIdx
			t.encodeValueToBuffer(ts.values)
			e := ts.values.writeIdx
			ts.vIndxs[i] = idxs{s, e}
		}
	}

	s := ts.values.writeIdx
	t.encodeValueToBuffer(ts.values)
	e := ts.values.writeIdx
	ts.keysIDs = append(ts.keysIDs, t.Key().ID())
	ts.vIndxs = append(ts.vIndxs, idxs{s, e})
}

////////////////////////////////////////////
////////////////////////////////////////////
////////////////////////////////////////////
////////////////////////////////////////////
func (ts *TagSet2) TagKeyExists(key TagKey) bool {
}

////////////////////////////////////////////
func (ts *TagSet2) TagValueString(key TagKeyString) string {
}

func (ts *TagSet2) TagValueBool(key TagKeyBool) bool {
}

func (ts *TagSet2) TagValueInt(key TagKeyInt64) int64 {
}

func (ts *TagSet2) NewTagSet(tcs []TagChange) (*TagSet2, error) {

}

func (ts *TagSet2) insert(keyIdx int, b []byte) {}

func (ts *TagSet2) set(keyIdx int, b []byte) {}

func (ts *TagSet2) update(keyIdx int, b []byte) {}

func (ts *TagSet2) clear(keyIdx int, b []byte) {}

////////////////////////////////////////////
////////////////////////////////////////////
type TagSetBuilder struct{}

////////////////////////////////////////////
func (tsb *TagSetBuilder) insert(tk TagKey, b []byte) bool {
	return true
}

func (tsb *TagSetBuilder) InsertString(tk TagKeyString, s string) bool {
	return true
}

func (tsb *TagSetBuilder) InsertBool(tk TagKeyBool, b bool) bool {
	return true
}

func (tsb *TagSetBuilder) InsertInt64(tk TagKeyInt64, i int64) bool {
	return true
}

////////////////////////////////////////////
func (tsb *TagSetBuilder) set(tk TagKey, b []byte) {
}

func (tsb *TagSetBuilder) SetString(tk TagKeyString, s string) {
}

func (tsb *TagSetBuilder) SetBool(tk TagKeyBool, b bool) {
}

func (tsb *TagSetBuilder) SetInt64(tk TagKeyInt64, i int64) {
}

////////////////////////////////////////////
func (tsb *TagSetBuilder) update(tk TagKey, b []byte) bool {
	return true
}

func (tsb *TagSetBuilder) UpdateString(tk TagKeyString, s string) bool {
	return true
}

func (tsb *TagSetBuilder) UpdateBool(tk TagKeyBool, b bool) bool {
	return true
}

func (tsb *TagSetBuilder) UpdateInt64(tk TagKeyInt64, i int64) bool {
	return true
}

////////////////////////////////////////////
func (tsb *TagSetBuilder) Clear(tk TagKey) bool {
	return true
}

////////////////////////////////////////////
func (tsb *TagSetBuilder) FromEmpty() TagSet2 {
}

func (tsb *TagSetBuilder) FromTagSet(old *TagSet2) {
}

func (tsb *TagSetBuilder) Build() TagSet2 {
}

////////////////////////////////////////////
////////////////////////////////////////////
type TagKey interface {
	id() int
}

////////////////////////////////////////////
////////////////////////////////////////////
type TagChange struct {
	k  TagKey
	v  []byte
	op TagOp
}

func (ks *TagKeyString) tagChange(s string, op TagOp) TagChange {}

func (ks *TagKeyBool) tagChange(b bool, op TagOp) TagChange {}

func (ks *TagKeyBool) tagChange(i int64, op TagOp) TagChange {}
