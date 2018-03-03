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
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

func TestImplementsViewExporter(t *testing.T) {
	var (
		err      error
		exporter view.Exporter
	)

	exporter, err = NewExporter()
	if err != nil {
		t.Fatalf("want nil; got %v", err)
	}
	if exporter == nil {
		t.Fatalf("Exported must implement view.Exporter")
	}
}

type point struct {
	value float64
	tags  []string
}

type Mock struct {
	mutex      sync.Mutex
	counts     map[string][]point
	gauges     map[string][]point
	histograms map[string][]point
}

func (m *Mock) Count(name string, value int64, tags []string, rate float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.counts == nil {
		m.counts = map[string][]point{}
	}

	m.counts[name] = append(m.counts[name], point{
		value: float64(value),
		tags:  tags,
	})

	return nil
}

func (m *Mock) Gauge(name string, value float64, tags []string, rate float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.gauges == nil {
		m.gauges = map[string][]point{}
	}

	m.gauges[name] = append(m.gauges[name], point{
		value: value,
		tags:  tags,
	})

	return nil
}

func (m *Mock) Histogram(name string, value float64, tags []string, rate float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.histograms == nil {
		m.histograms = map[string][]point{}
	}

	m.histograms[name] = append(m.histograms[name], point{
		value: value,
		tags:  tags,
	})

	return nil
}

func validateCount(t *testing.T, m *Mock, name string, want int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.counts == nil {
		t.Fatalf("want not nil; got nil")
	}
	if got := len(m.counts[name]); got == 0 {
		t.Errorf("want 1; got 0")
	}

	var sum float64
	for _, item := range m.counts[name] {
		if want := []string{"key:value"}; !reflect.DeepEqual(want, item.tags) {
			t.Errorf("want %v; got %v", want, item.tags)
		}
		sum += item.value
	}

	if sum != float64(want) {
		t.Errorf("want %v; got %v", want, sum)
	}
}

func validateHistogram(t *testing.T, m *Mock, name string, want int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.histograms == nil {
		t.Fatalf("want not nil; got nil")
	}
	if got := len(m.histograms[name]); got == 0 {
		t.Errorf("want 1; got 0")
	}

	var sum float64
	for _, item := range m.histograms[name] {
		if want := []string{"key:value"}; !reflect.DeepEqual(want, item.tags) {
			t.Errorf("want %v; got %v", want, item.tags)
		}
		sum += item.value
	}

	if sum != float64(want) {
		t.Errorf("want %v; got %v", want, sum)
	}
}

func validateGauge(t *testing.T, m *Mock, name string, want int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.gauges == nil {
		t.Fatalf("want not nil; got nil")
	}
	if got := len(m.gauges[name]); got == 0 {
		t.Errorf("want 1; got 0")
	}

	for _, item := range m.gauges[name] {
		if want := []string{"key:value"}; !reflect.DeepEqual(want, item.tags) {
			t.Errorf("want %v; got %v", want, item.tags)
		}

		if item.value != float64(want) {
			t.Errorf("want %v; got %v", want, item.value)
		}
	}
}

func TestExporter(t *testing.T) {
	var (
		tagKey, _ = tag.NewKey("key")
		tagValue  = "value"
		m, _      = stats.Int64("m", "description connections", "")
		want      = int64(1)
	)

	testCases := map[string]struct {
		Name        string
		Want        int64
		Aggregation view.Aggregation
		Validate    func(t *testing.T, m *Mock, name string, want int64)
	}{
		"count": {
			Name:        "counts",
			Want:        1,
			Aggregation: view.CountAggregation{},
			Validate:    validateCount,
		},
		"sum": {
			Name:        "sum",
			Want:        1,
			Aggregation: view.SumAggregation{},
			Validate:    validateCount,
		},
		"mean": {
			Name:        "mean",
			Want:        1,
			Aggregation: view.MeanAggregation{},
			Validate:    validateGauge,
		},
		"distribution": {
			Name:        "distribution",
			Want:        100,
			Aggregation: view.DistributionAggregation{100},
			Validate:    validateHistogram,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			var (
				myView, _   = view.New(tc.Name, "blah", []tag.Key{tagKey}, m, tc.Aggregation)
				mock        = &Mock{}
				exporter, _ = NewExporter(WithClient(mock))
				interval    = 100 * time.Millisecond
			)

			view.RegisterExporter(exporter)
			view.SetReportingPeriod(interval)

			if err := myView.Subscribe(); err != nil {
				t.Fatalf("want nil; got %v", err)
			}

			// When
			ctx, _ := tag.New(context.Background(), tag.Insert(tagKey, tagValue))

			stats.Record(ctx, m.M(want))
			time.Sleep(2 * interval)

			// Then
			myView.Unsubscribe()

			tc.Validate(t, mock, tc.Name, tc.Want)
		})
	}
}

