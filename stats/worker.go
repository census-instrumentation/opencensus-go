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

package stats

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opencensus.io/stats/aggregation"
	"go.opencensus.io/stats/internal"
	"go.opencensus.io/tag"
)

var (
	viewsMu sync.Mutex
	views   = make(map[string]*View)
)

func init() {
	go internal.DefaultWorker.Start(func(viewName string, rows []*internal.Row, start, end time.Time) {
		viewsMu.Lock()
		view := views[viewName]
		viewsMu.Unlock()

		if view == nil {
			return
		}
		r := make([]*Row, 0, len(rows))
		for _, rr := range rows {
			r = append(r, &Row{
				Tags: rr.Tags,
				Data: rr.Data,
			})
		}
		viewData := &ViewData{
			View:  view,
			Start: start,
			End:   time.Now(),
			Rows:  r,
		}
		exportersMu.Lock()
		for e := range exporters {
			e.Export(viewData)
		}
		exportersMu.Unlock()
	})
}

// FindView returns a registered view associated with this name.
// If no registered view is found, nil is returned.
func FindView(name string) (v *View) {
	viewsMu.Lock()
	defer viewsMu.Unlock()

	return views[name]
}

// RegisterView registers view. It returns an error if the view is already registered.
//
// Subscription automatically registers a view.
// Most users will not register directly but register via subscription.
// Registeration can be used by libraries to claim a view name.
//
// Unregister the view once the view is not required anymore.
func RegisterView(v *View) error {
	viewsMu.Lock()
	defer viewsMu.Unlock()

	name := v.Name()
	if _, ok := views[name]; ok {
		return fmt.Errorf("view %q is already registered", name)
	}

	views[name] = v
	return nil
}

// UnregisterView removes the previously registered view. It returns an error
// if the view wasn't registered. All data collected and not reported for the
// corresponding view will be lost. The view is automatically be unsubscribed.
func UnregisterView(v *View) error {
	if err := internal.DefaultWorker.Unsubscribe(v.Name()); err != nil {
		return err
	}
	viewsMu.Lock()
	delete(views, v.Name())
	viewsMu.Unlock()
	return nil
}

// Subscribe subscribes a view. Once a view is subscribed, it reports data
// via the exporters.
// During subscription, if the view wasn't registered, it will be automatically
// registered. Once the view is no longer needed to export data,
// user should unsubscribe from the view.
func (v *View) Subscribe() error {
	if err := RegisterView(v); err != nil {
		return err
	}
	wa, isCum := newWindowAggregator(v)
	m := v.Measure()
	return internal.DefaultWorker.Subscribe(&internal.ViewWorker{
		Name:    v.Name(),
		TagKeys: v.TagKeys(),
		MeasureDesc: &internal.MeasureDesc{
			Name: m.Name(),
		},
		NewData:             newData(v),
		NewWindowAggregator: wa,
		IsCumulative:        isCum,
	})
}

// Unsubscribe unsubscribes a previously subscribed channel.
// Data will not be exported from this view once unsubscription happens.
// If no more subscriber for v exists and the the ad hoc
// collection for this view isn't active, data stops being collected for this
// view.
func (v *View) Unsubscribe() error {
	return internal.DefaultWorker.Unsubscribe(v.Name())
}

// TODO(jbd): Implement func (v *View) RetrieveData() ([]*Row, error).

// Record records one or multiple measurements with the same tags at once.
// If there are any tags in the context, measurements will be tagged with them.
func Record(ctx context.Context, ms ...Measurement) {
	vals := make([]internal.Measurement, 0, len(ms))
	for _, m := range ms {
		vals = append(vals, internal.Measurement{
			MeasureName: m.m.Name(),
			Value:       m.v,
		})
	}
	internal.DefaultWorker.Record(time.Now(), tag.FromContext(ctx), vals)
}

// SetReportingPeriod sets the interval between reporting aggregated views in
// the program. If duration is less than or
// equal to zero, it enables the default behavior.
func SetReportingPeriod(d time.Duration) {
	internal.DefaultWorker.SetReportingDuration(d)
}

func newData(v *View) func() aggregation.Data {
	switch agg := v.Aggregation().(type) {
	case *CountAggregation:
		return func() aggregation.Data {
			return aggregation.NewCountData(0)
		}
	case CountAggregation:
		return func() aggregation.Data {
			return aggregation.NewCountData(0)
		}

	case SumAggregation:
		return func() aggregation.Data {
			return aggregation.NewSumData(0)
		}
	case *SumAggregation:
		return func() aggregation.Data {
			return aggregation.NewSumData(0)
		}

	case *MeanAggregation:
		return func() aggregation.Data {
			return aggregation.NewMeanData(0, 0)
		}
	case MeanAggregation:
		return func() aggregation.Data {
			return aggregation.NewMeanData(0, 0)
		}

	case DistributionAggregation:
		return func() aggregation.Data {
			return aggregation.NewDistributionData(agg)
		}
	case *DistributionAggregation:
		return func() aggregation.Data {
			return aggregation.NewDistributionData(*agg)
		}
	}
	return nil
}

func newWindowAggregator(v *View) (agg func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator, isCum bool) {
	switch w := v.Window().(type) {
	case *Cumulative:
		return func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator {
			return aggregation.NewCumulativeAggregator(start, fn)
		}, true
	case Cumulative:
		return func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator {
			return aggregation.NewCumulativeAggregator(start, fn)
		}, true
	case *Interval:
		return func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator {
			return aggregation.NewIntervalAggregator(w.Duration, w.Intervals, start, fn)
		}, false
	case Interval:
		return func(start time.Time, fn func() aggregation.Data) aggregation.WindowAggregator {
			return aggregation.NewIntervalAggregator(w.Duration, w.Intervals, start, fn)
		}, false
	}
	panic("unknown window")
}
