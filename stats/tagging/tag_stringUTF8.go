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
	"fmt"
)

// keyStringUTF8 is implementation for keys which values are of type string.
type keyStringUTF8 struct {
	name string
	id   int32
}

func (ks *keyStringUTF8) Name() string {
	return ks.name
}

func (ks *keyStringUTF8) ID() int32 {
	return ks.id
}

func (ks *keyStringUTF8) Type() keyType {
	return keyTypeStringUTF8
}

func (ks *keyStringUTF8) CreateMutation(v string, mb MutationBehavior) *mutationStringUTF8 {
	return &mutationStringUTF8{
		tag:      ks.CreateTag(v),
		behavior: mb,
	}
}

func (ks *keyStringUTF8) CreateTag(v string) *tagStringUTF8 {
	return &tagStringUTF8{
		k: ks,
		v: v,
	}
}

func (ks *keyStringUTF8) String() string {
	return fmt.Sprintf("%T:'%s'", ks, ks.name)
}

// mutationStringUTF8 represents a mutation for a tag of type string.
type mutationStringUTF8 struct {
	tag      *tagStringUTF8
	behavior MutationBehavior
}

func (ms *mutationStringUTF8) Tag() Tag {
	return ms.tag
}

func (ms *mutationStringUTF8) Behavior() MutationBehavior {
	return ms.behavior
}

// tagStringUTF8 is the tuple (key, value) implementation for tags of value
// type string.
type tagStringUTF8 struct {
	k KeyStringUTF8
	v string
}

func (ts *tagStringUTF8) Key() Key {
	return ts.k
}

func (ts *tagStringUTF8) Value() interface{} {
	return ts.v
}

func (ts *tagStringUTF8) setKeyFromBytes(fullSig []byte, idx int) (int, error) {
	s, endIdx, err := decodeVarintString(fullSig, idx)
	if err != nil {
		return idx, err
	}
	ts.k, err = DefaultKeyManager().CreateKeyStringUTF8(s)
	if err != nil {
		return idx, err
	}
	return endIdx, nil
}

func (ts *tagStringUTF8) setValueFromBytes(fullSig []byte, idx int) (int, error) {
	s, endIdx, err := decodeVarintString(fullSig, idx)
	if err != nil {
		return idx, err
	}
	ts.v = s
	return endIdx, nil
}

func (ts *tagStringUTF8) setValueFromBytesKnownLength(valuesSig []byte, idx int, length int) (int, error) {
	endIdx := idx + length
	if endIdx > len(valuesSig) {
		return idx, fmt.Errorf("unexpected end while tagStringUTF8.setValueFromBytesKnownLength '%x' starting at idx '%v'", valuesSig, idx)
	}
	ts.v = string(valuesSig[idx:endIdx])
	return endIdx, nil
}

func (ts *tagStringUTF8) encodeValueToBuffer(dst *buffer) {
	dst.writeStringUTF8(ts.v)
}

func (ts *tagStringUTF8) encodeKeyToBuffer(dst *buffer) {
	dst.writeStringUTF8(ts.k.Name())
}

func (ts *tagStringUTF8) String() string {
	return fmt.Sprintf("{%s, %s}", ts.k.Name(), ts.v)
}
