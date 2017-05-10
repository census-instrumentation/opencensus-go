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

import "fmt"

// TagSetBuilder is the data structure used to build new TagSet. Its purpose
// to ensure TagSet can be built from multiple pieces over time but that it is
// immutable once built.
type TagSetBuilder struct {
	ts *TagSet
}

// StartFromEmpty starts building a new TagSet.
func (tb *TagSetBuilder) StartFromEmpty() *TagSetBuilder {
	tb.ts = &TagSet{
		m: make(map[Key]Tag),
	}
	return tb
}

// StartFromTags starts building a new TagSet from a slice of tags.
func (tb *TagSetBuilder) StartFromTags(tags []Tag) {
	m := make(map[Key]Tag, len(tags))
	for _, t := range tags {
		m[t.Key()] = t
	}

	tb.ts = &TagSet{
		m: m,
	}
}

// StartFromTagSet starts building a new TagSet from an existing TagSet.
func (tb *TagSetBuilder) StartFromTagSet(ts *TagSet) {
	var m map[Key]Tag
	if len(ts.m) == 0 {
		m = make(map[Key]Tag)
	} else {
		m = make(map[Key]Tag, len(ts.m))
		for k, t := range ts.m {
			m[k] = t
		}
	}
	tb.ts = &TagSet{
		m: m,
	}
}

// StartFromEncoded starts building a new TagSet from an encoded []byte.
func (tb *TagSetBuilder) StartFromEncoded(encoded []byte) error {
	var err error
	tb.ts, err = DecodeFromFullSignatureToTagSet(encoded)
	if err != nil {
		return fmt.Errorf("NewContextWithWireFormat(_) failed. %v", err)
	}
	return nil
}

// AddMutations applies multiple mutations to the TagSet being built.
func (tb *TagSetBuilder) AddMutations(muts ...Mutation) {
	ts := tb.ts
	for _, m := range muts {
		t := m.Tag()
		k := t.Key()
		switch m.Behavior() {
		case BehaviorReplace:
			if _, ok := ts.m[k]; ok {
				ts.m[k] = t
			}
		case BehaviorAdd:
			if _, ok := ts.m[k]; !ok {
				ts.m[k] = t
			}
		case BehaviorAddOrReplace:
			ts.m[k] = t
		default:
			panic(fmt.Sprintf("mutation type is %v. This is a bug and should never happen.", m.Behavior()))
		}
	}
}

// AddOrReplaceTag adds a Tag to the TagSet being built. If the TagSet
// already contains a Tag with the same key it is replaced by the new Tag.
func (tb *TagSetBuilder) AddOrReplaceTag(t Tag) {
	tb.ts.m[t.Key()] = t
}

// Build returns the built TagSet and clears the builder.
func (tb *TagSetBuilder) Build() *TagSet {
	ret := tb.ts
	tb.ts = nil
	return ret
}
