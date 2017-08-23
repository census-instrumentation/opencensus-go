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
	"fmt"
	"time"

	"golang.org/x/net/context"
)

// maximumSizeBeforeReporting is the maximum number of records that can be
// recorded before exporting recorded usages.
// TODO(iamm2): make this dependent on the aggregated size, not the count of
// measures being recorded.
const maximumSizeBeforeReporting int64 = 5 * 1000 * 1000

// channelWorker is the private state of a single census worker goroutine. It
// is expected to only be a single such census worker goroutine per process.
// All communication with external consumers of census package is done through
// channels exclusively.
// REVIEWERS: If this becomes a bottleneck, then recordUsage can be revisited
// for optimization. It is also straightforward to add more workers and a
// dispatcher could be put in front of them.
type channelWorker struct {
	collector            *usageCollector
	viewMeasurementsSize int64
	lastReportingTime    time.Time
	reportingPeriod      *reportingPeriod
	inputs               chan interface{}
	// TODO(iamm2) add a quit channel for graceful shutdown.
	maxWaitTimer *time.Ticker
}

func (w *channelWorker) registerMeasureDesc(md *MeasureDesc) error {
	mr := &measureDescRegistration{
		md:  md,
		err: make(chan error),
	}
	w.inputs <- mr
	return <-mr.err
}

func (w *channelWorker) unregisterMeasureDesc(mName string) error {
	mu := &measureDescUnregistration{
		mn:  mName,
		err: make(chan error),
	}
	w.inputs <- mu
	return <-mu.err
}

func (w *channelWorker) registerViewDesc(avd AggregationViewDesc, c chan *View) error {
	vd := avd.viewDesc()
	vd.vChans = make(map[chan *View]struct{})
	vd.vChans[c] = struct{}{}
	vr := &viewDescRegistration{
		avd: avd,
		err: make(chan error),
	}
	w.inputs <- vr
	return <-vr.err
}

func (w *channelWorker) unregisterViewDesc(vwName string) error {
	vu := &viewDescUnregistration{
		vn:  vwName,
		err: make(chan error),
	}
	w.inputs <- vu
	return <-vu.err
}

func (w *channelWorker) subscribeToViewDesc(vn string, c chan *View) error {
	vs := &viewDescSubscription{
		vn:  vn,
		c:   c,
		err: make(chan error),
	}
	w.inputs <- vs
	return <-vs.err
}

func (w *channelWorker) unsubscribeFromViewDesc(vn string, c chan *View) error {
	vu := &viewDescUnsubscription{
		vn:  vn,
		c:   c,
		err: make(chan error),
	}
	w.inputs <- vu
	return <-vu.err
}

func (w *channelWorker) recordMeasurement(ctx context.Context, md *MeasureDesc, value float64) {
	ct := ctx.Value(censusKey{})
	if ct == nil {
		ct = make(contextTags)
	}

	w.inputs <- &singleRecord{
		ct: ct.(contextTags),
		v:  value,
		md: md,
	}
}

func (w *channelWorker) recordManyMeasurement(ctx context.Context, mds []*MeasureDesc, values []float64) {
	ct := ctx.Value(censusKey{})
	if ct == nil {
		ct = make(contextTags)
	}

	w.inputs <- &multiRecords{
		ct:  ct.(contextTags),
		vs:  values,
		mds: mds,
	}
}

func (w *channelWorker) changeCallbackPeriod(min time.Duration, max time.Duration) {
	rf := &reportingPeriod{min, max}
	w.inputs <- rf
}

func (w *channelWorker) registerMeasureDescHandler(md *MeasureDesc) error {
	return w.collector.registerMeasureDesc(md)
}

func (w *channelWorker) unregisterMeasureDescHandler(mName string) error {
	return w.collector.unregisterMeasureDesc(mName)
}

func (w *channelWorker) registerViewDescHandler(avd AggregationViewDesc) error {
	return w.collector.registerViewDesc(avd, time.Now())
}

func (w *channelWorker) unregisterViewDescHandler(vwName string) error {
	return w.collector.unregisterViewDesc(vwName)
}

func (w *channelWorker) subscribeToViewDescHandler(vwName string, c chan *View) error {
	return w.collector.subscribeToViewDesc(vwName, c)
}

func (w *channelWorker) unsubscribeFromViewDescHandler(vwName string, c chan *View) error {
	return w.collector.unsubscribeFromViewDesc(vwName, c)
}

func (w *channelWorker) recordMeasurementHandler(sr *singleRecord) {
	if err := w.collector.recordMeasurement(time.Now(), sr.ct, sr.md, sr.v); err != nil {
		// TODO(iamm2): log that measureDesc is not registered.
		return
	}
	w.tryReportUsageIfMemoryUsageTooHigh()
}

