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

import "golang.org/x/net/context"

type ctxKey struct{}

// FromContext returns the TagSet stored in the context. The TagSet shoudln't
// be modified.
func FromContext(ctx context.Context) *TagSet {
	ts, ok := ctx.Value(ctxKey{}).(*TagSet)
	if !ok {
		ts = newTagSet(0)
	}
	return ts
}

// ContextWithNewTagSet creates a new context from the old one replacing any
// existing TagSet with the new parameter TagSet ts.
func ContextWithNewTagSet(ctx context.Context, ts *TagSet) (context.Context, error) {
	return context.WithValue(ctx, ctxKey{}, ts), nil
}

// ContextWithDerivedTagSet creates a new context from the old one replacing any
// existing TagSet. The new TagSet contains the tags already presents in the
// existing TagSet to which the mutations ms are applied
func ContextWithDerivedTagSet(ctx context.Context, tcs ...TagChange) context.Context {
	builder := &TagSetBuilder{}

	oldTs, ok := ctx.Value(ctxKey{}).(*TagSet)
	if !ok {
		builder.StartFromEmpty()
	} else {
		builder.StartFromTagSet(oldTs)
	}

	builder.Apply(tcs...)
	return context.WithValue(ctx, ctxKey{}, builder.Build())
}
