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
		createCommands      []func(km *keysManager) (Key, error)
		wantCount           int
		wantCountAfterClear int
	}

	testSet := []testData{
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k2") },
			},
			2,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			1,
			0,
		},
		{
			[]func(km *keysManager) (Key, error){},
			0,
			0,
		},
	}

	for i, td := range testSet {
		km := newKeysManager()
		for j, f := range td.createCommands {
			_, err := f(km)
			if err != nil {
				t.Errorf("Test case '%v', function '%v': got error %v, want no error calling keysManager.createKeyXYZ(...)", i, j, err)
			}
		}
		gotCount := km.count()
		if gotCount != td.wantCount {
			t.Errorf("Test case '%v': got keys count %v, want keys count %v", i, gotCount, td.wantCount)
		}

		km.clear()
		gotCountAfterClear := km.count()
		if gotCountAfterClear != td.wantCountAfterClear {
			t.Errorf("Test case '%v': got keys count %v, want keys count %v after clear()", i, gotCountAfterClear, td.wantCountAfterClear)
		}
	}
}

func Test_KeysManager_InvalidKeyErrors(t *testing.T) {
	type testData struct {
		createCommands []func(km *keysManager) (Key, error)
		wantCount      int
	}

	testSet := []testData{
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
			},
			1,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k\xb0") },
			},
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k1") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k\xb0") },
			},
			1,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x19") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x7f") },
			},
			0,
		},
		{
			[]func(km *keysManager) (Key, error){
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x19") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x20") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x7e") },
				func(km *keysManager) (Key, error) { return km.createKeyString("k\x7f") },
			},
			2,
		},
		{
			[]func(km *keysManager) (Key, error){},
			0,
		},
	}

	for i, td := range testSet {
		km := newKeysManager()
		for _, f := range td.createCommands {
			_, _ = f(km)
		}
		gotCount := km.count()
		if gotCount != td.wantCount {
			t.Errorf("Test case '%v': got keys count %v, want keys count %v", i, gotCount, td.wantCount)
		}
	}
}
