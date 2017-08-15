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

// Package stats defines the stats collection API and its native Go
// implementation.

package tags

import "context"

// CreateKeyString creates/retrieves the *KeyString identified by name.
var CreateKeyString func(name string) (*KeyString, error)

// CreateChangeString creates a command to change a tag a.k.a (Key,Value) pair.
// This change command is expected to be passed as argument to
// ContextWithDerivedTagSet to modify a tags set context.
var CreateChangeString func(k *KeyString, s string, op TagOp) TagChange

// ContextWithChanges creates a new context where the census tags are replaced
// with a new set of tags. The original context set of tags is unchanged. The
// new set of tags is the result of the old set of tags to which the Tagchanges
// are applied.
// NOT supported in v0.1 and is subject to change
var ContextWithChanges func(ctx context.Context, tcs ...TagChange) context.Context

// Extra functionality not supported/needed for V1.
/*var	CreateKeyInt64 func(name string) (*KeyInt64, error)
var	CreateKeyBool func(name string) (*KeyBool, error)
var	CreateKeyInt64 func(k *KeyInt64, i int64, op TagOp) TagChange
var	CreateChangeBool func(k *KeyBool, b bool, op TagOp) TagChange
*/

func init() {
	CreateKeyString = DefaultKeyManager().CreateKeyString
	CreateChangeString = func(k *KeyString, s string, op TagOp) TagChange {
		return k.CreateChange(s, op)
	}
	ContextWithChanges = func(ctx context.Context, tcs ...TagChange) context.Context {
		return ContextWithDerivedTagSet(ctx, tcs...)
	}
}
