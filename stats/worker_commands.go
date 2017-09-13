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

// Package stats defines the stats collection API and its native Go
// implementation.
package stats

import (
	"fmt"
	"time"

	"github.com/census-instrumentation/opencensus-go/tags"
)

type command interface {
	handleCommand(w *worker)
}

// getMeasureByNameReq is the command to get a measure given its name.
type getMeasureByNameReq struct {
	name string
	c    chan *getMeasureByNameResp
}

type getMeasureByNameResp struct {
	m   Measure
	err error
}

func (cmd *getMeasureByNameReq) handleCommand(w *worker) {
	if m, ok := w.measuresByName[cmd.name]; ok {
		cmd.c <- &getMeasureByNameResp{
			m,
			nil,
		}
		return
	}
	cmd.c <- &getMeasureByNameResp{
		nil,
		fmt.Errorf("no measure named '%v' is registered", cmd.name),
	}
}

// registerMeasureReq is the command to register a measure with the library.
type registerMeasureReq struct {
	m   Measure
	err chan error
}

func (cmd *registerMeasureReq) handleCommand(w *worker) {
	cmd.err <- w.tryRegisterMeasure(cmd.m)
}

// deleteMeasureReq is the command to delete a measure from the library.
type deleteMeasureReq struct {
	m   Measure
	err chan error
}

func (cmd *deleteMeasureReq) handleCommand(w *worker) {
	m, ok := w.measuresByName[cmd.m.Name()]
	if !ok {
		cmd.err <- nil
		return
	}

	if m != cmd.m {
		cmd.err <- nil
		return
	}

	if m.viewsCount() != 0 {
		cmd.err <- fmt.Errorf("cannot delete measure '%v'. All views referring to it must be unregistered first", cmd.m.Name())
		return
	}

	delete(w.measuresByName, cmd.m.Name())
	delete(w.measures, cmd.m)
	cmd.err <- nil
}

// getViewByNameReq is the command to get a view given its name.
type getViewByNameReq struct {
	name string
	c    chan *getViewByNameResp
}

type getViewByNameResp struct {
	v   View
	err error
}

func (cmd *getViewByNameReq) handleCommand(w *worker) {
	if v, ok := w.viewsByName[cmd.name]; ok {
		cmd.c <- &getViewByNameResp{
			v,
			nil,
		}
		return
	}
	cmd.c <- &getViewByNameResp{
		nil,
		fmt.Errorf("no view named '%v' is registered", cmd.name),
	}
}

// registerViewReq is the command to register a view with the library.
type registerViewReq struct {
	v   View
	err chan error
}

func (cmd *registerViewReq) handleCommand(w *worker) {
	cmd.err <- w.tryRegisterView(cmd.v)
}

// unregisterViewReq is the command to unregister a view from the library.
type unregisterViewReq struct {
	v   View
	err chan error
}

func (cmd *unregisterViewReq) handleCommand(w *worker) {
	v, ok := w.viewsByName[cmd.v.Name()]
	if !ok {
		cmd.err <- nil
		return
	}

	if v != cmd.v {
		cmd.err <- nil
		return
	}

	if v.isCollecting() {
		cmd.err <- fmt.Errorf("cannot unregister view '%v'. All subscriptions to it must be unsubscribed and its forced collection must be stopped first", cmd.v.Name())
		return
	}

	delete(w.viewsByName, cmd.v.Name())
	delete(w.views, cmd.v)
	cmd.v.Measure().removeView(v)
	cmd.err <- nil
}

// subscribeToViewReq is the command to subscribe to a view.
type subscribeToViewReq struct {
	v   View
	c   chan *ViewData
	err chan error
}

func (cmd *subscribeToViewReq) handleCommand(w *worker) {
	if cmd.v.subscriptionExists(cmd.c) {
		cmd.err <- nil
		return
	}
	if err := w.tryRegisterView(cmd.v); err != nil {
		cmd.err <- fmt.Errorf("%v. Hence cannot subscribe to channel", err)
		return
	}

	cmd.v.addSubscription(cmd.c)

	cmd.err <- nil
}

// unsubscribeFromViewReq is the command to unsubscribe to a view. Has no
// impact on the data collection for client that are pulling data from the
// library.
type unsubscribeFromViewReq struct {
	v   View
	c   chan *ViewData
	err chan error
}

