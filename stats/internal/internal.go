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

package internal

import (
	"fmt"
	"time"

	"go.opencensus.io/stats/aggregation"
	"go.opencensus.io/tag"
)

var DefaultWorker = newWorker()

var defaultReportingDuration = 10 * time.Second

type MeasureDesc struct {
	Name string
}

type measureRef struct {
	views map[*ViewWorker]struct{}
}

type command interface {
	handle(w *Worker)
}

type Worker struct {
	measures      map[string]*measureRef
	subscriptions map[string]*ViewWorker
	startTimes    map[*ViewWorker]time.Time

	c          chan command
	timer      *time.Ticker
	quit, done chan bool
}

func newWorker() *Worker {
	return &Worker{
		measures:      make(map[string]*measureRef),
		subscriptions: make(map[string]*ViewWorker),
		startTimes:    make(map[*ViewWorker]time.Time),
		c:             make(chan command),
		timer:         time.NewTicker(defaultReportingDuration),
		quit:          make(chan bool),
		done:          make(chan bool),
	}
}

type OnCollect func(viewName string, rows []*Row, start, end time.Time)

func (w *Worker) Start(o OnCollect) {
	for {
		select {
		case cmd := <-w.c:
			if cmd != nil {
				cmd.handle(w)
			}
		case <-w.timer.C:
			w.reportUsage(time.Now(), o)
		case <-w.quit:
			w.timer.Stop()
			close(w.c)
			w.done <- true
			return
		}
	}
}

func (w *Worker) Stop() {
	w.quit <- true
	<-w.done
}

func (w *Worker) reportUsage(reportTime time.Time, o OnCollect) {
	for _, w := range w.subscriptions {
		rows := w.collectedRows(reportTime)
		// if v.IsCumulative {
		// 	s, ok := w.startTimes[v]
		// 	if !ok {
		// 		w.startTimes[v] = start
		// 	} else {
		// 		start = s
		// 	}
		// }
		// Make sure collector is never going
		// to mutate the exported data.
		rows = deepCopyRowData(rows)
		o(w.Name, rows, reportTime, time.Now())
		if !w.IsCumulative {
			w.clearRows()
		}
	}
}

type setReportingDurationCmd struct {
	dur    time.Duration
	result chan error
}

func (cmd *setReportingDurationCmd) handle(w *Worker) {
	w.timer.Stop()
	if cmd.dur <= 0 {
		w.timer = time.NewTicker(defaultReportingDuration)
	} else {
		w.timer = time.NewTicker(cmd.dur)
	}
	cmd.result <- nil
}

func (w *Worker) SetReportingDuration(d time.Duration) {
	cmd := &setReportingDurationCmd{
		dur:    d,
		result: make(chan error),
	}
	w.c <- cmd
	<-cmd.result
}

func (w *Worker) registerMeasureDesc(desc *MeasureDesc) error {
	name := desc.Name
	if _, ok := w.measures[name]; ok {
		// the measure is already registered so there is nothing to do and the
		// command is considered successful.
		return nil
	}
	w.measures[name] = &measureRef{
		views: make(map[*ViewWorker]struct{}),
	}
	return nil
}

type deleteMeasureDescCmd struct {
	name   string
	result chan error
}

func (cmd *deleteMeasureDescCmd) handle(w *Worker) {
	ref, ok := w.measures[cmd.name]
	if !ok {
		cmd.result <- nil
		return
	}
	if c := len(ref.views); c > 0 {
		cmd.result <- fmt.Errorf("cannot delete; measure %q used by %v registered views", cmd.name, c)
		return
	}
	delete(w.measures, cmd.name)
	cmd.result <- nil
}

func (w *Worker) DeleteMeasure(name string) error {
	cmd := &deleteMeasureDescCmd{
		name:   name,
		result: make(chan error),
	}
	w.c <- cmd
	return <-cmd.result
}

type subscribeViewCmd struct {
	worker *ViewWorker
	result chan error
}

func (cmd *subscribeViewCmd) handle(w *Worker) {
	v := cmd.worker
	if _, ok := w.subscriptions[v.Name]; ok {
		cmd.result <- fmt.Errorf("cannot register view %q; another view with the same name is already registered", v.Name)
		return
	}
	if err := w.registerMeasureDesc(v.MeasureDesc); err != nil {
		cmd.result <- fmt.Errorf("cannot register view %q: %v", v.Name, err)
		return
	}
	v.signatures = make(map[string]aggregation.WindowAggregator)
	v.startTime = time.Now()
	w.subscriptions[v.Name] = v
	ref := w.measures[v.MeasureDesc.Name]
	ref.views[v] = struct{}{}
	cmd.result <- nil
}

func (w *Worker) Subscribe(worker *ViewWorker) error {
	cmd := &subscribeViewCmd{
		worker: worker,
		result: make(chan error),
	}
	w.c <- cmd
	return <-cmd.result
}

type unsubscribeViewDescCmd struct {
	name   string
	result chan error
}

func (cmd *unsubscribeViewDescCmd) handle(w *Worker) {
	worker, ok := w.subscriptions[cmd.name]
	if !ok {
		cmd.result <- nil
		return
	}
	delete(w.subscriptions, cmd.name)
	ref := w.measures[worker.MeasureDesc.Name]
	delete(ref.views, worker)
	worker.clearRows()
	cmd.result <- nil
}

func (w *Worker) Unsubscribe(name string) error {
	cmd := &unsubscribeViewDescCmd{
		name:   name,
		result: make(chan error),
	}
	w.c <- cmd
	return <-cmd.result
}

func (w *Worker) RetrieveData(name string) {

}

type Measurement struct {
	MeasureName string
	Value       interface{}
}

type recordCmd struct {
	now time.Time
	tm  *tag.Map
	ms  []Measurement
}

func (cmd *recordCmd) handle(w *Worker) {
	for _, m := range cmd.ms {
		ref := w.measures[m.MeasureName]
		for v := range ref.views {
			// TODO(jbd): Buffer sample.
			v.addSample(cmd.tm, m.Value, cmd.now)
		}
	}
}

func (w *Worker) Record(now time.Time, m *tag.Map, ms []Measurement) {
	cmd := &recordCmd{
		now: now,
		tm:  m,
		ms:  ms,
	}
	w.c <- cmd
}

func deepCopyRowData(rows []*Row) []*Row {
	newRows := make([]*Row, 0, len(rows))
	for _, r := range rows {
		newRows = append(newRows, &Row{
			Data: r.Data.Clone(),
			Tags: r.Tags,
		})
	}
	return newRows
}
