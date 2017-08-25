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
	"errors"
	"testing"
)

func Test_Worker_MeasureRegistration(t *testing.T) {
	someError := errors.New("some error")
	mf1 := NewMeasureFloat64("MF1", "desc MF1")
	mf2 := NewMeasureFloat64("MF2", "desc MF2")
	mf1SameName := NewMeasureInt64("MF1", "desc MI1. Duplicate measure with same name as MF1.")
	vw1 := NewViewFloat64("VF1", "desc VF1", nil, mf1, nil, nil)
	type registerWant struct {
		m   Measure
		err error
	}
	type unregisterWant struct {
		m   Measure
		err error
	}
	type byNameWant struct {
		name string
		m    Measure
		err  error
	}

	type testCase struct {
		label   string
		regs    []registerWant
		views   []View
		unregs  []unregisterWant
		bynames []byNameWant
	}
	tcs := []testCase{
		{
			"0",
			[]registerWant{
				{
					mf1,
					nil,
				},
			},
			[]View{},
			[]unregisterWant{},
			[]byNameWant{
				{
					mf1.Name(),
					mf1,
					nil,
				},
				{
					mf2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"1",
			[]registerWant{
				{
					mf1,
					nil,
				},
				{
					mf1,
					nil,
				},
			},
			[]View{},
			[]unregisterWant{},
			[]byNameWant{
				{
					mf1.Name(),
					mf1,
					nil,
				},
				{
					mf2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"2",
			[]registerWant{
				{
					mf1,
					nil,
				},
				{
					mf1,
					nil,
				},
			},
			[]View{},
			[]unregisterWant{
				{
					mf1,
					nil,
				},
			},
			[]byNameWant{
				{
					mf1.Name(),
					nil,
					someError,
				},
				{
					mf2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"3",
			[]registerWant{
				{
					mf1,
					nil,
				},
			},
			[]View{
				vw1,
			},
			[]unregisterWant{
				{
					mf1,
					someError,
				},
			},
			[]byNameWant{
				{
					mf1.Name(),
					mf1,
					nil,
				},
				{
					mf2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"4",
			[]registerWant{
				{
					mf1,
					nil,
				},
				{
					mf2,
					nil,
				},
			},
			[]View{
				vw1,
			},
			[]unregisterWant{
				{
					mf1,
					someError,
				},
				{
					mf2,
					nil,
				},
			},
			[]byNameWant{
				{
					mf1.Name(),
					mf1,
					nil,
				},
				{
					mf2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"5",
			[]registerWant{
				{
					mf1,
					nil,
				},
				{
					mf2,
					nil,
				},
				{
					mf1SameName,
					someError,
				},
			},
			[]View{},
			[]unregisterWant{
				{
					mf1SameName,
					nil,
				},
			},
			[]byNameWant{
				{
					mf1.Name(),
					mf1,
					nil,
				},
				{
					mf2.Name(),
					mf2,
					nil,
				},
			},
		},
	}

	for _, tc := range tcs {
		defaultWorker.stop()
		defaultWorker = newWorker()
		go defaultWorker.start()

		for _, reg := range tc.regs {
			err := RegisterMeasure(reg.m)
			if (err != nil) != (reg.err != nil) {
				t.Errorf("RegisterMeasure. got error %v, want %v. Test case: %v", err, reg.err, tc.label)
			}
		}

		for _, vw := range tc.views {
			err := RegisterView(vw)
			if err != nil {
				t.Errorf("RegisterView. got error %v, want no error. Test case: %v", err, tc.label)
			}
		}

		for _, unreg := range tc.unregs {
			err := UnregisterMeasure(unreg.m)
			if (err != nil) != (unreg.err != nil) {
				t.Errorf("UnregisterMeasure. got error %v, want %v. Test case: %v", err, unreg.err, tc.label)
			}
		}

		for _, byname := range tc.bynames {
			m, err := GetMeasureByName(byname.name)
			if (err != nil) != (byname.err != nil) {
				t.Errorf("GetMeasureByName. got error %v, want %v. Test case: %v", err, byname.err, tc.label)
			}

			if m != byname.m {
				t.Errorf("GetMeasureByName. got measure %v, want measure %v. Test case: %v", m, byname.m, tc.label)
			}
		}
		defaultWorker.stop()
	}
}

func Test_Worker_ViewRegistration(t *testing.T) {
	someError := errors.New("some error")
	mf1 := NewMeasureFloat64("MF1", "desc MF1")
	mf2 := NewMeasureFloat64("MF2", "desc MF2")

	vw1 := NewViewFloat64("VF1", "desc VF1", nil, mf1, nil, nil)
	vw1SameName := NewViewFloat64("VF1", "desc VF1.  Duplicate view with same name as VF1.", nil, mf1, nil, nil)
	vw2 := NewViewFloat64("VF2", "desc VF2", nil, mf2, nil, nil)
	sc1 := make(chan *ViewData)

	type registerWant struct {
		v        View
		err      error
		forAdhoc bool
	}
	type unregisterWant struct {
		v   View
		err error
	}
	type byNameWant struct {
		name string
		v    View
		err  error
	}

	type subscription struct {
		c   chan *ViewData
		v   View
		err error
	}
	type testCase struct {
		label         string
		regs          []registerWant
		subscriptions []subscription
		unregs        []unregisterWant
		bynames       []byNameWant
	}
	tcs := []testCase{
		{
			"0",
			[]registerWant{
				{
					vw1,
					nil,
					true,
				},
			},
			[]subscription{
				{
					sc1,
					vw1,
					nil,
				},
			},
			[]unregisterWant{
				{
					vw1,
					someError,
				},
			},
			[]byNameWant{
				{
					vw1.Name(),
					vw1,
					nil,
				},
				{
					vw2.Name(),
					nil,
					someError,
				},
			},
		},
		{
			"1",
			[]registerWant{
				{
					vw1,
					nil,
					false,
				},
				{
					vw2,
					nil,
					true,
				},
			},
			[]subscription{
				{
					sc1,
					vw1,
					nil,
				},
			},
			[]unregisterWant{
				{
					vw1,
					someError,
				},
				{
					vw2,
					someError,
				},
			},
			[]byNameWant{
				{
					vw1.Name(),
					vw1,
					nil,
				},
				{
					vw2.Name(),
					vw2,
					nil,
				},
			},
		},
		{
			"2",
			[]registerWant{
				{
					vw1,
					nil,
					true,
				},
			},
			[]subscription{
				{
					sc1,
					vw1,
					nil,
				},
				{
					sc1,
					vw1SameName,
					someError,
				},
			},
			[]unregisterWant{
				{
					vw1,
					someError,
				},
				{
					vw1SameName,
					nil,
				},
			},
			[]byNameWant{
				{
					vw1.Name(),
					vw1,
					nil,
				},
			},
		},
	}

	for _, tc := range tcs {
		defaultWorker.stop()
		defaultWorker = newWorker()
		go defaultWorker.start()

		_ = RegisterMeasure(mf1)
		_ = RegisterMeasure(mf2)

		for _, reg := range tc.regs {
			err := RegisterView(reg.v)
			if (err != nil) != (reg.err != nil) {
				t.Errorf("RegisterView. got error %v, want %v. Test case: %v", err, reg.err, tc.label)
			}
			StartCollectionForAdhoc(reg.v)
		}

		for _, s := range tc.subscriptions {
			err := SubscribeToView(s.v, s.c)
			if (err != nil) != (s.err != nil) {
				t.Errorf("SubscribeToView. got error %v, want %v. Test case: %v", err, s.err, tc.label)
			}
		}

		for _, unreg := range tc.unregs {
			err := UnregisterView(unreg.v)
			if (err != nil) != (unreg.err != nil) {
				t.Errorf("UnregisterView. got error %v, want %v. Test case: %v", err, unreg.err, tc.label)
			}
		}

		for _, byname := range tc.bynames {
			v, err := GetViewByName(byname.name)
			if (err != nil) != (byname.err != nil) {
				t.Errorf("GetViewByName. got error %v, want %v. Test case: %v", err, byname.err, tc.label)
			}

			if v != byname.v {
				t.Errorf("GetViewByName. got view '%v' '%v', want view %v. Test case: %v", v.Name(), v.Description(), byname.v, tc.label)
			}
		}
		defaultWorker.stop()
	}
}

/*
func Test_Worker_AdhocCollection(t *testing.T) {
	restartDefaultWorker()
}

func Test_Worker_Subscription(t *testing.T) {
	restartDefaultWorker()
}
*/

func Test_Worker_RecordFloat64(t *testing.T) {
	// register mf1, mf2 register vw1
	// record(mf1) -> nothing
	// record(mf2) -> nothing
	// adhoc vw1
	// record(mf1) -> 1
	// record(mf1)*2 -> 2
	// record(mf2) -> nothing
}

/*
func Test_Worker_RecordInt64(t *testing.T) {
	restartDefaultWorker()
}
*/

/*
		mf1 := NewMeasureFloat64("MF1", "desc MF1")
	type record struct {
		measurement measurementFloat64
		tags        []tagString
	}

[]record{
				{
					measurementFloat64{mf1, 1},
					[]tagString{{k1, "v1"}},
				},
				{
					measurementFloat64{mf1, 5},
					[]tagString{{k1, "v1"}},
				},
			},

		defaultWorker.stop()
		defaultWorker = newWorker()

		go defaultWorker.start()
		if err := RegisterMeasure(mf1); err != nil {
			t.Errorf("RegisterMeasure. Got error '%v'", err)
		}
		if err := RegisterView(vw1); err != nil {
			t.Errorf("RegisterView. Got error '%v'", err)
		}
		StartCollectionForAdhoc(vw1)
		ctx := tags.ContextWithNewTagSet(context.Background(), tsb.Build())
		RecordFloat64(ctx, r.measurement.m, r.measurement.v)

			gotRows, err := RetrieveData(vw1)
			if err != nil {
				t.Errorf("RetrieveData. Got error '%v' for test case: '%v'", err, tc.label)
			}
				StopCollectionForAdhoc(vw1)

		defaultWorker.stop()
*/
