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

package zpages

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.opencensus.io/internal/testpb"
	"go.opencensus.io/stats/view"
)

func TestRpcz(t *testing.T) {
	client, cleanup := testpb.NewTestClient(t)
	defer cleanup()

	_, err := client.Single(context.Background(), &testpb.FooRequest{})
	if err != nil {
		t.Fatal(err)
	}

	view.SetReportingPeriod(time.Millisecond)
	time.Sleep(2 * time.Millisecond)
	view.SetReportingPeriod(time.Second)

	mu.Lock()
	defer mu.Unlock()

	if len(snaps) == 0 {
		t.Fatal("Expected len(snaps) > 0")
	}

	snapshot, ok := snaps[methodKey{"testpb.Foo/Single", false}]
	if !ok {
		t.Fatal("Expected method stats not recorded")
	}

	last := snapshot.LatencyMinute.intervals[snapshot.LatencyMinute.lastUpdate]
	if got, want := last.distribution.Count, int64(1); got != want {
		t.Errorf("snapshot.CountTotal = %d; want %d", got, want)
	}
}

func TestGetStatsPage(t *testing.T) {
	// Reset views.
	for v := range viewType {
		view.Unregister(v)
		view.Register(v)
	}

	zpages := httptest.NewServer(Handler)
	defer zpages.Close()

	client, done := testpb.NewTestClient(t)
	defer done()

	ctx := context.Background()
	client.Single(ctx, &testpb.FooRequest{Echo: make([]byte, 256), SleepNanos: int64(10 * time.Millisecond)})
	client.Single(ctx, &testpb.FooRequest{Fail: true})

	view.SetReportingPeriod(10 * time.Millisecond)

	var stats *statsPage
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		stats = getStatsPage()
		mu.Unlock()
		if len(stats.StatGroups[1].Snapshots) > 0 {
			break
		}
	}
	if stats == nil {
		t.Fatal("stats == nil")
	}
	WriteTextRpczPage(os.Stdout)

	mu.Lock()
	if got, want := len(stats.StatGroups[0].Snapshots), 1; got != want {
		t.Fatalf("stats.StatGroups[0].Snapshots = %v; want len = 1", got)
	}
	snapshot := stats.StatGroups[0].Snapshots[0]
	if got, want := snapshot.Method, "testpb.Foo/Single"; got != want {
		t.Errorf("snapshot.Method = %q; want %q", got, want)
	}
	_, _, latency := snapshot.LatencyHour.read()
	if got, want := latency.Count, int64(2); got != want {
		t.Errorf("latency.Count = %d; want %d", got, want)
	}
	_, _, errors := snapshot.ErrorsHour.read()
	if got, want := errors.Count, int64(1); got != want {
		t.Errorf("errors.Count = %d; want %d", got, want)
	}
	mu.Unlock()

	resp, _ := http.Get(zpages.URL + "/rpcz")
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Fatalf("resp.StatusCode = %d; want %d", got, want)
	}
}
