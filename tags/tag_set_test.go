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

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/net/context"
)

func TestContext(t *testing.T) {
	k1, _ := KeyStringByName("k1")
	k2, _ := KeyStringByName("k2")

	tsb := NewTagSetBuilder(nil)
	tsb.InsertString(k1, "v1")
	tsb.InsertString(k2, "v2")
	want := tsb.Build()

	ctx := NewContext(context.Background(), want)
	got := FromContext(ctx)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("TagSet = %#v; want %#v", got, want)
	}
}

func TestTagSetBuilder(t *testing.T) {
	k1, _ := KeyStringByName("k1")
	k2, _ := KeyStringByName("k2")
	k3, _ := KeyStringByName("k3")
	k4, _ := KeyStringByName("k4")
	k5, _ := KeyStringByName("k5")

	initial := makeTestTagSet(5)

	tests := []struct {
		name     string
		initial  *TagSet
		modifier func(tbs *TagSetBuilder) *TagSetBuilder
		want     *TagSet
	}{
		{
			name:    "from empty; insert",
			initial: nil,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.InsertString(k5, "v5")
				return tbs
			},
			want: makeTestTagSet(2, 4, 5),
		},
		{
			name:    "from empty; insert existing",
			initial: nil,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.InsertString(k1, "v1")
				return tbs
			},
			want: makeTestTagSet(1, 2, 4),
		},
		{
			name:    "from empty; update",
			initial: nil,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.UpdateString(k1, "v1")
				return tbs
			},
			want: makeTestTagSet(2, 4),
		},
		{
			name:    "from empty; update unexisting",
			initial: nil,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.UpdateString(k5, "v5")
				return tbs
			},
			want: makeTestTagSet(2, 4),
		},
		{
			name:    "from existing; upsert",
			initial: initial,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.UpsertString(k5, "v5")
				return tbs
			},
			want: makeTestTagSet(2, 4, 5),
		},
		{
			name:    "from existing; delete",
			initial: initial,
			modifier: func(tbs *TagSetBuilder) *TagSetBuilder {
				tbs.Delete(k2)
				return tbs
			},
			want: makeTestTagSet(4, 5),
		},
	}

	for _, tt := range tests {
		tsb := NewTagSetBuilder(tt.initial)
		tsb.InsertString(k1, "v1")
		tsb.InsertString(k2, "v2")
		tsb.UpdateString(k3, "v3")
		tsb.UpsertString(k4, "v4")
		tsb.InsertString(k2, "v2")
		tsb.Delete(k1)
		tsb = tt.modifier(tsb)

		got := tsb.Build()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: got %v; want %v", tt.name, got, tt.want)
		}
	}
}

func makeTestTagSet(ids ...int) *TagSet {
	ts := newTagSet(len(ids))
	for _, v := range ids {
		k, _ := KeyStringByName(fmt.Sprintf("k%d", v))
		ts.m[k] = []byte(fmt.Sprintf("v%d", v))
	}
	return ts
}