func (cmd *unsubscribeFromViewReq) handleCommand(w *worker) {
	cmd.v.deleteSubscription(cmd.c)

	if !cmd.v.isCollecting() {
		// this was the last subscription and view is not collecting anymore.
		// The collected data can be cleared.
		cmd.v.clearRows()
	}

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// startForcedCollection is the command to start collecting data for a view
// without subscribing to it.
type startForcedCollectionReq struct {
	v   View
	err chan error
}

func (cmd *startForcedCollectionReq) handleCommand(w *worker) {
	if err := w.tryRegisterView(cmd.v); err != nil {
		cmd.err <- fmt.Errorf("%v. Hence cannot start forced collection", err)
		return
	}

	cmd.v.startForcedCollection()

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// stopForcedCollectionReq is the command to signal to the library that no more
// clients will be requesting data for a view. Has no impact on the
// subscriptions.
type stopForcedCollectionReq struct {
	v   View
	err chan error
}

func (cmd *stopForcedCollectionReq) handleCommand(w *worker) {
	cmd.v.stopForcedCollection()

	if !cmd.v.isCollecting() {
		cmd.v.clearRows()
	}

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// retrieveDataReq is the command to retrieve data for a view.
type retrieveDataReq struct {
	now time.Time
	v   View
	c   chan *retrieveDataResp
}

type retrieveDataResp struct {
	rows []*Row
	err  error
}

func (cmd *retrieveDataReq) handleCommand(w *worker) {
	if _, ok := w.views[cmd.v]; !ok {
		cmd.c <- &retrieveDataResp{
			nil,
			fmt.Errorf("cannot retrieve data for view with name '%v' because it is not registered", cmd.v.Name()),
		}
		return
	}

	if !cmd.v.isCollecting() {
		cmd.c <- &retrieveDataResp{
			nil,
			fmt.Errorf("cannot retrieve data for view with name '%v' because no client is subscribed to it and its collection was not forcibly started", cmd.v.Name()),
		}
		return
	}
	cmd.c <- &retrieveDataResp{
		cmd.v.collectedRows(cmd.now),
		nil,
	}
}

// recordFloat64Req is the command to record data related to a measure.
type recordFloat64Req struct {
	now time.Time
	ts  *tags.TagSet
	mf  *MeasureFloat64
	v   float64
}

func (cmd *recordFloat64Req) handleCommand(w *worker) {
	if _, ok := w.measures[cmd.mf]; !ok {
		return
	}
	for v := range cmd.mf.views {
		v.addSample(cmd.ts, cmd.v, cmd.now)
	}
}

// recordInt64Req is the command to record data related to a measure.
type recordInt64Req struct {
	now time.Time
	ts  *tags.TagSet
	mi  *MeasureInt64
	v   int64
}

func (cmd *recordInt64Req) handleCommand(w *worker) {
	if _, ok := w.measures[cmd.mi]; !ok {
		return
	}
	for v := range cmd.mi.views {
		v.addSample(cmd.ts, cmd.v, cmd.now)
	}
}

// recordReq is the command to record data related to multiple measures
// at once.
type recordReq struct {
	now time.Time
	ts  *tags.TagSet
	ms  []Measurement
}

func (cmd *recordReq) handleCommand(w *worker) {
	for _, m := range cmd.ms {
		switch measurement := m.(type) {
		case *measurementFloat64:
			for v := range measurement.m.views {
				v.addSample(cmd.ts, measurement.v, cmd.now)
			}
		case *measurementInt64:
			for v := range measurement.m.views {
				v.addSample(cmd.ts, measurement.v, cmd.now)
			}
		default:
		}
	}
}

// setReportingPeriodReq is the command to modify the duration between
// reporting the collected data to the subscribed clients.
type setReportingPeriodReq struct {
	d time.Duration
	c chan bool
}

func (cmd *setReportingPeriodReq) handleCommand(w *worker) {
	w.timer.Stop()
	if cmd.d <= 0*time.Second {
		w.timer = time.NewTicker(defaultReportingDuration)
		return
	}
	w.timer = time.NewTicker(cmd.d)
	cmd.c <- true
}
