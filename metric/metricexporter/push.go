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

package metricexporter

import (
	"context"
	"fmt"
	"go.opencensus.io/metric"
	"go.opencensus.io/trace"
	"log"
	"time"
)

// Push exports metrics at regular intervals.
// Use Init to initialize a new Push metric exporter.
// Do not modify fields of Push after calling Run.
type Push struct {
	push          PushFunc
	quit, didQuit chan struct{}

	// Registry is the metric registry that metrics will be read from.
	// Init sets this to the default metric registry by default.
	Registry *metric.Registry

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

	// NewContext may be set to override the default context that will be
	// used for exporting.
	// By default, this is set to context.Background.
	NewContext func() context.Context
}

// PushFunc is a function that exports metrics, for example to a monitoring
// backend.
type PushFunc func(context.Context, []*metric.Metric) error

const (
	defaultReportingPeriod = 10 * time.Second
	defaultTimeout         = 5 * time.Second
)

// Init initializes a new Push metrics exporter that reads
// metrics from the given registry and pushes them to the given PushFunc.
func NewPush(push PushFunc) *Push {
	return &Push{
		push:            push,
		quit:            make(chan struct{}),
		didQuit:         make(chan struct{}),
		Registry:        metric.DefaultRegistry(),
		ReportingPeriod: defaultReportingPeriod,
		OnError: func(err error) {
			log.Printf("Error exporting metrics: %s", err)
		},
		Timeout:    defaultTimeout,
		NewContext: context.Background,
	}
}

// Run exports metrics periodically.
// Run should only be called once, and returns when Stop is called.
func (p *Push) Run() {
	ticker := time.NewTicker(p.ReportingPeriod)
	defer func() {
		ticker.Stop()
		close(p.quit)
		close(p.didQuit)
	}()
	for {
		select {
		case <-ticker.C:
			go p.Export()
		case <-p.quit:
			p.Export()
			return
		}
	}
}

// Export reads all metrics from the registry and exports them.
// Most users will rely on Run, which calls Export in a loop until stopped.
func (p *Push) Export() {
	ms := p.Registry.Read()
	if len(ms) == 0 {
		return
	}

	ctx, done := context.WithTimeout(p.NewContext(), p.Timeout)
	defer done()

	// Create a Span that is never sampled to avoid sending many uninteresting
	// traces.
	ctx, span := trace.StartSpan(
		ctx,
		"go.opencensus.io/metric/exporter.Export",
		trace.WithSampler(trace.ProbabilitySampler(0.0)),
	)
	defer span.End()

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
func (p *Push) Stop() {
	p.quit <- struct{}{}
	<-p.didQuit
}
