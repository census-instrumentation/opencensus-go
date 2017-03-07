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

import (
	"bytes"
	"fmt"
)

// TagsSet is the tags set representation in the context.
type TagsSet map[Key]Tag

// ApplyMutation applies a single mutation to the TagsSet
func (ts TagsSet) ApplyMutation(m Mutation) {
	t := m.Tag()
	k := t.Key()
	switch m.Behavior() {
	case BehaviorReplace:
		if _, ok := ts[k]; ok {
			ts[k] = t
		}
	case BehaviorAdd:
		if _, ok := ts[k]; !ok {
			ts[k] = t
		}
	case BehaviorAddOrReplace:
		ts[k] = t
	default:
		panic(fmt.Sprintf("mutation type is %v. This is a bug and should never happen.", m.Behavior()))
	}
}

// ApplyMutations applies multiple mutations to the TagsSet
func (ts TagsSet) ApplyMutations(ms ...Mutation) {
	for _, m := range ms {
		ts.ApplyMutation(m)
	}
}

func (ts TagsSet) String() string {
	var b bytes.Buffer
	for k, v := range ts {
		b.WriteString(fmt.Sprintf("{%v:%v} ", k.Name(), v))
	}
	return string(b.Bytes())
}
