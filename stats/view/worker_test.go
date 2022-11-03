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

package view

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/resource"

	"go.opencensus.io/metric/metricdata"
	"go.opencensus.io/metric/metricexport"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

func Test_Worker_ViewRegistration(t *testing.T) {
	someError := errors.New("some error")

	sc1 := make(chan *Data)

	type registration struct {
		c   chan *Data
		vID string
		err error
	}
	type testCase struct {
		label         string
		registrations []registration
	}
	tcs := []testCase{
		{
			"register v1ID",
			[]registration{
				{
					sc1,
					"v1ID",
					nil,
				},
			},
		},
		{
			"register v1ID+v2ID",
			[]registration{
				{
					sc1,
					"v1ID",
					nil,
				},
			},
		},
		{
			"register to v1ID; ??? to v1ID and view with same ID",
			[]registration{
				{
					sc1,
					"v1ID",
					nil,
				},
				{
					sc1,
					"v1SameNameID",
					someError,
				},
			},
		},
	}

	mf1 := stats.Float64("MF1/Test_Worker_ViewSubscription", "desc MF1", "unit")
	mf2 := stats.Float64("MF2/Test_Worker_ViewSubscription", "desc MF2", "unit")

	for _, tc := range tcs {
		t.Run(tc.label, func(t *testing.T) {
			restart()

			views := map[string]*View{
				"v1ID": {
					Name:        "VF1",
					Measure:     mf1,
					Aggregation: Count(),
				},
				"v1SameNameID": {
					Name:        "VF1",
					Description: "desc duplicate name VF1",
					Measure:     mf1,
					Aggregation: Sum(),
				},
				"v2ID": {
					Name:        "VF2",
					Measure:     mf2,
					Aggregation: Count(),
				},
				"vNilID": nil,
			}

			for _, r := range tc.registrations {
				v := views[r.vID]
				err := Register(v)
				if (err != nil) != (r.err != nil) {
					t.Errorf("%v: Register() = %v, want %v", tc.label, err, r.err)
				}
			}
		})
	}
}

func Test_Worker_MultiExport(t *testing.T) {
	restart()

	// This test reports the same data for the default worker and a secondary
	// worker, and ensures that the stats are kept independently.
	extraResource := resource.Resource{
		Type:   "additional",
		Labels: map[string]string{"key1": "value1", "key2": "value2"},
	}
	worker2 := NewMeter().(*worker)
	worker2.Start()
	worker2.SetResource(&extraResource)

	m := stats.Float64("Test_Worker_MultiExport/MF1", "desc MF1", "unit")
	key := tag.MustNewKey(("key"))
	count := &View{"VF1", "description", []tag.Key{key}, m, Count()}
	sum := &View{"VF2", "description", []tag.Key{}, m, Sum()}

	Register(count, sum)
	worker2.Register(count) // Don't compute the sum for worker2, to verify independence of computation.
	data := []struct {
		w     Meter
		tags  string // Tag values
		value float64
	}{{
		tags:  "a",
		value: 2.0,
	}, {
		tags:  "b",
		value: 3.0,
	}, {
		tags: "a", value: 2.5,
	}, {
		w: worker2, tags: "b", value: 1.0,
	},
	}

	for _, d := range data {
		ctx, err := tag.New(context.Background(), tag.Upsert(key, d.tags))
		if err != nil {
			t.Fatalf("%s: failed to add tag %q: %v", d.w, key.Name(), err)
		}
		if d.w != nil {
			d.w.Record(tag.FromContext(ctx), []stats.Measurement{m.M(d.value)}, nil)
		} else {
			stats.Record(ctx, m.M(d.value))
		}
	}

	makeKey := func(r *resource.Resource, view string) string {
		if r == nil {
			r = &resource.Resource{}
		}
		return resource.EncodeLabels(r.Labels) + "/" + view
	}

	// Format is Resource.Labels encoded as string, then
	wantPartialData := map[string][]*Row{
		makeKey(nil, count.Name): {
			{[]tag.Tag{{Key: key, Value: "a"}}, &CountData{Value: 2}},
			{[]tag.Tag{{Key: key, Value: "b"}}, &CountData{Value: 1}},
		},
		makeKey(nil, sum.Name): {
			{nil, &SumData{Value: 7.5}},
		},
		makeKey(&extraResource, count.Name): {
			{[]tag.Tag{{Key: key, Value: "b"}}, &CountData{Value: 1}},
		},
	}

	te := &testExporter{}
	metricexport.NewReader().ReadAndExport(te)
	for _, m := range te.metrics {
		key := makeKey(m.Resource, m.Descriptor.Name)
		want, ok := wantPartialData[key]
		if !ok {
			t.Errorf("Unexpected data for %q: %v", key, m)
			continue
		}
		gotTs := m.TimeSeries
		sort.Sort(byLabel(gotTs))

		for i, ts := range gotTs {
			for j, label := range ts.LabelValues {
				if want[i].Tags[j].Value != label.Value {
					t.Errorf("Mismatched tag values (want %q, got %q) for %v in %q", want[i].Tags[j].Value, label.Value, ts, key)
				}
			}
			switch wantValue := want[i].Data.(type) {
			case *CountData:
				got := ts.Points[0].Value.(int64)
				if wantValue.Value != got {
					t.Errorf("Mismatched value (want %d, got %d) for %v in %q", wantValue.Value, got, ts, key)
				}
			case *SumData:
				got := ts.Points[0].Value.(float64)
				if wantValue.Value != got {
					t.Errorf("Mismatched value (want %f, got %f) for %v in %q", wantValue.Value, got, ts, key)
				}
			default:
				t.Errorf("Unexpected type of data: %T for %v in %q", wantValue, want[i], key)
			}
		}
	}

	// Verify that worker has not been computing sum:
	got, err := worker2.RetrieveData(sum.Name)
	if err == nil {
		t.Errorf("%s: expected no data because it was not registered, got %#v", sum.Name, got)
	}

	Unregister(count, sum)
	worker2.Unregister(count)
	worker2.Stop()
}

