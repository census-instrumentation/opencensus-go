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
	"strings"
	"testing"
	"time"
)

func TestPushExporter_Run(t *testing.T) {
	var pe PushExporter
	r := NewRegistry()
	r.AddProducer(&constProducer{&Metric{}})
	exported := make(chan bool, 1)
	pe.Init(r, func(ctx context.Context, ms []*Metric) error {
		_, ok := ctx.Deadline()
		if !ok {
			t.Fatal("Expected a deadline")
		}
		select {
		case exported <- true:
		default:
		}
		return nil
	})
	pe.ReportingPeriod = 100 * time.Millisecond

	go pe.Run()
	defer pe.Stop()

	select {
	case _ = <-exported:
	case <-time.After(1 * time.Second):
		t.Fatal("PushFunc should have been called")
	}
}

func TestPushExporter_Run_panic(t *testing.T) {
	var pe PushExporter
	r := NewRegistry()
	r.AddProducer(&constProducer{&Metric{}})
	errs := make(chan error, 1)
	pe.Init(r, func(ctx context.Context, ms []*Metric) error {
		panic("test")
	})
	pe.ReportingPeriod = 100 * time.Millisecond
	pe.OnError = func(err error) {
		errs <- err
	}

	go pe.Run()
	defer pe.Stop()

	select {
	case err := <-errs:
		if !strings.Contains(err.Error(), "test") {
			t.Error("Should contain the panic arg")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("OnError should be called")
	}
}

func TestPushExporter_Stop(t *testing.T) {
	var pe PushExporter
	r := NewRegistry()
	r.AddProducer(&constProducer{&Metric{}})
	exported := make(chan bool, 1)
	pe.Init(r, func(ctx context.Context, ms []*Metric) error {
		select {
		case exported <- true:
		default:
			t.Fatal("Export should only be called once")
		}
		return nil
	})
	pe.ReportingPeriod = time.Hour // prevent timer-based push

	go pe.Run()
	pe.Stop()

	select {
	case _ = <-exported:
	default:
		t.Fatal("PushFunc should have been called before Stop returns")
	}
}
