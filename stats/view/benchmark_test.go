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
//

package view

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
)

var (
	m  = stats.Float64("m", "", "")
	k1 = tag.MustNewKey("k1")
	k2 = tag.MustNewKey("k2")
	k3 = tag.MustNewKey("k3")
	k4 = tag.MustNewKey("k4")
	k5 = tag.MustNewKey("k5")
	k6 = tag.MustNewKey("k6")
	k7 = tag.MustNewKey("k7")
	k8 = tag.MustNewKey("k8")

	view = &View{
		Measure:     m,
		Aggregation: Distribution(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
		TagKeys:     []tag.Key{k1, k2},
	}
)

// BenchmarkRecordReqCommand benchmarks calling the internal recording machinery
// directly.
func BenchmarkRecordReqCommand(b *testing.B) {
	w := NewMeter().(*worker)

	register := &registerViewReq{views: []*View{view}, err: make(chan error, 1)}
	register.handleCommand(w)
	if err := <-register.err; err != nil {
		b.Fatal(err)
	}

	ctxs := prepareContexts(10)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		record := &recordReq{
			ms: []stats.Measurement{
				m.M(1),
				m.M(1),
				m.M(1),
				m.M(1),
				m.M(1),
				m.M(1),
				m.M(1),
				m.M(1),
			},
			tm: tag.FromContext(ctxs[i%len(ctxs)]),
			t:  time.Now(),
		}
		record.handleCommand(w)
	}
}

func BenchmarkRecordViaStats(b *testing.B) {

	meter := NewMeter()
	meter.Start()
	defer meter.Stop()
	meter.Register(view)
	defer meter.Unregister(view)

	ctxs := prepareContexts(10)
	rec := stats.WithRecorder(meter)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		stats.RecordWithOptions(ctxs[i%len(ctxs)], rec, stats.WithMeasurements(m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1), m.M(1)))
	}

}

func prepareContexts(tagCount int) []context.Context {
	ctxs := make([]context.Context, 0, tagCount)
	for i := 0; i < tagCount; i++ {
		ctx, _ := tag.New(context.Background(),
			tag.Upsert(k1, fmt.Sprintf("v%d", i)),
			tag.Upsert(k2, fmt.Sprintf("v%d", i)),
			tag.Upsert(k3, fmt.Sprintf("v%d", i)),
			tag.Upsert(k4, fmt.Sprintf("v%d", i)),
			tag.Upsert(k5, fmt.Sprintf("v%d", i)),
			tag.Upsert(k6, fmt.Sprintf("v%d", i)),
			tag.Upsert(k7, fmt.Sprintf("v%d", i)),
			tag.Upsert(k8, fmt.Sprintf("v%d", i)),
		)
		ctxs = append(ctxs, ctx)
	}

	return ctxs
}
