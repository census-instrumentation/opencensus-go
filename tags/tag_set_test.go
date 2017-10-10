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
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestContext(t *testing.T) {
	k1, _ := NewStringKey("k1")
	k2, _ := NewStringKey("k2")

	want := NewTagSet(nil,
		InsertString(k1, "v1"),
		InsertString(k2, "v2"),
	)

	ctx := NewContext(context.Background(), want)
	got := FromContext(ctx)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("TagSet = %#v; want %#v", got, want)
	}
}

func TestNewTagSet(t *testing.T) {
	k1, _ := NewStringKey("k1")
	k2, _ := NewStringKey("k2")
	k3, _ := NewStringKey("k3")
	k4, _ := NewStringKey("k4")
	k5, _ := NewStringKey("k5")

	initial := makeTestTagSet(5)

	tests := []struct {
		name    string
		initial *TagSet
		mods    []Mutator
		want    *TagSet
	}{
		{
			name:    "from empty; insert",
			initial: nil,
			mods: []Mutator{
				InsertString(k5, "v5"),
			},
			want: makeTestTagSet(2, 4, 5),
		},
		{
			name:    "from empty; insert existing",
			initial: nil,
			mods: []Mutator{
				InsertString(k1, "v1"),
			},
			want: makeTestTagSet(1, 2, 4),
		},
		{
			name:    "from empty; update",
			initial: nil,
			mods: []Mutator{
				UpdateString(k1, "v1"),
			},
			want: makeTestTagSet(2, 4),
		},
		{
			name:    "from empty; update unexisting",
			initial: nil,
			mods: []Mutator{
				UpdateString(k5, "v5"),
			},
			want: makeTestTagSet(2, 4),
		},
		{
			name:    "from existing; upsert",
			initial: initial,
			mods: []Mutator{
				UpsertString(k5, "v5"),
			},
			want: makeTestTagSet(2, 4, 5),
		},
		{
			name:    "from existing; delete",
			initial: initial,
			mods: []Mutator{
				Delete(k2),
			},
			want: makeTestTagSet(4, 5),
		},
	}

	for _, tt := range tests {
		mods := []Mutator{
			InsertString(k1, "v1"),
			InsertString(k2, "v2"),
			UpdateString(k3, "v3"),
			UpsertString(k4, "v4"),
			InsertString(k2, "v2"),
			Delete(k1),
		}
		mods = append(mods, tt.mods...)
		got := NewTagSet(tt.initial, mods...)

		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: got %v; want %v", tt.name, got, tt.want)
		}
	}
}

func makeTestTagSet(ids ...int) *TagSet {
	ts := newTagSet(len(ids))
	for _, v := range ids {
		k, _ := NewStringKey(fmt.Sprintf("k%d", v))
		ts.m[k] = []byte(fmt.Sprintf("v%d", v))
	}
	return ts
}
