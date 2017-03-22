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

// TagsSet is the object holding the tags stored in context.
type TagsSet struct {
	m map[Key]Tag
}

func (ts *TagsSet) String() string {
	var b bytes.Buffer
	for k, v := range ts.m {
		b.WriteString(fmt.Sprintf("{%v:%v} ", k.Name(), v))
	}
	return string(b.Bytes())
}
