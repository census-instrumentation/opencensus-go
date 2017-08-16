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

import "testing"

func Test_Tagset_Insert(t *testing.T) {
	type want struct {
		k Key
		v string
	}
	type testData struct {
		insert []*Tag
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testSet := []testData{
		{
			[]*Tag{
				&Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]*want{
				&want{
					k1,
					"v1",
				},
				&want{
					k2,
					"",
				},
			},
		},
		{
			[]*Tag{
				&Tag{
					k1,
					[]byte("v1"),
				},
				&Tag{
					k1,
					[]byte("v1new"),
				},
			},
			[]*want{
				&want{
					k1,
					"v1",
				},
				&want{
					k2,
					"",
				},
			},
		},
	}

	for i, td := range testSet {
		ts := newTagSet(0)
		for _, pair := range td.insert {
			_ = ts.insertBytes(pair.K, pair.V)
		}

		for _, pair := range td.want {
			got, _ := ts.ValueAsString(pair.k)
			if got != pair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, pair.k, got, pair.v)
			}
		}
	}
}

func Test_Tagset_Upsert(t *testing.T) {
	type want struct {
		k Key
		v string
	}
	type testData struct {
		insert []*Tag
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testSet := []testData{
		{
			[]*Tag{
				&Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]*want{
				&want{
					k1,
					"v1",
				},
				&want{
					k2,
					"",
				},
			},
		},
		{
			[]*Tag{
				&Tag{
					k1,
					[]byte("v1"),
				},
				&Tag{
					k1,
					[]byte("v1new"),
				},
			},
			[]*want{
				&want{
					k1,
					"v1new",
				},
				&want{
					k2,
					"",
				},
			},
		},
	}

	for i, td := range testSet {
		ts := newTagSet(0)
		for _, pair := range td.insert {
			ts.upsertBytes(pair.K, pair.V)
		}

		for _, pair := range td.want {
			got, _ := ts.ValueAsString(pair.k)
			if got != pair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, pair.k, got, pair.v)
			}
		}
	}
}
