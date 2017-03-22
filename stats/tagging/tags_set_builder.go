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

import "fmt"

// TagsSetBuilder is the data structure used to build new TagsSet. Its purpose
// to ensure TagsSet can be built from multiple pieces over time but that it is
// immutable once built.
type TagsSetBuilder struct {
	ts *TagsSet
}

// StartFromEmpty starts building a new TagsSet.
func (tb *TagsSetBuilder) StartFromEmpty() *TagsSetBuilder {
	tb.ts = &TagsSet{
		m: make(map[Key]Tag),
	}
	return tb
}

// StartFromTagsSet starts building a new TagsSet from an existing TagSet.
func (tb *TagsSetBuilder) StartFromTagsSet(ts *TagsSet) {
	var m map[Key]Tag
	if len(ts.m) == 0 {
		m = make(map[Key]Tag)
	} else {
		m = make(map[Key]Tag, len(ts.m))
		for k, t := range ts.m {
			m[k] = t
		}
	}
	tb.ts = &TagsSet{
		m: m,
	}
}

// StartFromEncoded starts building a new TagsSet from an encoded []byte.
func (tb *TagsSetBuilder) StartFromEncoded(encoded []byte) error {
	var err error
	tb.ts, err = DecodeFromFullSignatureToTagsSet(encoded)
	if err != nil {
		return fmt.Errorf("NewContextWithWireFormat(_) failed. %v", err)
	}
	return nil
}

// AddMutations applies multiple mutations to the TagsSet being built.
func (tb *TagsSetBuilder) AddMutations(muts ...Mutation) {
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

// AddOrReplaceTag adds a Tag to the TagsSet being built. If the TagsSet
// already contains a Tag with the same key it is replaced by the new Tag.
func (tb *TagsSetBuilder) AddOrReplaceTag(t Tag) {
	tb.ts.m[t.Key()] = t
}

// Build returns the built TagsSet and clears the builder.
func (tb *TagsSetBuilder) Build() *TagsSet {
	return tb.ts
}
