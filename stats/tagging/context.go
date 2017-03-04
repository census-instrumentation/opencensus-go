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

import "golang.org/x/net/context"

type ctxKey struct{}

func FromContext(ctx context.Context) TagsSet {
	ts, ok := ctx.Value(ctxKey{}).(TagsSet)
	if !ok {
		ts = nil
	}
	return ts
}

// NewContextWithTagsSet creates a new context containing the new TagsSet.
func NewContextWithTagsSet(ctx context.Context, ts TagsSet) context.Context {
	return context.WithValue(ctx, ctxKey{}, ts)
}

// NewContextWithMutations creates a new context containing a new TagsSet. The
// new TagsSet is constructed from the existing TagsSet to which the mutations
// are applied.
func NewContextWithMutations(ctx context.Context, mut ...Mutation) context.Context {
	parentTagsSet, _ := ctx.Value(ctxKey{}).(TagsSet)

	return context.WithValue(ctx, ctxKey{}, newTagsSet(parentTagsSet, mut...))
}

func newTagsSet(oldTs TagsSet, ms ...Mutation) TagsSet {
	newTs := make(TagsSet)
	for k, t := range oldTs {
		newTs[k] = t
	}

	newTs.ApplyMutations(ms...)
	return newTs
}