func (w *channelWorker) recordMultiMeasurementHandler(mr *multiRecords) {
	if err := w.collector.recordManyMeasurement(time.Now(), mr.ct, mr.mds, mr.vs); err != nil {
		// TODO(iamm2): log that measureDesc is not registered.
		return
	}
	w.tryReportUsageIfMemoryUsageTooHigh()
}

func (w *channelWorker) changeCallbackPeriodHandler(rp *reportingPeriod) {
	w.reportingPeriod = rp
	if w.reportingPeriod.max > 0 && w.reportingPeriod.min > 0 {
		w.resetTimer(w.reportingPeriod.max)
		return
	}
	w.maxWaitTimer.Stop()
}

// maxWaitTimeElapsedHandler is called after w.reportingPeriod.max has elapsed.
// It is possible that in the meantime the memory usage became too high and the
// metrics were reported before this duration elapsed. The metrics must be
// reported after at least w.reportingPeriod.min has elapsed since the last
// report.
func (w *channelWorker) maxWaitTimeElapsedHandler() {
	w.tryReportUsageIfMinTimeElapsed()
}

func (w *channelWorker) tryReportUsageIfMemoryUsageTooHigh() {
	// TODO(iamm2): implement safety code to deal with high memory usage.
	// In this case it should try to report the metrics to free
	// some memory. But it can only do so if at least w.reportingPeriod.min
	// elapsed since the last report.
	// if w.viewMeasurementsSize <= maximumSizeBeforeReporting {
	//	 return
	//}
	//w.tryReportUsageIfMinTimeElapsed()
}

func (w *channelWorker) tryReportUsageIfMinTimeElapsed() {
	sinceLastReport := time.Since(w.lastReportingTime)
	if sinceLastReport < w.reportingPeriod.min {
		//TODO(iamm2): if the library gets a spike of recordMeasurement and
		// we reach our memory limit (TBD), before the w.reportingPeriod.min
		// elapsed then this reset will be called repeatedly.
		// How to optimize this?
		w.resetTimer(w.reportingPeriod.min - sinceLastReport)
		return
	}
	w.reportUsage()
	w.resetTimer(w.reportingPeriod.max)
}

func (w *channelWorker) reportUsage() {
	now := time.Now()
	views := w.collector.retrieveViews(now)

	for _, vw := range views {
		for c := range vw.ViewDesc.vChans {
			select {
			case c <- vw:
			default:
				// TODO(iamm2) log data was dropped
			}
		}
	}
	w.lastReportingTime = now
}

func (w *channelWorker) resetTimer(d time.Duration) {
	w.maxWaitTimer.Stop()
	w.maxWaitTimer = time.NewTicker(d)
}

func newChannelWorker() *channelWorker {
	cw := &channelWorker{
		collector: &usageCollector{
			mDescriptors: make(map[string]*MeasureDesc),
			vDescriptors: make(map[string]AggregationViewDesc),
		},
		inputs:       make(chan interface{}, 8192),
		maxWaitTimer: time.NewTicker(24 * time.Hour),
	}

	go func() {
		for {
			select {
			case i := <-cw.inputs:
				switch cmd := i.(type) {
				case *measureDescRegistration:
					cmd.err <- cw.registerMeasureDescHandler(cmd.md)
				case *measureDescUnregistration:
					cmd.err <- cw.unregisterMeasureDescHandler(cmd.mn)
				case *viewDescRegistration:
					cmd.err <- cw.registerViewDescHandler(cmd.avd)
				case *viewDescUnregistration:
					cmd.err <- cw.unregisterViewDescHandler(cmd.vn)
				case *viewDescSubscription:
					cmd.err <- cw.subscribeToViewDescHandler(cmd.vn, cmd.c)
				case *viewDescUnsubscription:
					cmd.err <- cw.unsubscribeFromViewDescHandler(cmd.vn, cmd.c)
				case *reportingPeriod:
					cw.changeCallbackPeriodHandler(cmd)
				case *multiRecords:
					cw.recordMultiMeasurementHandler(cmd)
				case *singleRecord:
					cw.recordMeasurementHandler(cmd)
				default:
					panic(fmt.Sprintf("Unexpected command %v", cmd))
				}
			case <-cw.maxWaitTimer.C:
				cw.maxWaitTimeElapsedHandler()
			}

		}
	}()

	return cw
}

// init initializes the single channel worker and starts the background
// processing loop that reads messages from all the queues and processes them.
func init() {
	cw := newChannelWorker()
	RegisterMeasureDesc = cw.registerMeasureDesc
	UnregisterMeasureDesc = cw.unregisterMeasureDesc
	RegisterViewDesc = cw.registerViewDesc
	UnregisterViewDesc = cw.unregisterViewDesc
	SubscribeToView = cw.subscribeToViewDescHandler
	UnsubscribeFromView = cw.unsubscribeFromViewDescHandler
	RecordMeasurement = cw.recordMeasurement
	RecordManyMeasurement = cw.recordManyMeasurement
	SetCallbackPeriod = cw.changeCallbackPeriod
}
