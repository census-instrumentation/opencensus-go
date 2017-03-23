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

package tagging

// TagsSet2 is the object holding the tags stored in context.
type TagsSet2 struct {
	keysIDs []int32
	vIndxs  []idxs
	values  *buffer
}

// idxs is a convenience data structure to hold start/end indexes.
type idxs struct {
	s, e int
}

func (ts *TagsSet2) Add(t Tag) bool {
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

func (ts *TagsSet2) Replace(t Tag) bool {
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

func (ts *TagsSet2) AddOrReplace(t Tag) {
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
