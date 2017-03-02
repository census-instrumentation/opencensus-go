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

package stats

import (
	"fmt"
	"testing"
)

func TestRegisterMeasureDesc(t *testing.T) {
	type testData struct {
		mds            []MeasureDesc
		wantErrorCount int
	}
	measurementUnitString := &MeasurementUnit{}
	md1 := NewMeasureDescFloat64("name1", "desc", measurementUnitString)
	md2 := NewMeasureDescFloat64("name2", "desc", measurementUnitString)
	md3 := NewMeasureDescFloat64("name1", "desc", measurementUnitString)

	testCases := []testData{
		{
			[]MeasureDesc{md1, md2},
			0,
		},
		{
			[]MeasureDesc{md2, md1},
			0,
		},
		{
			[]MeasureDesc{md1, md2, md3},
			1,
		},
	}

	for i, td := range testCases {
		cw := newChannelWorker()
		var got int
		for _, md := range td.mds {
			if err := cw.registerMeasureDesc(md); err != nil {
				got++
			}
		}
		if got != td.wantErrorCount {
			t.Errorf("got '%v', want '%v' when registering test case %v", got, td.wantErrorCount, i)
		}
	}
}

func TestUnregisterMeasureDesc(t *testing.T) {
	type testData struct {
		mds        []MeasureDesc
		unregister string
		want       string
	}

	md1 := NewMeasureDescFloat64("name1", "desc", &MeasurementUnit{})
	md2 := NewMeasureDescFloat64("name2", "desc", &MeasurementUnit{})

	testCases := []testData{
		{
			[]MeasureDesc{md1, md2},
			"name1",
			"",
		},
		{
			[]MeasureDesc{md1, md2},
			"name2",
			"",
		},
		{
			[]MeasureDesc{md1, md2},
			"name3",
			fmt.Sprintf("no measure descriptor with the name %s is registered", "name3"),
		},
	}

	for i, td := range testCases {
		cw := newChannelWorker()
		for _, md := range td.mds {
			if err := cw.registerMeasureDesc(md); err != nil {
				t.Errorf("got '%v' during registration test case %v, want no error", err, i)
			}
		}

		var got string
		if err := cw.unregisterMeasureDesc(td.unregister); err != nil {
			got = err.Error()
		}

		if got != td.want {
			t.Errorf("got '%v', want '%v' when unregistering test case %v", got, td.want, i)
		}

		// re-register a measureDesc with the same name as the one unregistered
		tmp := NewMeasureDescFloat64(td.unregister, "desc", &MeasurementUnit{})

		if err := cw.registerMeasureDesc(tmp); err != nil {
			t.Errorf("got '%v' during registration '%v' after registering test case %v, want no error", err, tmp, i)
		}
	}
}
