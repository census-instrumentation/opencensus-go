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

package metric

import (
	"context"
	"fmt"
	"log"
	"time"
)

// PushExporter exports metrics at regular intervals.
// Use Init to initialize a new PushExporter.
// Do not modify fields of PushExporter after calling Run.
type PushExporter struct {
	push          PushFunc
	registry      *Registry
	quit, didQuit chan struct{}

	// ReportingPeriod is the interval between exports. A new export will
	// be started every ReportingPeriod, even if the previous one is not
	// yet completed.
	// ReportingPeriod should be set higher than Timeout.
	ReportingPeriod time.Duration

	// OnError may be provided to customize the logging of errors returned
	// from push calls. By default, the error message is logged to the
	// standard error stream.
	OnError func(err error)

	// Timeout that will be applied to the context for each push call.
	Timeout time.Duration
}

// PushFunc is a function that exports metrics, for example to a monitoring
// backend.
type PushFunc func(context.Context, []*Metric) error

const (
	defaultReportingPeriod = 10 * time.Second
	defaultTimeout         = 5 * time.Second
)

// Init initializes a new PushExporter that reads
// metrics from the given registry and pushes them to the given PushFunc.
func (p *PushExporter) Init(reg *Registry, push PushFunc) {
	*p = PushExporter{
		push:            push,
		registry:        reg,
		quit:            make(chan struct{}),
		didQuit:         make(chan struct{}),
		ReportingPeriod: defaultReportingPeriod,
		OnError: func(err error) {
			log.Printf("Error exporting metrics: %s", err)
		},
		Timeout: defaultTimeout,
	}
}

// Run exports metrics periodically.
// Run should only be called once, and returns when Stop is called.
func (p *PushExporter) Run() {
	ticker := time.NewTicker(p.ReportingPeriod)
	defer func() {
		ticker.Stop()
		close(p.quit)
		close(p.didQuit)
	}()
	for {
		select {
		case <-ticker.C:
			go p.export()
		case <-p.quit:
			p.export()
			return
		}
	}
}

func (p *PushExporter) export() {
	ms := p.registry.Read()
	if len(ms) == 0 {
		return
	}

	ctx, done := context.WithTimeout(context.Background(), p.Timeout)
	defer done()
	defer func() {
		e := recover()
		if e != nil {
			p.OnError(fmt.Errorf("PushFunc panic: %s", e))
		}
	}()

	err := p.push(ctx, ms)
	if err != nil {
		p.OnError(err)
	}
}

// Stop causes Run to return after exporting metrics one last time.
// Only call stop after Run has been called.
// Stop may only be called once.
func (p *PushExporter) Stop() {
	p.quit <- struct{}{}
	<-p.didQuit
}
