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

// keyBool is implementation for keys which values are of type bool.
type keyBool struct {
	name string
	id   int32
}

func (kb *keyBool) Name() string {
	return kb.name
}

func (kb *keyBool) ID() int32 {
	return kb.id
}

func (kb *keyBool) Type() keyType {
	return keyTypeBool
}

func (kb *keyBool) CreateMutation(v bool, mb MutationBehavior) *mutationBool {
	return &mutationBool{
		tag:      kb.CreateTag(v),
		behavior: mb,
	}
}

func (kb *keyBool) CreateTag(v bool) *tagBool {
	return &tagBool{
		k: kb,
		v: v,
	}
}

func (kb *keyBool) String() string {
	return fmt.Sprintf("%T:'%s'", kb, kb.name)
}

// mutationBool represents a mutation for a tag of type bool.
type mutationBool struct {
	tag      *tagBool
	behavior MutationBehavior
}

func (mb *mutationBool) Tag() Tag {
	return mb.tag
}

func (mb *mutationBool) Behavior() MutationBehavior {
	return mb.behavior
}

// tagBool is the tuple (key, value) implementation for tags of value type
// bool.
type tagBool struct {
	k KeyBool
	v bool
}

func (tb *tagBool) Key() Key {
	return tb.k
}

func (tb *tagBool) Value() interface{} {
	return tb.v
}

func (tb *tagBool) setKeyFromBytes(fullSig []byte, idx int) (int, error) {
	s, endIdx, err := decodeVarintString(fullSig, idx)
	if err != nil {
		return idx, err
	}
	tb.k, err = DefaultKeyManager().CreateKeyBool(s)
	if err != nil {
		return idx, err
	}
	return endIdx, nil
}

func (tb *tagBool) setValueFromBytes(fullSig []byte, idx int) (int, error) {
	var (
		length int
		err    error
	)

	length, idx, err = decodeVarint(fullSig, idx)
	if err != nil {
		return idx, err
	}

	endIdx := idx + length
	if endIdx > len(fullSig) {
		return idx, fmt.Errorf("unexpected end while tagBool.setValueFromBytes '%x' starting at idx '%v'", fullSig, idx)
	}

	if fullSig[idx] == 0 {
		tb.v = false
	} else {
		tb.v = true
	}
	return endIdx, nil
}

func (tb *tagBool) setValueFromBytesKnownLength(valuesSig []byte, idx int, length int) (int, error) {
	endIdx := idx + length
	if endIdx > len(valuesSig) {
		return idx, fmt.Errorf("unexpected end while tagBool.setValueFromBytesKnownLength '%x' starting at idx '%v'", valuesSig, idx)
	}

	if valuesSig[idx] == 0 {
		tb.v = false
	} else {
		tb.v = true
	}
	return endIdx, nil
}

func (tb *tagBool) encodeValueToBuffer(dst *buffer) {
	if tb.v {
		dst.writeValueTrue()
		return
	}
	dst.writeValueFalse()
}

func (tb *tagBool) encodeKeyToBuffer(dst *buffer) {
	dst.writeStringUTF8(tb.k.Name())
}

func (tb *tagBool) String() string {
	return fmt.Sprintf("{%s, %v}", tb.k.Name(), tb.v)
}
