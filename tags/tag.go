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

package tags

// Change is the interface that all changes types need to implements. A
// change is a data structure holding a key, a value and an operation. The
// changes value types supported are string, int64 and bool.
type Change interface {
	Tag() Tag
	TagOp() TagOp
}

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	Key() Key
	Value() interface{}
	setKeyFromBytes(fullSig []byte, idx int) (newIdx int, err error)
	setValueFromBytes(fullSig []byte, idx int) (newIdx int, err error)
	setValueFromBytesKnownLength(valuesSig []byte, idx int, len int) (newIdx int, err error)
	encodeValueToBuffer(dst *buffer)
	encodeKeyToBuffer(dst *buffer)
}

type tagSliceByName []Tag

func (ts tagSliceByName) Len() int { return len(ts) }

func (ts tagSliceByName) Swap(i, j int) { ts[i], ts[j] = ts[j], ts[i] }

func (ts tagSliceByName) Less(i, j int) bool { return ts[i].Key().Name() < ts[j].Key().Name() }

// TagOp defines the types of operations allowed.
type TagOp byte

const (
	// TagOp is not a valid operation. It is here just to detect that a TagOp isn't set.
	TagOpInvalid TagOp = iota

	// TagInsert adds the (key, value) to a set if the set doesn't already
	// contain a tag with the same key. Otherwise it is a no-op.
	TagOpInsert

	// TagOpSet adds the (key, value) to a set regardless if the set doesn't
	// contains a (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpSet

	// TagOpReplace replaces the (key, value) in a set if the set contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpReplace
)
