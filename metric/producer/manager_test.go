// Copyright 2019, OpenCensus Authors
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

package producer

import (
	"testing"
	"go.opencensus.io/metric/metricdata"
)

type testProducer struct {
	name string
}

var (
	myProd1 = NewTestProducer("foo")
	myProd2 = NewTestProducer("bar")
	myProd3 = NewTestProducer("foobar")
)

func NewTestProducer(name string) *testProducer {
	return &testProducer{name}
}

func (mp *testProducer) Read() []*metricdata.Metric {
	return nil
}

func TestAdd(t *testing.T) {
	Add(myProd1)
	Add(myProd2)

	got := GetAll()
	want := []*testProducer{myProd1, myProd2}
	checkSlice("add test", got, want, t)
}

func TestAddSame(t *testing.T) {
	Add(myProd1)
	Add(myProd2)
	Add(myProd1)

	got := GetAll()
	want := []*testProducer{myProd1, myProd2}
	checkSlice("add test", got, want, t)
}

func TestDelete(t *testing.T) {
	Add(myProd1)
	Add(myProd2)
	Add(myProd3)
	Delete(myProd2)

	got := GetAll()
	want := []*testProducer{myProd1, myProd3}
	checkSlice("add test", got, want, t)
}

func TestDeleteNonExisting(t *testing.T) {
	Add(myProd1)
	Add(myProd3)
	Delete(myProd2)

	got := GetAll()
	want := []*testProducer{myProd1, myProd3}
	checkSlice("add test", got, want, t)
}

func checkSlice(testName string, got []Producer, want []*testProducer, t *testing.T) {
	gotLen := len(got)
	wantLen := len(want)
	if gotLen != wantLen {
		t.Errorf("test: %s, got len: %d want: %d\n", testName, gotLen, wantLen)
	} else {
		for i := 0 ; i<gotLen; i++ {
			if got[i] != want[i] {
				t.Errorf("test: %s, at index %d, got %p, want %p\n", testName, i, got[i], want[i])
			}
		}
	}
}
