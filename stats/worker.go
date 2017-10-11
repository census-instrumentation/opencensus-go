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
	"errors"
	"fmt"
	"time"

	"github.com/census-instrumentation/opencensus-go/tags"
	"golang.org/x/net/context"
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

// MeasureByName returns the registered measure associated with name.
func MeasureByName(name string) (Measure, error) {
	req := &getMeasureByNameReq{
		name: name,
		c:    make(chan *getMeasureByNameResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.m, resp.err
}

// DeleteMeasure deletes an existing measure to allow for creation of a new
// measure with the same name. It returns an error if the measure cannot be
// deleted (if one or multiple registered views refer to it).
func DeleteMeasure(m Measure) error {
	req := &deleteMeasureReq{
		m:   m,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// FindView returns a registered view associated with this name.
func FindView(name string) (*View, error) {
	req := &getViewByNameReq{
		name: name,
		c:    make(chan *getViewByNameResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.v, resp.err
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
// corresponding view will be lost. All clients subscribed to this view are
// unsubscribed automatically and their subscriptions channels closed.
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

// Subscribe subscribes a channel to a View. If the view wasn't already
// registered, it will be automatically registered. It allows for many clients
// to consume the same ViewData with a single registration. -i.e. the aggregate
// of the collected measurements will be reported to the calling code through
// channel c. To avoid data loss, clients must ensure that channel sends
// proceed in a timely manner. The calling code is responsible for using a
// buffered channel or blocking on the channel waiting for the collected data.
func (v *View) Subscribe(c chan *ViewData) error {
	if v == nil {
		return errors.New("cannot subscribe nil view")
	}
	req := &subscribeToViewReq{
		v:   v,
		c:   c,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// Unsubscribe unsubscribes a previously subscribed channel from the
// View subscriptions. If no more subscriber for v exists and the the ad hoc
// collection for this view isn't active, data stops being collected for this
// view.
func (v *View) Unsubscribe(c chan *ViewData) error {
	if v == nil {
		return errors.New("cannot unsubscribe nil view")
	}
	req := &unsubscribeFromViewReq{
		v:   v,
		c:   c,
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

// RecordFloat64 records a float64 value against a measure and the tags passed
// as part of the context.
func RecordFloat64(ctx context.Context, mf *MeasureFloat64, v float64) {
	req := &recordFloat64Req{
		now: time.Now(),
		ts:  tags.FromContext(ctx),
		mf:  mf,
		v:   v,
	}
	defaultWorker.c <- req
}

// RecordInt64 records an int64 value against a measure and the tags passed as
// part of the context.
func RecordInt64(ctx context.Context, mi *MeasureInt64, v int64) {
	req := &recordInt64Req{
		now: time.Now(),
		ts:  tags.FromContext(ctx),
		mi:  mi,
		v:   v,
	}
	defaultWorker.c <- req
}

// Record records one or multiple measurements with the same tags at once.
func Record(ctx context.Context, ms ...Measurement) {
	req := &recordReq{
		now: time.Now(),
		ts:  tags.FromContext(ctx),
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
		if v.subscriptionsCount() == 0 {
			continue
		}

		viewData := &ViewData{
			V:    v,
			Rows: v.collectedRows(now),
		}

		for c, s := range v.subscriptions() {
			select {
			case c <- viewData:
				return
			default:
				s.droppedViewData++
			}
		}

		if _, ok := v.Window().(*WindowCumulative); !ok {
			v.clearRows()
		}
	}
}

// RestartWorker is used for testing only. It stops the old worker and creates
// a new worker. It should never be called by production code.
func RestartWorker() {
	defaultWorker.stop()
	defaultWorker = newWorker()
	go defaultWorker.start()
}
