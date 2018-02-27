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
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/internal"
	"go.opencensus.io/tag"
)

func init() {
	defaultWorker = newWorker()
	go defaultWorker.start()
	internal.DefaultRecorder = record
}

type measureRef struct {
	measure string
	views   map[*viewInternal]struct{}
}

type worker struct {
	measures   map[string]*measureRef
	views      map[string]*viewInternal
	startTimes map[*viewInternal]time.Time

	timer      *time.Ticker
	c          chan command
	quit, done chan bool
}

var defaultWorker *worker

var defaultReportingDuration = 10 * time.Second

// Find returns a subscribed view associated with this name.
// If no subscribed view is found, nil is returned.
func Find(name string) (v *View) {
	req := &getViewByNameReq{
		name: name,
		c:    make(chan *getViewByNameResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.v
}

// Deprecated: Registering is a no-op. Use the Subscribe function.
func Register(_ *View) error {
	return nil
}

// Deprecated: Unregistering is a no-op, see: Unsubscribe.
func Unregister(_ *View) error {
	return nil
}

// Deprecated: Use the Subscribe function.
func (v *View) Subscribe() error {
	req := &subscribeToViewReq{
		v:   v,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// Subscribe begins collecting data for the given views.
// Once a view is subscribed, it reports data to the registered exporters.
func Subscribe(views ...*View) error {
	var errstr []string
	for _, v := range views {
		err := v.Subscribe()
		if err != nil {
			errstr = append(errstr, fmt.Sprintf("%s: %v", v.Name, err))
		}
	}
	if len(errstr) > 0 {
		return errors.New(strings.Join(errstr, "\n"))
	}
	return nil
}

// Unsubscribe unsubscribes a previously subscribed view.
// Data will not be exported from this view once unsubscription happens.
func (v *View) Unsubscribe() error {
	req := &unsubscribeFromViewReq{
		v:   v.Name,
		err: make(chan error),
	}
	defaultWorker.c <- req
	return <-req.err
}

// RetrieveData returns the current collected data for the view.
// Deprecated: Use RetrieveData(name)
func (v *View) RetrieveData() ([]*Row, error) {
	if v == nil {
		return nil, errors.New("cannot retrieve data from nil view")
	}
	return RetrieveData(v.Name)
}

func RetrieveData(viewName string) ([]*Row, error) {
	req := &retrieveDataReq{
		now: time.Now(),
		v:   viewName,
		c:   make(chan *retrieveDataResp),
	}
	defaultWorker.c <- req
	resp := <-req.c
	return resp.rows, resp.err
}

func record(tags *tag.Map, ms interface{}) {
	req := &recordReq{
		tm: tags,
		ms: ms.([]stats.Measurement),
	}
	defaultWorker.c <- req
}

// SetReportingPeriod sets the interval between reporting aggregated views in
// the program. If duration is less than or
// equal to zero, it enables the default behavior.
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
		measures:   make(map[string]*measureRef),
		views:      make(map[string]*viewInternal),
		startTimes: make(map[*viewInternal]time.Time),
		timer:      time.NewTicker(defaultReportingDuration),
		c:          make(chan command, 1024),
		quit:       make(chan bool),
		done:       make(chan bool),
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

func (w *worker) getMeasureRef(name string) *measureRef {
	if mr, ok := w.measures[name]; ok {
		return mr
	}
	mr := &measureRef{
		measure: name,
		views:   make(map[*viewInternal]struct{}),
	}
	w.measures[name] = mr
	return mr
}

func (w *worker) tryRegisterView(v *View) (*viewInternal, error) {
	if err := checkViewName(v.Name); err != nil {
		return nil, err
	}
	sort.Slice(v.GroupByTags, func(i, j int) bool {
		return v.GroupByTags[i].Name() < v.GroupByTags[j].Name()
	})
	if x, ok := w.views[v.Name]; ok {
		if !x.definition.Equal(v) {
			return nil, fmt.Errorf("cannot subscribe view %q; a different view with the same name is already subscribed", v.Name)
		}

		// the view is already registered so there is nothing to do and the
		// command is considered successful.
		return x, nil
	}

	if v.MeasureName == "" {
		return nil, fmt.Errorf("cannot subscribe view %q: measure not set", v.Name)
	}
	m := stats.FindMeasure(v.MeasureName)
	if m == nil {
		return nil, fmt.Errorf("cannot subscribe view %q: measure %q not found", v.Name, v.MeasureName)
	}
	var agg = v.Aggregation
	if agg == nil {
		return nil, fmt.Errorf("cannot subscribe view %q: aggregation not set", v.Name)
	}
	vi := newViewInternal(v, m)
	w.views[v.Name] = vi
	ref := w.getMeasureRef(v.MeasureName)
	ref.views[vi] = struct{}{}

	return vi, nil
}

func (w *worker) reportUsage(now time.Time) {
	for _, v := range w.views {
		if !v.isSubscribed() {
			continue
		}
		rows := v.collectedRows()
		_, ok := w.startTimes[v]
		if !ok {
			w.startTimes[v] = now
		}
		// Make sure collector is never going
		// to mutate the exported data.
		rows = deepCopyRowData(rows)
		viewData := &Data{
			Measure: v.measure,
			View:    &v.definition,
			Start:   w.startTimes[v],
			End:     time.Now(),
			Rows:    rows,
		}
		exportersMu.Lock()
		for e := range exporters {
			e.ExportView(viewData)
		}
		exportersMu.Unlock()
	}
}

func deepCopyRowData(rows []*Row) []*Row {
	newRows := make([]*Row, 0, len(rows))
	for _, r := range rows {
		newRows = append(newRows, &Row{
			Data: r.Data.clone(),
			Tags: r.Tags,
		})
	}
	return newRows
}