func makeView(key tag.Key, m stats.Measure, name string, aggregation view.Aggregation) *view.View {
	v, err := view.New(
		name,
		"blah",
		[]tag.Key{key},
		m,
		aggregation,
	)
	if err != nil {
		log.Fatalln(err)
	}
	return v
}

func TestLiveExporter(t *testing.T) {
	addr := os.Getenv("DATADOG_ADDR")
	if addr == "" {
		t.SkipNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exporter, err := NewExporter(WithStatsAddr(addr))
	if err != nil {
		t.Fatalf("want nil; got %v", err)
	}

	view.RegisterExporter(exporter)
	view.SetReportingPeriod(3 * time.Second)

	m, err := stats.Int64("my.org/measure/openconns", "open connections", "")
	if err != nil {
		log.Fatal(err)
	}

	var (
		key, _       = tag.NewKey("key")
		sum          = makeView(key, m, "opencensus.sample.sum", view.SumAggregation{})
		count        = makeView(key, m, "opencensus.sample.count", view.CountAggregation{})
		mean         = makeView(key, m, "opencensus.sample.mean", view.MeanAggregation{})
		distribution = makeView(key, m, "opencensus.sample.distribution", view.DistributionAggregation{1, 2, 3, 4, 5})
	)

	sum.Subscribe()
	defer sum.Unsubscribe()

	count.Subscribe()
	defer count.Unsubscribe()

	mean.Subscribe()
	defer mean.Unsubscribe()

	distribution.Subscribe()
	defer distribution.Unsubscribe()

	ctx, _ = tag.New(ctx, tag.Insert(key, "value"))
	iterations := 60
	for i := 1; i <= iterations; i++ {
		value := int64(i % 4)
		fmt.Printf("%v of %v: stats.Record(%v)\n", i, iterations, value)
		stats.Record(ctx, m.M(value))
		time.Sleep(time.Second)
	}
}

func TestFixMetricName(t *testing.T) {
	testCases := map[string]struct {
		Name     string
		Expected string
		Ok       bool
	}{
		"simple": {
			Name:     "abc",
			Expected: "abc",
			Ok:       true,
		},
		"invalid name": {
			Name:     "",
			Expected: "",
			Ok:       false,
		},
		"dot ok": {
			Name:     "abc.",
			Expected: "abc.",
			Ok:       true,
		},
		"must starts with alpha": {
			Name:     "0123abc",
			Expected: "abc",
			Ok:       true,
		},
		"slashes to dots": {
			Name:     "a/b/c",
			Expected: "a.b.c",
			Ok:       true,
		},
		"replace invalid with _": {
			Name:     "abc[]",
			Expected: "abc__",
			Ok:       true,
		},
		"strips if too long": {
			Name:     makeStringN(maxMetricNameLength + 1),
			Expected: makeStringN(maxMetricNameLength),
			Ok:       true,
		},
	}

	for label, tc := range testCases {
		t.Run(label, func(t *testing.T) {
			name, ok := fixMetricName(tc.Name)
			if want := tc.Ok; ok != want {
				t.Errorf("want %v; got %v", want, ok)
			}
			if want := tc.Expected; name != want {
				t.Errorf("want %v; got %v", want, name)
			}
		})
	}
}

// makeStringN returns a string of the specified length
func makeStringN(length int) string {
	var content []byte
	for i := 0; i < length; i++ {
		content = append(content, "a"...)
	}
	return string(content)
}

func TestTypeCast(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		var foo view.AggregationData
		_, ok := foo.(*view.CountData)
		if ok {
			t.Errorf("want false; got true")
		}
	})

	t.Run("nil", func(t *testing.T) {
		var (
			want                      = view.CountData(123)
			data view.AggregationData = &want
		)

		got, ok := data.(*view.CountData)
		if !ok {
			t.Errorf("want true; got false")
		}
		if got != &want {
			t.Errorf("want %v; got %v", want, got)
		}
	})
}

func TestBuildConfig(t *testing.T) {
	t.Run("sets default addr", func(t *testing.T) {
		c, err := buildConfig()
		if err != nil {
			t.Errorf("want nil; got %v", err)
		}
		if c.statsAddr != defaultStatsAddr {
			t.Errorf("want %v; got %v", defaultStatsAddr, c.statsAddr)
		}
	})
}
