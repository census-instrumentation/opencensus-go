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

// FromContext returns the TagsSet stored in the context. The TagSet shoudln't
// be modified.
func FromContext(ctx context.Context) *TagsSet {
	ts, ok := ctx.Value(ctxKey{}).(*TagsSet)
	if !ok {
		ts = nil
	}
	return ts
}

// FromContextWireFormat returns the TagsSet stored in the context encoded in
// the library custom wire format. This wire format is understood by the same
// libraries in other languages.
func FromContextWireFormat(ctx context.Context) []byte {
	ts := FromContext(ctx)
	encoded := EncodeToFullSignature(ts)
	return encoded
}

// ContextWithNewTagsSet creates a new context from the old one replacing any
// existing TagsSet with the new parameter TagsSet ts.
func ContextWithNewTagsSet(ctx context.Context, ts *TagsSet) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, ts), nil
}

// ContextWithDerivedTagsSet creates a new context from the old one replacing any
// existing TagsSet. The new TagsSet contains the tags already presents in the
// existing TagsSet to which the mutations ms are applied
func ContextWithDerivedTagsSet(ctx context.Context, mut ...Mutation) context.Context {
	builder := &TagsSetBuilder{}

	oldTs, ok := ctx.Value(ctxKey{}).(*TagsSet)
	if !ok {
		builder.StartFromEmpty()
	} else {
		builder.StartFromTagsSet(oldTs)
	}
	builder.AddMutations(mut...)
	return context.WithValue(ctx, ctxKey{}, builder.Build())
}
