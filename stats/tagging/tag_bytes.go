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

// keyBytes is implementation for keys which values are of type string.
type keyBytes struct {
	name string
}

func (kb *keyBytes) Name() string {
	return kb.name
}

func (kb *keyBytes) Type() keyType {
	return keyTypeBytes
}

func (kb *keyBytes) CreateMutation(v []byte, mb MutationBehavior) *mutationBytes {
	return &mutationBytes{
		tag:      kb.CreateTag(v),
		behavior: mb,
	}
}

func (kb *keyBytes) CreateTag(v []byte) *tagBytes {
	return &tagBytes{
		k: kb,
		v: v,
	}
}

func (kb *keyBytes) String() string {
	return fmt.Sprintf("%T:'%s'", kb, kb.name)
}

// mutationBytes represents a mutation for a tag of type string.
type mutationBytes struct {
	tag      *tagBytes
	behavior MutationBehavior
}

func (mb *mutationBytes) Tag() Tag {
	return mb.tag
}

func (mb *mutationBytes) Behavior() MutationBehavior {
	return mb.behavior
}

// tagBytes is the tuple (key, value) implementation for tags of value type
// string.
type tagBytes struct {
	k KeyBytes
	v []byte
}

func (tb *tagBytes) Key() Key {
	return tb.k
}

func (tb *tagBytes) setKeyFromBytes(fullSig []byte, idx int) (int, error) {
	s, endIdx, err := decodeVarintString(fullSig, idx)
	if err != nil {
		return idx, err
	}
	tb.k, err = DefaultKeyManager().CreateKeyBytes(s)
	if err != nil {
		return idx, err
	}
	return endIdx, nil
}

func (tb *tagBytes) setValueFromBytes(fullSig []byte, idx int) (int, error) {
	b, endIdx, err := decodeVarintBytes(fullSig, idx)
	if err != nil {
		return idx, err
	}
	tb.v = b
	return endIdx, nil
}

func (tb *tagBytes) setValueFromBytesKnownLength(valuesSig []byte, idx int, length int) (int, error) {
	endIdx := idx + length
	if endIdx > len(valuesSig) {
		return idx, fmt.Errorf("unexpected end while tagBytes.setValueFromBytesKnownLength '%x' starting at idx '%v'", valuesSig, idx)
	}
	tb.v = valuesSig[idx:endIdx]
	return endIdx, nil
}

func (tb *tagBytes) encodeValueToBuffer(dst *bytes.Buffer) {
	encodeVarintBytes(dst, tb.v)
}

func (tb *tagBytes) encodeKeyToBuffer(dst *bytes.Buffer) {
	encodeVarintString(dst, tb.k.Name())
}

func (tb *tagBytes) String() string {
	return fmt.Sprintf("{%s, %x}", tb.k.Name(), tb.v)
}
