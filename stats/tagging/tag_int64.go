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
	"encoding/binary"
	"fmt"
)

// keyInt64 is implementation for keys which values are of type int64.
type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) Type() keyType {
	return keyTypeInt64
}

func (ki *keyInt64) CreateMutation(v int64, mb MutationBehavior) *mutationInt64 {
	return &mutationInt64{
		tag:      ki.CreateTag(v),
		behavior: mb,
	}
}

func (ki *keyInt64) CreateTag(v int64) *tagInt64 {
	return &tagInt64{
		k: ki,
		v: v,
	}
}

func (ki *keyInt64) String() string {
	return fmt.Sprintf("%T:'%s'", ki, ki.name)
}

// mutationInt64 represents a mutation for a tag of type int64.
type mutationInt64 struct {
	tag      *tagInt64
	behavior MutationBehavior
}

func (mi *mutationInt64) Tag() Tag {
	return mi.tag
}

func (mi *mutationInt64) Behavior() MutationBehavior {
	return mi.behavior
}

// tagInt64 is the tuple (key, value) implementation for tags of value type
// int64.
type tagInt64 struct {
	k *keyInt64
	v int64
}

func (ti *tagInt64) Key() Key {
	return ti.k
}

func (ti *tagInt64) setKeyFromBytes(fullSig []byte, idx int) (int, error) {
	s, endIdx, err := decodeVarintString(fullSig, idx)
	if err != nil {
		return idx, err
	}
	ti.k, err = DefaultKeyManager().CreateKeyInt64(s)
	if err != nil {
		return idx, err
	}
	return endIdx, nil
}

func (ti *tagInt64) setValueFromBytes(fullSig []byte, idx int) (int, error) {
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
		return idx, fmt.Errorf("unexpected end while tagInt64.setValueFromBytes '%x' starting at idx '%v'", fullSig, idx)
	}

	ti.v = int64(binary.LittleEndian.Uint64(fullSig[idx:endIdx]))
	return endIdx, nil
}

func (ti *tagInt64) setValueFromBytesKnownLength(valuesSig []byte, idx int, length int) (int, error) {
	endIdx := idx + length
	if endIdx > len(valuesSig) {
		return idx, fmt.Errorf("unexpected end while tagInt64.setValueFromBytesKnownLength '%x' starting at idx '%v'", valuesSig, idx)
	}

	ti.v = int64(binary.LittleEndian.Uint64(valuesSig[idx:endIdx]))
	return endIdx, nil
}

func (ti *tagInt64) encodeValueToBuffer(dst *bytes.Buffer) {
	encodeVarint(dst, 8)

	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(ti.v))
	dst.Write(bytes)
}

func (ti *tagInt64) encodeKeyToBuffer(dst *bytes.Buffer) {
	encodeVarintString(dst, ti.k.name)
}

func (ti *tagInt64) String() string {
	return fmt.Sprintf("{%s, %v}", ti.k.name, ti.v)
}
