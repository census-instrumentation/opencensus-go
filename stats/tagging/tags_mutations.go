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

// Mutation is the interface that all mutations types need to implements. A
// mutation is a data structure holding a key, a value and a behavior. The
// mutations value types supported are string, int64 and bool.
type Mutation interface {
	Tag() Tag
	Behavior() MutationBehavior
}

// mutationString represents a mutation for a tag of type string.
type mutationString struct {
	*tagString
	behavior MutationBehavior
}

func (ms *mutationString) Tag() Tag {
	return ms.tagString
}

func (ms *mutationString) Behavior() MutationBehavior {
	return ms.behavior
}

// mutationInt64 represents a mutation for a tag of type int64.
type mutationInt64 struct {
	*tagInt64
	behavior MutationBehavior
}

func (mi *mutationInt64) Tag() Tag {
	return mi.tagInt64
}

func (mi *mutationInt64) Behavior() MutationBehavior {
	return mi.behavior
}

// mutationBool represents a mutation for a tag of type bool.
type mutationBool struct {
	*tagBool
	behavior MutationBehavior
}

func (mb *mutationBool) Tag() Tag {
	return mb.tagBool
}

func (mb *mutationBool) Behavior() MutationBehavior {
	return mb.behavior
}

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
