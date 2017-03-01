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

import "bytes"

// Key is the interface for all key types.
type Key interface {
	Name() string
	Type() keyType
}

// Mutation is the interface that all mutations types need to implements. A
// mutation is a data structure holding a key, a value and a behavior. The
// mutations value types supported are string, int64 and bool.
type Mutation interface {
	Tag() Tag
	Behavior() MutationBehavior
}

// Tag is the tuple (key, value) interface for all tag types.
type Tag interface {
	Key() Key
	setKeyFromBytes(fullSig []byte, idx int) (newIdx int, err error)
	setValueFromBytes(fullSig []byte, idx int) (newIdx int, err error)
	setValueFromBytesKnownLength(valuesSig []byte, idx int, len int) (newIdx int, err error)
	encodeValueToBuffer(dst *bytes.Buffer)
	encodeKeyToBuffer(dst *bytes.Buffer)
}

type tagSliceByName []Tag

func (ts tagSliceByName) Len() int { return len(ts) }

func (ts tagSliceByName) Swap(i, j int) { ts[i], ts[j] = ts[j], ts[i] }

func (ts tagSliceByName) Less(i, j int) bool { return ts[i].Key().Name() < ts[j].Key().Name() }

// KeyType defines the types of keys allowed.
type keyType byte

const (
	keyTypeStringUTF8 keyType = iota
	keyTypeInt64
	keyTypeBool
	keyTypeBytes
)

// MutationBehavior defines the types of mutations allowed.
type MutationBehavior byte

const (
	// BehaviorUnknown is not a valid behavior. It is here just to detect that
	// a MutationBehavior isn't set.
	BehaviorUnknown MutationBehavior = iota

	// BehaviorReplace replaces the (key, value) in a set if the set already
	// contains a (key, value) pair with the same key. Otherwise it is a no-op.
	BehaviorReplace

	// BehaviorAdd adds the (key, value) in a set if the set doesn't contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	BehaviorAdd

	// BehaviorAddOrReplace replaces the (key, value) in a set if the set
	// contains a (key, value) pair with the same key. Otherwise it adds the
	// (key, value) to the set.
	BehaviorAddOrReplace
)
