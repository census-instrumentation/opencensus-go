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

// Package stats defines the stats collection API and its native Go
// implementation.

package stats2

import (
	"context"
	"time"
)

type worker struct {
	measuresByName map[string]Measure
	measures       map[Measure]bool
	viewsByName    map[string]View
	views          map[View]bool

	c chan command
}

func newWorker() *worker {
	return &worker{}
}

// GetMeasureByName returns the registered measure associated with name.
var GetMeasureByName func(name string) (Measure, error)

// RegisterMeasure registers a measure. It returns an error if a measure with
// the same name is already registered.
var RegisterMeasure func(m Measure) error

// UnregisterMeasure de-registers a measure. It returns an error if the measure
// is not already registered.
var UnregisterMeasure func(m Measure) error

// GetViewByName returns the registered view associated with this name.
var GetViewByName func(name string) (View, error)

// RegisterView registers view. It returns an error if the view cannot be
// registered. Subsequent calls to Record with the same measure as the one in
// the view will NOT cause the usage to be recorded unless a consumer is
// subscribed to the view or StartCollectionForAdhoc for this view is called.
var RegisterView func(v View) error

// UnregisterView deletes the previously registered view. It returns an error
// if the view wasn't registered. All data collected and not reported for the
// corresponding view will be lost. All clients subscribed to this view are
// unsubscribed automatically and their subscriptions channels closed.
var UnregisterView func(v View) error

// SubscribeToView subscribes a client to a View. If the view wasn't already
// registered, it will be automatically registered. It allows for many clients
// to consume the same ViewData with a single registration. -i.e. the aggregate
// of the collected measurements will be reported to the calling code through
// channel c. To avoid data loss, clients must ensure that channel sends
// proceed in a timely manner. The calling code is responsible for using a
// buffered channel or blocking on the channel waiting for the collected data.
var SubscribeToView func(v View, c chan *ViewData) error

// UnsubscribeFromView unsubscribes a previously subscribed channel from the
// View subscriptions. If no more subscriber for v exists and the the ad hoc
// collection for this view isn't active, data stops being collected for this
// view.
var UnsubscribeFromView func(v View, c chan *ViewData) error

// StartCollectionForAdhoc starts data collection for this view even if no
// listeners are subscribed to it.
var StartCollectionForAdhoc func(v View) error

// StopCollectionForAdhoc stops data collection for this view unless at least
// 1 listener is subscribed to it.
var StopCollectionForAdhoc func(v View) error

// RetrieveData returns the current collected data for the view.
var RetrieveData func(v View) ([]*Rows, error)

// RecordFloat64 records a float64 value against a measure and the tags passed
// as part of the context.
var RecordFloat64 func(ctx context.Context, mf MeasureFloat64, v float64)

// RecordInt64 records an int64 value against a measure and the tags passed as
// part of the context.
var RecordInt64 func(ctx context.Context, mf MeasureInt64, v int64)

// Record records one or multiple measurements with the same tags at once.
var Record func(ctx context.Context, ms []Measurement)

// SetReportingPeriod sets the interval between reporting aggregated views in
// the program. Calling SetReportingPeriod with duration argument equal to zero
// enables the default behavior.
var SetReportingPeriod func(d time.Duration)

func init() {
	w := newWorker()
	GetMeasureByName = w.getMeasureByName
	RegisterMeasure = w.registerMeasure
	UnregisterMeasure = w.unregisterMeasure
	GetViewByName = w.getViewByName
	RegisterView = w.registerView
	UnregisterView = w.unregisterView
	SubscribeToView = w.subscribeToView
	UnsubscribeFromView = w.unsubscribeFromView
	StartCollectionForAdhoc = w.startCollectionForAdhoc
	StopCollectionForAdhoc = w.stopCollectionForAdhoc
	RetrieveData = w.retrieveData
	RecordFloat64 = w.recordFloat64
	RecordInt64 = w.recordInt64
	Record = w.record
	SetReportingPeriod = w.setReportingPeriod

	w.start()
}

func (w *worker) start() {
	for {
		cmd := <-w.c
		cmd.handleCommand(w)
	}
}

func (w *worker) getMeasureByName(name string) (Measure, error) {
	return nil, nil
}

func (w *worker) registerMeasure(m Measure) error {
	return nil
}

func (w *worker) unregisterMeasure(m Measure) error {
	return nil
}

func (w *worker) getViewByName(name string) (View, error) {
	return nil, nil
}

func (w *worker) registerView(v View) error {
	// if &view registered return true
	// if other view with same name  return false

	// if measure !registered
	//  if register(measure) == fail return false

	// registerview and return success
	return nil
}

func (w *worker) unregisterView(v View) error {
	return nil
}

func (w *worker) subscribeToView(v View, c chan *ViewData) error {
	// if view !registered
	// success = registerview
	// if fail return error
	//subscribe and return true
	return nil
}

func (w *worker) unsubscribeFromView(v View, c chan *ViewData) error {
	return nil
}

func (w *worker) startCollectionForAdhoc(v View) error {
	return nil
}

func (w *worker) stopCollectionForAdhoc(v View) error {
	return nil
}

func (w *worker) retrieveData(v View) ([]*Rows, error) {
	return nil, nil
}

func (w *worker) recordFloat64(ctx context.Context, mf MeasureFloat64, v float64) {

}

func (w *worker) recordInt64(ctx context.Context, mf MeasureInt64, v int64) {

}

func (w *worker) record(ctx context.Context, ms []Measurement) {

}

func (w *worker) setReportingPeriod(d time.Duration) {

}
