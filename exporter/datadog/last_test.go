// Copyright 2018, OpenCensus Authors
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

package datadog // import "go.opencensus.io/exporter/datadog"

import (
	"testing"

	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func TestLastValues(t *testing.T) {
	var (
		viewName   = "viewName"
		tagHash    = uint64(123)
		value      = view.CountData(123)
		want       = &value
		lastValues = newLastValues()
	)

	// store
	lastValues.store(viewName, tagHash, want)

	// retrieve
	got, ok := lastValues.lookup(viewName, tagHash)
	if !ok {
		t.Error("want true; got false")
	}
	if want != got {
		t.Errorf("want %v; got %v", want, got)
	}
}

func BenchmarkLastValuesLookup(t *testing.B) {
	var (
		viewName   = "viewName"
		tagHash    = uint64(123)
		value      = view.CountData(123)
		want       = &value
		lastValues = newLastValues()
	)

	lastValues.store(viewName, tagHash, want)

	for i := 0; i < t.N; i++ {
		got, ok := lastValues.lookup(viewName, tagHash)
		if !ok {
			t.Error("want true; got false")
		}
		if got := got; want != got {
			t.Errorf("want %v; got %v", want, got)
		}
	}
}

func TestHasher(t *testing.T) {
	var h = newHasher()

	t.Run("nil", func(t *testing.T) {
		if want := uint64(17241709254077376921); h.Hash(nil) != want {
			t.Errorf("want %v; got %v", want, h.Hash(nil))
		}
	})

	t.Run("tag", func(t *testing.T) {
		var (
			key1, _ = tag.NewKey("key1")
			tag1    = tag.Tag{
				Key:   key1,
				Value: "value",
			}
			tags = []tag.Tag{tag1}
		)

		if want := uint64(10941932480703334251); h.Hash(tags) != want {
			t.Errorf("want %v; got %v", want, h.Hash(tags))
		}
	})
}

func BenchmarkHasher(t *testing.B) {
	var (
		key1, _ = tag.NewKey("key1")
		key2, _ = tag.NewKey("key2")
		tag1    = tag.Tag{
			Key:   key1,
			Value: "value",
		}
		tag2 = tag.Tag{
			Key:   key2,
			Value: "value",
		}
		tags = []tag.Tag{tag1, tag2}
		h    = newHasher()
	)

	for i := 0; i < t.N; i++ {
		h.Hash(tags)
	}
}
