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

// TagMap is the object holding the tags stored in context.
type TagMap struct {
	keysIDs []int32
	vIndxs  []idxs
	values  *buffer
}

func (tm *TagMap) Apply(cs ...Change) *TagMap {
	n := tm.copy()
	for _, c := range cs {
		n.apply(c)
	}
	return n
}

func (tm *TagMap) apply(c Change) {
	return tm
}

func (tm *TagMap) copy() *TagMap {
	return tm
}

// idxs is a convenience data structure to hold start/end indexes.
type idxs struct {
	s, e int
}

func (ts *TagMap) Add(t Tag) bool {
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

func (ts *TagMap) Replace(t Tag) bool {
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

func (ts *TagMap) AddOrReplace(t Tag) {
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
func (ts *TagMap) TagKeyExists(key TagKey) bool {
}

////////////////////////////////////////////
func (ts *TagMap) TagValueString(key TagKeyString) string {
}

func (ts *TagMap) TagValueBool(key TagKeyBool) bool {
}

func (ts *TagMap) TagValueInt(key TagKeyInt64) int64 {
}

func (ts *TagMap) NewTagSet(tcs []TagChange) (*TagSet2, error) {

}

func (ts *TagMap) insert(keyIdx int, b []byte) {}

func (ts *TagMap) set(keyIdx int, b []byte) {}

func (ts *TagMap) update(keyIdx int, b []byte) {}

func (ts *TagMap) clear(keyIdx int, b []byte) {}

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
