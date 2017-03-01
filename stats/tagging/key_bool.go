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

// keyBool is implementation for keys which values are of type bool.
type keyBool struct {
	name string
}

func (kb *keyBool) Name() string {
	return kb.name
}

func (kb *keyBool) CreateMutation(v bool, mb MutationBehavior) *mutationBool {
	return &mutationBool{
		tagBool: &tagBool{
			keyBool: kb,
			v:       v,
		},
		behavior: mb,
	}
}

func (kb *keyBool) CreateTag(b bool) *tagBool {
	return &tagBool{
		keyBool: kb,
		v:       b,
	}
}

func (kb *keyBool) writeKeyToBuffer(dst *bytes.Buffer) {
	name := kb.Name()
	dst.WriteByte(byte(keyTypeBool))
	if len(name) == 0 {
		dst.Write(int32ToBytes(0))
	}
	dst.Write(int32ToBytes(len(name)))
	dst.Write([]byte(name))
}

func (kb *keyBool) String() string {
	return fmt.Sprintf("%T:'%s'", kb, kb.name)
}
