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
	"fmt"
	"time"

	"github.com/google/working-instrumentation-go/tags"
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

// unregisterMeasureReq is the command to unregister a measure from the library.
type unregisterMeasureReq struct {
	m   Measure
	err chan error
}

func (cmd *unregisterMeasureReq) handleCommand(w *worker) {
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
		cmd.err <- fmt.Errorf("cannot unregister measure '%v'. All views referring to it must be unregistered first", cmd.m.Name())
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

	if v.subscriptionsCount() != 0 {
		cmd.err <- fmt.Errorf("cannot unregister view '%v'. All subscriptions to it must be unsubscribed first", cmd.v.Name())
		return
	}

	if v.isCollectingForAdhoc() {
		cmd.err <- fmt.Errorf("cannot unregister view '%v'. Its adhoc collection must be stopped first", cmd.v.Name())
		return
	}

	delete(w.viewsByName, cmd.v.Name())
	delete(w.views, cmd.v)
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

	if cmd.v.subscriptionsCount() == 0 || !cmd.v.isCollectingForAdhoc() {
		// this is the first subscription and isCollectingForAdhoc() is
		// disabled. Hence we need to start collecting data for this view. This
		// is done by adding it to the measure.
		cmd.v.measure().addView(cmd.v)
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

	if cmd.v.subscriptionsCount() == 0 && !cmd.v.isCollectingForAdhoc() {
		// this was the last subscription and isCollectingForAdhoc() is
		// disabled. Hence we need to stop collecting data for this view. This
		// is done by removing it from the measure.
		cmd.v.measure().removeView(cmd.v)
		cmd.v.clearRows()
	}

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// startCollectionForAdhocReq is the command to start collecting data for a
// view without subscribing to it.
type startCollectionForAdhocReq struct {
	v   View
	err chan error
}

func (cmd *startCollectionForAdhocReq) handleCommand(w *worker) {
	if err := w.tryRegisterView(cmd.v); err != nil {
		cmd.err <- fmt.Errorf("%v. Hence cannot start collection for adhoc", err)
		return
	}

	if cmd.v.subscriptionsCount() == 0 || !cmd.v.isCollectingForAdhoc() {
		// there are no subscriptions and isCollectingForAdhoc() is disabled.
		// Hence we need to start collecting data for this view. This is done
		// by adding it to the measure.
		cmd.v.measure().addView(cmd.v)
	}

	cmd.v.startCollectingForAdhoc()

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// stopCollectionForAdhocReq is the command to signal to the library that no
// more clients will be requesting data for a view. Has no impact on the
// subscriptions.
type stopCollectionForAdhocReq struct {
	v   View
	err chan error
}

func (cmd *stopCollectionForAdhocReq) handleCommand(w *worker) {
	cmd.v.stopCollectingForAdhoc()

	if cmd.v.subscriptionsCount() == 0 {
		// there are no subscriptions and isCollectingForAdhoc() is disabled.
		// Hence we need to stop collecting data for this view. This
		// is done by removing it from the measure.
		cmd.v.measure().removeView(cmd.v)
		cmd.v.clearRows()
	}

	// we always return nil because this operation never fails. However we
	// still need to return something on the channel to signal to the waiting
	// go routine that the operation completed.
	cmd.err <- nil
}

// retrieveDataReq is the command to retrieve data for a view.
type retrieveDataReq struct {
	v View
	c chan *retrieveDataResp
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

	if cmd.v.subscriptionsCount() == 0 && !cmd.v.isCollectingForAdhoc() {
		cmd.c <- &retrieveDataResp{
			nil,
			fmt.Errorf("cannot retrieve data for view with name '%v' because no client is subscribed to it and adhoc collection was not started for it explicitly", cmd.v.Name()),
		}
		return
	}

	cmd.c <- &retrieveDataResp{
		cmd.v.collectedRows(),
		nil,
	}
}

// recordFloat64Req is the command to record data related to a measure.
type recordFloat64Req struct {
	ts *tags.TagSet
	mf *MeasureFloat64
	v  float64
}

func (cmd *recordFloat64Req) handleCommand(w *worker) {
	if _, ok := w.measures[cmd.mf]; !ok {
		return
	}
	for v := range cmd.mf.views {
		v.addSample(cmd.ts, cmd.v)
	}
}

// recordInt64Req is the command to record data related to a measure.
type recordInt64Req struct {
	ts *tags.TagSet
	mi *MeasureInt64
	v  int64
}

func (cmd *recordInt64Req) handleCommand(w *worker) {
	if _, ok := w.measures[cmd.mi]; !ok {
		return
	}
	for v := range cmd.mi.views {
		v.addSample(cmd.ts, cmd.v)
	}
}

// recordReq is the command to record data related to multiple measures
// at once.
type recordReq struct {
	ts *tags.TagSet
	ms []Measurement
}

func (cmd *recordReq) handleCommand(w *worker) {
	for _, m := range cmd.ms {
		switch measurement := m.(type) {
		case *measurementFloat64:
			for v := range measurement.m.views {
				v.addSample(cmd.ts, measurement.v)
			}
		case *measurementInt64:
			for v := range measurement.m.views {
				v.addSample(cmd.ts, measurement.v)
			}
		default:
		}
	}
}

// setReportingPeriodReq is the command to modify the duration between
// reporting the collected data to the subscribed clients.
type setReportingPeriodReq struct {
	d time.Duration
}

func (cmd *setReportingPeriodReq) handleCommand(w *worker) {
	w.timer.Stop()
	if cmd.d <= 0*time.Second {
		w.timer = time.NewTicker(defaultReportingDuration)
		return
	}
	w.timer = time.NewTicker(cmd.d)
}