func Test_Worker_RecordFloat64(t *testing.T) {
	restart()

	someError := errors.New("some error")
	m := stats.Float64("Test_Worker_RecordFloat64/MF1", "desc MF1", "unit")

	k1 := tag.MustNewKey("k1")
	k2 := tag.MustNewKey("k2")
	ctx, err := tag.New(context.Background(),
		tag.Insert(k1, "v1"),
		tag.Insert(k2, "v2"),
	)
	if err != nil {
		t.Fatal(err)
	}

	v1 := &View{"VF1", "desc VF1", []tag.Key{k1, k2}, m, Count()}
	v2 := &View{"VF2", "desc VF2", []tag.Key{k1, k2}, m, Count()}

	type want struct {
		v    *View
		rows []*Row
		err  error
	}
	type testCase struct {
		label         string
		registrations []*View
		records       []float64
		wants         []want
	}

	tcs := []testCase{
		{
			label:         "0",
			registrations: []*View{},
			records:       []float64{1, 1},
			wants:         []want{{v1, nil, someError}, {v2, nil, someError}},
		},
		{
			label:         "1",
			registrations: []*View{v1},
			records:       []float64{1, 1},
			wants: []want{
				{
					v1,
					[]*Row{
						{
							[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
							&CountData{Value: 2},
						},
					},
					nil,
				},
				{v2, nil, someError},
			},
		},
		{
			label:         "2",
			registrations: []*View{v1, v2},
			records:       []float64{1, 1},
			wants: []want{
				{
					v1,
					[]*Row{
						{
							[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
							&CountData{Value: 2},
						},
					},
					nil,
				},
				{
					v2,
					[]*Row{
						{
							[]tag.Tag{{Key: k1, Value: "v1"}, {Key: k2, Value: "v2"}},
							&CountData{Value: 2},
						},
					},
					nil,
				},
			},
		},
	}

	for _, tc := range tcs {
		for _, v := range tc.registrations {
			if err := Register(v); err != nil {
				t.Fatalf("%v: Register(%v) = %v; want no errors", tc.label, v.Name, err)
			}
		}

		for _, value := range tc.records {
			stats.Record(ctx, m.M(value))
		}

		for _, w := range tc.wants {
			gotRows, err := RetrieveData(w.v.Name)
			for i := range gotRows {
				switch data := gotRows[i].Data.(type) {
				case *CountData:
					data.Start = time.Time{}
				case *SumData:
					data.Start = time.Time{}
				case *DistributionData:
					data.Start = time.Time{}
				}
			}
			if (err != nil) != (w.err != nil) {
				t.Fatalf("%s: RetrieveData(%v) = %v; want error = %v", tc.label, w.v.Name, err, w.err)
			}
			if diff := cmp.Diff(gotRows, w.rows); diff != "" {
				t.Errorf("%v: unexpected row (got-, want+): %s", tc.label, diff)
				break
			}
		}

		// Cleaning up.
		Unregister(tc.registrations...)
	}
}

func TestReportUsage(t *testing.T) {
	ctx := context.Background()

	m := stats.Int64("measure", "desc", "unit")

	tests := []struct {
		name         string
		view         *View
		wantMaxCount int64
	}{
		{
			name:         "cum",
			view:         &View{Name: "cum1", Measure: m, Aggregation: Count()},
			wantMaxCount: 8,
		},
		{
			name:         "cum2",
			view:         &View{Name: "cum1", Measure: m, Aggregation: Count()},
			wantMaxCount: 8,
		},
	}

	for _, tt := range tests {
		restart()
		SetReportingPeriod(25 * time.Millisecond)

		if err := Register(tt.view); err != nil {
			t.Fatalf("%v: cannot register: %v", tt.name, err)
		}

		e := &countExporter{}
		RegisterExporter(e)

		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))

		time.Sleep(50 * time.Millisecond)

		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))
		stats.Record(ctx, m.M(1))

		time.Sleep(50 * time.Millisecond)

		e.Lock()
		count := e.count
		e.Unlock()
		if got, want := count, tt.wantMaxCount; got > want {
			t.Errorf("%v: got count data = %v; want at most %v", tt.name, got, want)
		}
	}

}

