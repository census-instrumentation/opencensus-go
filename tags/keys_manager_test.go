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

func Test_KeysManager_NoErrors(t *testing.T) {
	type testData struct {
		createCommands      []func() (Key, error)
		wantCount           int
		wantCountAfterClear int
	}

	testSet := []testData{
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k2") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k3") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k4") },
			},
			4,
			0,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
			},
			1,
			0,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k1") },
			},
			1,
			0,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k1") },
			},
			1,
			0,
		},
		{
			[]func() (Key, error){
			},
			0,
			0,
		},
	}

	for i, td := range testSet {
		DefaultKeyManager().Clear()
		for j, f := range td.createCommands {
			_, err := f()
			if err != nil {
				t.Errorf("got error %v, want no error calling DefaultKeyManager().CreateKeyXYZ(...). Test case: %v, function: %v", err, i, j)
			}
		}
		gotCount := DefaultKeyManager().Count()
		if gotCount != td.wantCount {
			t.Errorf("got keys count %v, want keys count %v", gotCount, td.wantCount)
		}

		DefaultKeyManager().Clear()
		gotCountAfterClear := DefaultKeyManager().Count()
		if gotCountAfterClear != td.wantCountAfterClear {
			t.Errorf("got keys count %v, want keys count %v after Clear()", gotCountAfterClear, td.wantCountAfterClear)
		}
	}
}

func Test_KeysManager_ExpectErrors(t *testing.T) {
	type testData struct {
		createCommands []func() (Key, error)
		wantErrCount   int
	}

	testSet := []testData{
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
			},
			2,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k1") },
			},
			1,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyInt64("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
			},
			1,
		},
		{
			[]func() (Key, error){
				func() (Key, error) { return DefaultKeyManager().CreateKeyBool("k1") },
				func() (Key, error) { return DefaultKeyManager().CreateKeyString("k1") },
			},
			1,
		},
	}

	for i, td := range testSet {
		gotErrCount := 0
		for _, f := range td.createCommands {
			_, err := f()
			if err != nil {
				gotErrCount++
			}
		}

		if gotErrCount != td.wantErrCount {
			t.Errorf("got errors count %v, want errors count %v. Test case %v", gotErrCount, td.wantErrCount, i)
		}
		DefaultKeyManager().Clear()
	}
}
