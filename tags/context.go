// Copyright 2017, OpenCensus Authors
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

import "golang.org/x/net/context"

// FromContext returns the TagSet stored in the context.
func FromContext(ctx context.Context) *TagSet {
	// The returned TagSet shouldn't be mutated.
	ts := ctx.Value(tagSetCtxKey)
	if ts == nil {
		return newTagSet(0)
	}
	return ts.(*TagSet)
}

// NewContext creates a new context with the given tag set.
// To propagate a tag set to downstream methods and downstream RPCs, add a tag set
// to the current context. NewContext will return a copy of the current context,
// and put the tag set into the returned one.
// If there is already a tag set in the current context, it will be replaced with ts.
func NewContext(ctx context.Context, ts *TagSet) context.Context {
	return context.WithValue(ctx, tagSetCtxKey, ts)
}

type ctxKey struct{}

var tagSetCtxKey = ctxKey{}
