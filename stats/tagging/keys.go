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

// Key is the interface for all key types.
type Key interface {
	Name() string
}

// keyString is implementation for keys which values are of type string.
type keyString struct {
	name string
}

func (ks *keyString) Name() string {
	return ks.name
}

func (ks *keyString) CreateMutation(v string, mb MutationBehavior) *mutationString {
	return &mutationString{
		tagString: &tagString{
			keyString: ks,
			v:         v,
		},
		behavior: mb,
	}
}

func (ks *keyString) createTag(s string) *tagString {
	return &tagString{
		keyString: ks,
		v:         s,
	}
}

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

func (kb *keyBool) createTag(b bool) *tagBool {
	return &tagBool{
		keyBool: kb,
		v:       b,
	}
}

// keyInt64 is implementation for keys which values are of type int64.
type keyInt64 struct {
	name string
}

func (ki *keyInt64) Name() string {
	return ki.name
}

func (ki *keyInt64) CreateMutation(v int64, mb MutationBehavior) *mutationInt64 {
	return &mutationInt64{
		tagInt64: &tagInt64{
			keyInt64: ki,
			v:        v,
		},
		behavior: mb,
	}
}

func (ki *keyInt64) createTag(i int64) *tagInt64 {
	return &tagInt64{
		keyInt64: ki,
		v:        i,
	}
}

// KeyType defines the types of keys allowed.
type keyType byte

const (
	// keyTypeUnknown is not a valid KeyType. It is here just to detect that a
	// keyType isn't set.
	keyTypeUnknown keyType = iota
	keyTypeString
	keyTypeBool
	keyTypeInt64
)
