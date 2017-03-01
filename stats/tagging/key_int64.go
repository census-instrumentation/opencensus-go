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

// keyInt64 is implementation for keys which values are of type int64.
type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) CreateMutation(v int64, mb MutationBehavior) *mutationInt64 {
	mu := &mutationInt64{
		tagInt64: &tagInt64{
			keyInt64: ki,
			v:        v,
		},
		behavior: mb,
	}
	return mu
}

func (ki *keyInt64) CreateTag(i int64) *tagInt64 {
	return &tagInt64{
		keyInt64: ki,
		v:        i,
	}
}

func (ki *keyInt64) writeKeyToBuffer(dst *bytes.Buffer) {
	name := ki.Name()
	dst.WriteByte(byte(keyTypeInt64))
	if len(name) == 0 {
		dst.Write(int32ToBytes(0))
	}
	dst.Write(int32ToBytes(len(name)))
	dst.Write([]byte(name))
}

func (ki *keyInt64) String() string {
	return fmt.Sprintf("%T:'%s'", ki, ki.name)
}