func Test_SetReportingPeriodReqNeverBlocks(t *testing.T) {
	t.Parallel()

	worker := NewMeter().(*worker)
	durations := []time.Duration{-1, 0, 10, 100 * time.Millisecond}
	for i, duration := range durations {
		ackChan := make(chan bool, 1)
		cmd := &setReportingPeriodReq{c: ackChan, d: duration}
		cmd.handleCommand(worker)

		select {
		case <-ackChan:
		case <-time.After(500 * time.Millisecond): // Arbitrarily using 500ms as the timeout duration.
			t.Errorf("#%d: duration %v blocks", i, duration)
		}
	}
}

func TestWorkerStarttime(t *testing.T) {
	restart()

	ctx := context.Background()
	m := stats.Int64("measure/TestWorkerStarttime", "desc", "unit")
	v := &View{
		Name:        "testview",
		Measure:     m,
		Aggregation: Count(),
	}

	SetReportingPeriod(25 * time.Millisecond)
	if err := Register(v); err != nil {
		t.Fatalf("cannot register to %v: %v", v.Name, err)
	}

	e := &vdExporter{}
	RegisterExporter(e)
	defer UnregisterExporter(e)

	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))

	time.Sleep(50 * time.Millisecond)

	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))
	stats.Record(ctx, m.M(1))

	time.Sleep(50 * time.Millisecond)

	e.Lock()
	if len(e.vds) == 0 {
		t.Fatal("Got no view data; want at least one")
	}

	var start time.Time
	for _, vd := range e.vds {
		if start.IsZero() {
			start = vd.Start
		}
		if !vd.Start.Equal(start) {
			t.Errorf("Cumulative view data start time = %v; want %v", vd.Start, start)
		}
	}
	e.Unlock()
}

