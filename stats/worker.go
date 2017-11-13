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

package stats

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opencensus.io/tag"
)

func init() {
	defaultWorker = newWorker()
	go defaultWorker.start()
}

type worker struct {
	measuresByName map[string]Measure
	measures       map[Measure]bool
	viewsByName    map[string]*View
	views          map[*View]bool

	timer      *time.Ticker
	c          chan command
	quit, done chan bool
}

var defaultWorker *worker

var defaultReportingDuration = 10 * time.Second

// NewMeasureFloat64 creates a new measure of type MeasureFloat64. It returns
// an error if a measure with the same name already exists.
func NewMeasureFloat64(name, description, unit string) (*MeasureFloat64, error) {
	m := &MeasureFloat64{
		name:        name,
		description: description,
		unit:        unit,
		views:       make(map[*View]bool),
	}

	req := &registerMeasureReq{
		m:   m,
		err: make(chan error),
	}
	defaultWorker.c <- req
	if err := <-req.err; err != nil {
		return nil, err
	}

	return m, nil
}

// NewMeasureInt64 creates a new measure of type MeasureInt64. It returns an
// error if a measure with the same name already exists.
func NewMeasureInt64(name, description, unit string) (*MeasureInt64, error) {
	m := &MeasureInt64{
		name:        name,
		description: description,
		unit:        unit,
		views:       make(map[*View]bool),
	}

	req := &registerMeasureReq{
		m:   m,
		err: make(chan error),
	}
	defaultWorker.c <- req
	if err := <-req.err; err != nil {
		return nil, err
	}

	return m, nil
}

// FindView returns a registered view associated with this name.
// If no registered view is found, ok is false.
func FindView(name string) (v *View, ok bool) {
	req := &getViewByNameReq{
		name: name,
		c:    make(chan *getViewByNameResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.v, resp.ok
}

// RegisterView registers view. It returns an error if the view cannot be
// registered. Subsequent calls to Record with the same measure as the one in
// the view will NOT cause the usage to be recorded unless a consumer is
// subscribed to the view or ForceCollect for this view is called.
func RegisterView(v *View) error {
	if v == nil {
		return errors.New("cannot RegisterView for nil view")
	}

	req := &registerViewReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// Unregister removes the previously registered view. It returns an error
// if the view wasn't registered. All data collected and not reported for the
// corresponding view will be lost. The view is automatically be unsubscribed.
func (v *View) Unregister() error {
	if v == nil {
		return errors.New("cannot UnregisterView for nil view")
	}
	req := &unregisterViewReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// Subscribe subscribes a view. Once a view is subscribed, it reports data
// via the exporters.
// During subscription, if the view wasn't registered, it will be automatically
// registered. Once the view is no longer needed to export data,
// user should unsubscribe from the view.
func (v *View) Subscribe() error {
	req := &subscribeToViewReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// Unsubscribe unsubscribes a previously subscribed channel.
// Data will not be exported from this view once unsubscription happens.
// If no more subscriber for v exists and the the ad hoc
// collection for this view isn't active, data stops being collected for this
// view.
func (v *View) Unsubscribe() error {
	req := &unsubscribeFromViewReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// ForceCollect starts data collection for this view even if no
// listeners are subscribed to it.
func (v *View) ForceCollect() error {
	if v == nil {
		return errors.New("cannot for collect nil view")
	}
	req := &startForcedCollectionReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// StopForceCollection stops data collection for this
// view unless at least 1 listener is subscribed to it.
func (v *View) StopForceCollection() error {
	if v == nil {
		return errors.New("cannot stop force collection for nil view")
	}
	req := &stopForcedCollectionReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// RetrieveData returns the current collected data for the view.
func (v *View) RetrieveData() ([]*Row, error) {
	if v == nil {
		return nil, errors.New("cannot retrieve data from nil view")
	}
	req := &retrieveDataReq{
		now: time.Now(),
		v:   v,
		c:   make(chan *retrieveDataResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.rows, resp.err
}

// Record records one or multiple measurements with the same tags at once.
// If there are any tags in the context, measurements will be tagged with them.
func Record(ctx context.Context, ms ...Measurement) {
	req := &recordReq{
		now: time.Now(),
		tm:  tag.FromContext(ctx),
		ms:  ms,
	}
	defaultWorker.c <- req
}

// SetReportingPeriod sets the interval between reporting aggregated views in
// the program. Calling SetReportingPeriod with duration argument less than or
// equal to zero enables the default behavior.
func SetReportingPeriod(d time.Duration) {
	// TODO(acetechnologist): ensure that the duration d is more than a certain
	// value. e.g. 1s
	req := &setReportingPeriodReq{
		d: d,
		c: make(chan bool),
	}
	defaultWorker.c <- req
	<-req.c // don't return until the timer is set to the new duration.
}

func newWorker() *worker {
	return &worker{
		measuresByName: make(map[string]Measure),
		measures:       make(map[Measure]bool),
		viewsByName:    make(map[string]*View),
		views:          make(map[*View]bool),
		timer:          time.NewTicker(defaultReportingDuration),
		c:              make(chan command),
		quit:           make(chan bool),
		done:           make(chan bool),
	}
}

func (w *worker) start() {
	for {
		select {
		case cmd := <-w.c:
			if cmd != nil {
				cmd.handleCommand(w)
			}
		case <-w.timer.C:
			w.reportUsage(time.Now())
		case <-w.quit:
			w.timer.Stop()
			close(w.c)
			w.done <- true
			return
		}
	}
}

func (w *worker) stop() {
	w.quit <- true
	<-w.done
}

func (w *worker) tryRegisterMeasure(m Measure) error {
	if x, ok := w.measuresByName[m.Name()]; ok {
		if x != m {
			return fmt.Errorf("cannot register measure %q; another measure with the same name is already registered", m.Name())
		}
		// the measure is already registered so there is nothing to do and the
		// command is considered successful.
		return nil
	}

	w.measuresByName[m.Name()] = m
	w.measures[m] = true
	return nil
}

func (w *worker) tryRegisterView(v *View) error {
	if x, ok := w.viewsByName[v.Name()]; ok {
		if x != v {
			return fmt.Errorf("cannot register view %q; another view with the same name is already registered", v.Name())
		}

		// the view is already registered so there is nothing to do and the
		// command is considered successful.
		return nil
	}

	// view is not registered and needs to be registered, but first its measure
	// needs to be registered.
	if err := w.tryRegisterMeasure(v.Measure()); err != nil {
		return fmt.Errorf("cannot register view %q: %v", v.Name(), err)
	}

	w.viewsByName[v.Name()] = v
	w.views[v] = true
	v.Measure().addView(v)
	return nil
}

func (w *worker) reportUsage(now time.Time) {
	for v := range w.views {
		if !v.isSubscribed() {
			continue
		}
		rows := v.collectedRows(now)
		viewData := &ViewData{
			View:  v,
			Start: now,
			End:   time.Now(),
			Rows:  rows,
		}
		exportersMu.Lock()
		for e := range exporters {
			e.Export(viewData)
		}
		exportersMu.Unlock()
		if _, ok := v.Window().(*CumulativeWindow); !ok {
			v.clearRows()
		}
	}
}
