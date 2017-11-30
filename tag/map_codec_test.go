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

package tag

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

var keys []Key

func Test_EncodeDecode_Set(t *testing.T) {
	k1, _ := NewKey("k1")
	k2, _ := NewKey("k2")
	k3, _ := NewKey("k3_")
	k4, _ := NewKey("k4")

	type pair struct {
		k Key
		v string
	}

	type testCase struct {
		label string
		pairs []pair
	}

	testCases := []testCase{
		{
			"0",
			[]pair{},
		},
		{
			"1",
			[]pair{
				{k1, "v1"},
			},
		},
		{
			"2",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
			},
		},
		{
			"3",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
				{k3, "v3"},
			},
		},
		{
			"4",
			[]pair{
				{k1, "v1"},
				{k2, "v2"},
				{k3, "v3"},
				{k4, "v4_"},
			},
		},
	}

	for _, tc := range testCases {
		mods := make([]Mutator, len(tc.pairs))
		for i, pair := range tc.pairs {
			mods[i] = Upsert(pair.k, pair.v)
		}
		ts, err := NewMap(context.Background(), mods...)
		if err != nil {
			t.Errorf("%v: NewMap = %v", tc.label, err)
		}
		encoded := Encode(ts)
		decoded, err := Decode(encoded)

		if err != nil {
			t.Errorf("%v: decoding encoded tag map failed: %v", tc.label, err)
		}

		got := make([]pair, 0)
		for k, v := range decoded.m {
			got = append(got, pair{k, string(v)})
		}
		want := tc.pairs

		sort.Slice(got, func(i, j int) bool { return got[i].k.name < got[j].k.name })
		sort.Slice(want, func(i, j int) bool { return got[i].k.name < got[j].k.name })

		if !reflect.DeepEqual(got, tc.pairs) {
			t.Errorf("%v: decoded tag map = %#v; want %#v", tc.label, got, want)
		}
	}
}
