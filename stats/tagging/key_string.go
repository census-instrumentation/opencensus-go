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

// keyStringUTF8 is implementation for keys which values are of type string.
type keyStringUTF8 struct {
	name string
}

func (ks *keyStringUTF8) Name() string {
	return ks.name
}

func (ks *keyStringUTF8) Type() keyType {
	return keyTypeStringUTF8
}

func (ks *keyStringUTF8) setFromBytes(bytes []byte, idx int32) (int32, error) {
	length, newIdx, err := readVarint(valuesSig, idx)
	if err != nil {
		return idx, err
	}
	end := newIdx + len
	if end >= len(bytes) {
		return idx, fmt.Errorf("unexpected end while keyStringUTF8.setFromBytes '%v' starting at idx '%v'", bytes, idx)
	}
	k.name = string(bytes[newIdx:end])
	return end, nil
}

func (ks *keyStringUTF8) CreateMutation(v string, mb MutationBehavior) *mutationString {
	return &mutationString{
		tagString: &tagString{
			keyString: ks,
			v:         v,
		},
		behavior: mb,
	}
}

func (ks *keyStringUTF8) CreateTag(s string) *tagString {
	return &tagString{
		keyString: ks,
		v:         s,
	}
}

func (ks *keyStringUTF8) writeKeyToBuffer(dst *bytes.Buffer) {
	name := ks.Name()
	dst.WriteByte(byte(keyTypeStringUTF8))
	if len(name) == 0 {
		dst.Write(int32ToBytes(0))
	}
	dst.Write(int32ToBytes(len(name)))
	dst.Write([]byte(name))
}

func (ks *keyStringUTF8) writeKeyToBuffer(dst *bytes.Buffer) {
	name := ks.Name()
	dst.WriteByte(byte(keyTypeStringUTF8))
	if len(name) == 0 {
		dst.Write(int32ToBytes(0))
	}
	dst.Write(int32ToBytes(len(name)))
	dst.Write([]byte(name))
}

func (ks *keyStringUTF8) String() string {
	return fmt.Sprintf("%T:'%s'", ks, ks.name)
}
