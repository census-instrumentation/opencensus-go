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
	tagMap *tags.Map
	ctx    context.Context
	key    tags.StringKey
)

func ExampleNewStringKey() {
	// Get a key to represent user OS.
	key, err := tags.NewStringKey("/my/namespace/user-os")
	if err != nil {
		log.Fatal(err)
	}
	_ = key // use key
}

func ExampleNewMap() {
	osKey, err := tags.NewStringKey("/my/namespace/user-os")
	if err != nil {
		log.Fatal(err)
	}
	userIDKey, err := tags.NewStringKey("/my/namespace/user-id")
	if err != nil {
		log.Fatal(err)
	}

	tagMap := tags.NewMap(nil,
		tags.InsertString(osKey, "macOS-10.12.5"),
		tags.UpsertString(userIDKey, "cde36753ed"),
	)
	_ = tagMap // use the tag map
}

func ExampleNewMap_replace() {
	oldTagMap := tags.FromContext(ctx)
	tagMap := tags.NewMap(oldTagMap,
		tags.InsertString(key, "macOS-10.12.5"),
		tags.UpsertString(key, "macOS-10.12.7"),
	)
	ctx = tags.NewContext(ctx, tagMap)

	_ = ctx // use context
}

func ExampleNewContext() {
	// Propagate the tag map in the current context.
	ctx := tags.NewContext(context.Background(), tagMap)

	_ = ctx // use context
}

func ExampleFromContext() {
	tagMap := tags.FromContext(ctx)

	_ = tagMap // use the tag map
}
