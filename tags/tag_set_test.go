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
	type testCase struct {
		insert []Tag
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testCases := []testCase{
		{
			[]Tag{
				Tag{
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
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
				Tag{
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

	for i, tc := range testCases {
		ts := newTagSet(0)
		for _, insertPair := range tc.insert {
			_ = ts.insertBytes(insertPair.K, insertPair.V)
		}

		for _, wantPair := range tc.want {
			got, _ := ts.ValueAsString(wantPair.k)
			if got != wantPair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, wantPair.k, got, wantPair.v)
			}
		}
	}
}

func Test_Tagset_Upsert(t *testing.T) {
	type want struct {
		k Key
		v string
	}
	type testCase struct {
		upsert []Tag
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testCases := []testCase{
		{
			[]Tag{
				Tag{
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
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
				Tag{
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

	for i, tc := range testCases {
		ts := newTagSet(0)
		for _, upsertPair := range tc.upsert {
			ts.upsertBytes(upsertPair.K, upsertPair.V)
		}

		for _, wantPair := range tc.want {
			got, _ := ts.ValueAsString(wantPair.k)
			if got != wantPair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, wantPair.k, got, wantPair.v)
			}
		}
	}
}

func Test_Tagset_Update(t *testing.T) {
	type want struct {
		k Key
		v string
	}
	type testCase struct {
		insert []Tag
		update []Tag
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testCases := []testCase{
		{
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Tag{
				Tag{
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
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
				Tag{
					k2,
					[]byte("v2"),
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
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Tag{
				Tag{
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
		{
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Tag{
				Tag{
					k1,
					[]byte("v1new"),
				},
				Tag{
					k1,
					[]byte("v1latest"),
				},
			},
			[]*want{
				&want{
					k1,
					"v1latest",
				},
				&want{
					k2,
					"",
				},
			},
		},
	}

	for i, tc := range testCases {
		ts := newTagSet(0)
		for _, insertPair := range tc.insert {
			_ = ts.insertBytes(insertPair.K, insertPair.V)
		}

		for _, updatePair := range tc.update {
			_ = ts.updateBytes(updatePair.K, updatePair.V)
		}

		for _, wantPair := range tc.want {
			got, _ := ts.ValueAsString(wantPair.k)
			if got != wantPair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, wantPair.k, got, wantPair.v)
			}
		}
	}
}

func Test_Tagset_Delete(t *testing.T) {
	type want struct {
		k Key
		v string
	}
	type testCase struct {
		insert []Tag
		delete []Key
		want   []*want
	}

	km := newKeysManager()
	k1, _ := km.createKeyString("k1")
	k2, _ := km.createKeyString("k2")
	testCases := []testCase{
		{
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Key{
				k2,
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
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
			},
			[]Key{
				k1,
			},
			[]*want{
				&want{
					k1,
					"",
				},
				&want{
					k2,
					"",
				},
			},
		},
		{
			[]Tag{
				Tag{
					k1,
					[]byte("v1"),
				},
				Tag{
					k1,
					[]byte("v1new"),
				},
			},
			[]Key{
				k1,
			},
			[]*want{
				&want{
					k1,
					"",
				},
				&want{
					k2,
					"",
				},
			},
		},
	}

	for i, tc := range testCases {
		ts := newTagSet(0)
		for _, insertPair := range tc.insert {
			ts.upsertBytes(insertPair.K, insertPair.V)
		}

		for _, deleteK := range tc.delete {
			ts.delete(deleteK)
		}

		for _, wantPair := range tc.want {
			got, _ := ts.ValueAsString(wantPair.k)
			if got != wantPair.v {
				t.Errorf("Test case '%v' key '%v': got string %v, want string %v", i, wantPair.k, got, wantPair.v)
			}
		}
	}
}
