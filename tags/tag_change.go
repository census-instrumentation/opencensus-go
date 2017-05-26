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

// TagOp defines the types of operations allowed.
type TagOp byte

const (
	// TagOp is not a valid operation. It is here just to detect that a TagOp isn't set.
	TagOpInvalid TagOp = iota

	// TagInsert adds the (key, value) to a set if the set doesn't already
	// contain a tag with the same key. Otherwise it is a no-op.
	TagOpInsert

	// TagOpUpdate replaces the (key, value) in a set if the set contains a
	// (key, value) pair with the same key. Otherwise it is a no-op.
	TagOpUpdate

	// TagOpUpsert adds the (key, value) to a set regardless if the set does
	// contain or doesn't contain a (key, value) pair with the same key.
	TagOpUpsert

	// TagOpDelete deletes the (key, value) from a set if it contain a pair
	// with the same key. Otherwise it is a no-op.
	TagOpDelete
)

// TagChange is the interface for tag changes. It is not expected to have
// multiple types implement it. Its main purpose is to only allow read
// operations on its fields and hide its the write operations.
type TagChange interface {
	Key() Key
	Value() []byte
	Op() TagOp
}

// tagChange implements TagChange
type tagChange struct {
	k  Key
	v  []byte
	op TagOp
}

func (tc *tagChange) Key() Key {
	return tc.k
}

func (tc *tagChange) Value() []byte {
	return tc.v
}

func (tc *tagChange) Op() TagOp {
	return tc.op
}