func TestUnregisterReportsUsage(t *testing.T) {
	restart()
	ctx := context.Background()

	m1 := stats.Int64("measure", "desc", "unit")
	view1 := &View{Name: "count", Measure: m1, Aggregation: Count()}
	m2 := stats.Int64("measure2", "desc", "unit")
	view2 := &View{Name: "count2", Measure: m2, Aggregation: Count()}

	SetReportingPeriod(time.Hour)

	if err := Register(view1, view2); err != nil {
		t.Fatalf("cannot register: %v", err)
	}

	e := &countExporter{}
	RegisterExporter(e)

	stats.Record(ctx, m1.M(1))
	stats.Record(ctx, m2.M(1))
	stats.Record(ctx, m2.M(1))

	Unregister(view2)

	// Unregister should only flush view2, so expect the count of 2.
	want := int64(2)

	e.Lock()
	got := e.totalCount
	e.Unlock()
	if got != want {
		t.Errorf("got count data = %v; want %v", got, want)
	}
}

func TestWorkerRace(t *testing.T) {
	restart()
	ctx := context.Background()

	m1 := stats.Int64("measure", "desc", "unit")
	view1 := &View{Name: "count", Measure: m1, Aggregation: Count()}
	m2 := stats.Int64("measure2", "desc", "unit")
	view2 := &View{Name: "count2", Measure: m2, Aggregation: Count()}

	// 1. This will export every microsecond.
	SetReportingPeriod(time.Microsecond)

	if err := Register(view1, view2); err != nil {
		t.Fatalf("cannot register: %v", err)
	}

	e := &countExporter{}
	RegisterExporter(e)

	// Synchronize and make sure every goroutine has terminated before we exit
	var waiter sync.WaitGroup
	waiter.Add(3)
	defer waiter.Wait()

	doneCh := make(chan bool)
	// 2. Record write routine at 700ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(700 * time.Nanosecond)
		defer tick.Stop()

		defer func() {
			close(doneCh)
		}()

		for i := 0; i < 1e3; i++ {
			stats.Record(ctx, m1.M(1))
			stats.Record(ctx, m2.M(1))
			stats.Record(ctx, m2.M(1))
			<-tick.C
		}
	}()

	// 2. Simulating RetrieveData 900ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(900 * time.Nanosecond)
		defer tick.Stop()

		for {
			select {
			case <-doneCh:
				return
			case <-tick.C:
				RetrieveData(view1.Name)
			}
		}
	}()

	// 4. Export via Reader routine at 800ns
	go func() {
		defer waiter.Done()
		tick := time.NewTicker(800 * time.Nanosecond)
		defer tick.Stop()

		reader := metricexport.Reader{}
		for {
			select {
			case <-doneCh:
				return
			case <-tick.C:
				// Perform some collection here
				reader.ReadAndExport(&testExporter{})
			}
		}
	}()
}

type testExporter struct {
	metrics []*metricdata.Metric
}

func (te *testExporter) ExportMetrics(ctx context.Context, metrics []*metricdata.Metric) error {
	te.metrics = metrics
	return nil
}

type countExporter struct {
	sync.Mutex
	count      int64
	totalCount int64
}

func (e *countExporter) ExportView(vd *Data) {
	if len(vd.Rows) == 0 {
		return
	}
	d := vd.Rows[0].Data.(*CountData)

	e.Lock()
	defer e.Unlock()
	e.count = d.Value
	e.totalCount += d.Value
}

type vdExporter struct {
	sync.Mutex
	vds []*Data
}

func (e *vdExporter) ExportView(vd *Data) {
	e.Lock()
	defer e.Unlock()

	e.vds = append(e.vds, vd)
}

// restart stops the current processors and creates a new one.
func restart() {
	defaultWorker.Stop()
	defaultWorker = NewMeter().(*worker)
	go defaultWorker.start()
}

// byTag implements sort.Interface for *metricdata.TimeSeries by Labels.
type byLabel []*metricdata.TimeSeries

func (ts byLabel) Len() int      { return len(ts) }
func (ts byLabel) Swap(i, j int) { ts[i], ts[j] = ts[j], ts[i] }
func (ts byLabel) Less(i, j int) bool {
	if len(ts[i].LabelValues) != len(ts[j].LabelValues) {
		return len(ts[i].LabelValues) < len(ts[j].LabelValues)
	}
	for k := range ts[i].LabelValues {
		if ts[i].LabelValues[k].Value != ts[j].LabelValues[k].Value {
			return ts[i].LabelValues[k].Value < ts[j].LabelValues[k].Value
		}
	}
	return false
}
