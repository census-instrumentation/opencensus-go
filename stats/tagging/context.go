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

type censusKey struct{}

func FromContext(ctx context.Context) TagsSet {
	ts, ok := ctx.Value(censusKey{}).(TagsSet)
	if !ok {
		ts = make(TagsSet)
	}
	return ts
}

// NewContextWithMutations creates a new census.Context from context and adds
// the tags to it.
func NewContextWithMutations(ctx context.Context, mut ...Mutation) context.Context {
	parentTagsSet, _ := ctx.Value(censusKey{}).(TagsSet)

	return context.WithValue(ctx, censusKey{}, newTagsSet(parentTagsSet, mut...))
}

func newTagsSet(oldTs TagsSet, ms ...Mutation) TagsSet {
	newTs := make(TagsSet)
	for k, t := range oldTs {
		newTs[k] = t
	}

	newTs.ApplyMutations(ms...)
	return newTs
}
