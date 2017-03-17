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

package stats

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/google/instrumentation-go/stats/tagging"
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

func (w *channelWorker) registerMeasureDesc(md MeasureDesc) error {
	if glog.V(3) {
		glog.Infof("registerMeasureDesc(_) registered MeasureDesc %v", md)
	}
	mr := &measureDescRegistration{
		md:  md,
		err: make(chan error),
	}
	w.inputs <- mr
	return <-mr.err
}

func (w *channelWorker) unregisterMeasureDesc(mName string) error {
	if glog.V(3) {
		glog.Infof("unregisterMeasureDesc(_) unregistered name %v", mName)
	}
	mu := &measureDescUnregistration{
		mn:  mName,
		err: make(chan error),
	}
	w.inputs <- mu
	return <-mu.err
}

func (w *channelWorker) registerViewDesc(vd ViewDesc) error {
	if glog.V(3) {
		glog.Infof("registerViewDesc(_) registered ViewDesc %v", vd.String())
	}
	vr := &viewDescRegistration{
		vd:  vd,
		err: make(chan error),
	}
	w.inputs <- vr
	return <-vr.err
}

func (w *channelWorker) unregisterViewDesc(vwName string) error {
	if glog.V(3) {
		glog.Infof("unregisterViewDesc(_) unregistered name %v", vwName)
	}
	vu := &viewDescUnregistration{
		vn:  vwName,
		err: make(chan error),
	}
	w.inputs <- vu
	return <-vu.err
}

func (w *channelWorker) subscribe(s Subscription) error {
	if glog.V(3) {
		glog.Infof("subscribeToViewDesc(_) with %v", s.String())
	}
	vs := &viewDescSubscription{
		s:   s,
		err: make(chan error),
	}
	w.inputs <- vs
	return <-vs.err
}

func (w *channelWorker) unsubscribe(s Subscription) error {
	if glog.V(3) {
		glog.Infof("unsubscribeFromViewDesc(_) with %v", s)
	}
	vu := &viewDescUnsubscription{
		s:   s,
		err: make(chan error),
	}
	w.inputs <- vu
	return <-vu.err
}

func (w *channelWorker) recordMeasurement(ctx context.Context, m Measurement) {
	ts := tagging.FromContext(ctx)

	w.inputs <- &singleRecord{
		ts: ts,
		m:  m,
	}
}

func (w *channelWorker) recordManyMeasurement(ctx context.Context, ms ...Measurement) {
	ts := tagging.FromContext(ctx)

	w.inputs <- &multiRecords{
		ts: ts,
		ms: ms,
	}
}

func (w *channelWorker) changeCallbackPeriod(min time.Duration, max time.Duration) {
	if glog.V(3) {
		glog.Infof("changeCallbackPeriod(_) min: %v, max: %v", min, max)
	}
	rf := &reportingPeriod{min, max}
	w.inputs <- rf
}

func (w *channelWorker) registerMeasureDescHandler(md MeasureDesc) error {
	return w.collector.registerMeasureDesc(md)
}

func (w *channelWorker) unregisterMeasureDescHandler(mName string) error {
	return w.collector.unregisterMeasureDesc(mName)
}

func (w *channelWorker) registerViewDescHandler(vd ViewDesc) error {
	return w.collector.registerViewDesc(vd, time.Now())
}

func (w *channelWorker) unregisterViewDescHandler(vwName string) error {
	return w.collector.unregisterViewDesc(vwName)
}

func (w *channelWorker) subscribeHandler(s Subscription) error {
	return w.collector.addSubscription(s)
}

func (w *channelWorker) unsubscribeHandler(s Subscription) error {
	return w.collector.unsubscribe(s)
}

func (w *channelWorker) recordMeasurementHandler(sr *singleRecord) {
	if err := w.collector.recordMeasurement(time.Now(), sr.ts, sr.m); err != nil {
		// TODO(iamm2): log that measureDesc is not registered.
		return
	}
	w.tryReportUsageIfMemoryUsageTooHigh()
}

func (w *channelWorker) recordMultiMeasurementHandler(mr *multiRecords) {
	if err := w.collector.recordManyMeasurement(time.Now(), mr.ts, mr.ms); err != nil {
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
	if glog.V(3) {
		glog.Info("reportUsage(_) started")
	}
	now := time.Now()
	views := w.collector.retrieveViews(now)
	if glog.V(3) {
		glog.Infof("reportUsage(_) %v views retrieved", len(views))
	}

	for _, vw := range views {
		vdc := vw.ViewDesc.ViewDescCommon()
		if glog.V(3) {
			glog.Infof("reportUsage(_) %v view's subscriptions %v", vdc.Name, len(vdc.subscriptions))
		}
		for subscription := range vdc.subscriptions {
			subscription.addView(vw)
		}
	}

	for s := range w.collector.subscriptions {
		s.reportUsage()
	}

	w.lastReportingTime = now
	if glog.V(3) {
		glog.Info("reportUsage(_) completed")
	}
}

func (w *channelWorker) resetTimer(d time.Duration) {
	w.maxWaitTimer.Stop()
	w.maxWaitTimer = time.NewTicker(d)
}

func newChannelWorker() *channelWorker {
	cw := &channelWorker{
		collector:         newUsageCollector(),
		lastReportingTime: time.Now(),
		reportingPeriod: &reportingPeriod{
			max: 10 * time.Second,
			min: 10 * time.Second,
		},
		inputs:       make(chan interface{}, 8192),
		maxWaitTimer: time.NewTicker(10 * time.Second),
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
					cmd.err <- cw.registerViewDescHandler(cmd.vd)
				case *viewDescUnregistration:
					cmd.err <- cw.unregisterViewDescHandler(cmd.vn)
				case *viewDescSubscription:
					cmd.err <- cw.subscribeHandler(cmd.s)
				case *viewDescUnsubscription:
					cmd.err <- cw.unsubscribeHandler(cmd.s)
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
	Subscribe = cw.subscribe
	Unsubscribe = cw.unsubscribe
	RecordMeasurement = cw.recordMeasurement
	RecordMeasurements = cw.recordManyMeasurement
	SetCallbackPeriod = cw.changeCallbackPeriod
	// TODO(acetechnologist): RetrieveViews =cw.retrieveViews
}
