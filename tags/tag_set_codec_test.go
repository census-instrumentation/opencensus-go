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
	"reflect"
	"sort"
	"testing"
)

func Test_EncodeDecode_TagSet(t *testing.T) {
	k1, _ := NewStringKey("k1")
	k2, _ := NewStringKey("k2")
	k3, _ := NewStringKey("k3 is very weird <>.,?/'\";:`~!@#$%^&*()_-+={[}]|\\")
	k4, _ := NewStringKey("k4")

	type pair struct {
		k StringKey
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
				{k4, "v4 is very weird <>.,?/'\";:`~!@#$%^&*()_-+={[}]|\\"},
			},
		},
	}

	for _, tc := range testCases {
		mods := make([]Mutator, len(tc.pairs))
		for i, pair := range tc.pairs {
			mods[i] = UpsertString(pair.k, pair.v)
		}
		ts := NewTagSet(nil, mods...)

		encoded := Encode(ts)
		decoded, err := Decode(encoded)

		if err != nil {
			t.Errorf("%v: decoding encoded tagSet failed: %v", tc.label, err)
		}

		got := make([]pair, 0)
		for k, v := range decoded.m {
			ks, ok := k.(StringKey)
			if !ok {
				t.Errorf("%v: wrong key type; got %T, want StringKey", tc.label, k)
			}
			got = append(got, pair{ks, string(v)})
		}
		want := tc.pairs

		sort.Slice(got, func(i, j int) bool { return got[i].k.Name() < got[j].k.Name() })
		sort.Slice(want, func(i, j int) bool { return got[i].k.Name() < got[j].k.Name() })

		if !reflect.DeepEqual(got, tc.pairs) {
			t.Errorf("%v: decoded tagSet = %#v; want %#v", tc.label, got, want)
		}
	}
}
