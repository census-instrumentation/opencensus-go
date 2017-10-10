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

package tags_test

import (
	"log"

	"golang.org/x/net/context"

	"github.com/census-instrumentation/opencensus-go/tags"
)

var (
	tagSet *tags.TagSet
	ctx    context.Context
	key    *tags.KeyString
)

func ExampleKeyStringByName() {
	// Get a key to represent user OS.
	key, err := tags.KeyStringByName("/my/namespace/user-os")
	if err != nil {
		log.Fatal(err)
	}
	_ = key // use key
}

func ExampleNewTagSetBuilder() {
	osKey, err := tags.KeyStringByName("/my/namespace/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tags.KeyStringByName("/my/namespace/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tsb := tags.NewTagSetBuilder(nil)
	tsb.InsertString(osKey, "macOS-10.12.5")
	tsb.UpsertString(userIDKey, "cde36753ed")
	tagSet := tsb.Build()

	_ = tagSet // use tag set
}

func ExampleNewTagSetBuilder_replace() {
	oldTagSet := tags.FromContext(ctx)
	tsb := tags.NewTagSetBuilder(oldTagSet)
	tsb.InsertString(key, "odin")
	tsb.UpdateString(key, "thor")
	tsb.UpsertString(key, "loki")
	ctx = tags.NewContext(ctx, tsb.Build())

	_ = ctx // use context
}

func ExampleNewContext() {
	// Propagate the tag set in the current context.
	ctx := tags.NewContext(context.Background(), tagSet)

	_ = ctx // use context
}

func ExampleFromContext() {
	tagSet := tags.FromContext(ctx)

	_ = tagSet // use tag set
}
