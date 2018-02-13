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

package trace

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.opencensus.io/internal/benchmarks/defs"
	"go.opencensus.io/plugin/ocgrpc"
)

func BenchmarkUntraced1QPS(b *testing.B) {
	benchmarkUntraced(b, 1)
}

func BenchmarkUntraced10QPS(b *testing.B) {
	benchmarkUntraced(b, 10)
}

func BenchmarkUntraced100QPS(b *testing.B) {
	benchmarkUntraced(b, 100)
}

func BenchmarkUntraced1000QPS(b *testing.B) {
	benchmarkUntraced(b, 1000)
}

func BenchmarkTraced1QPS(b *testing.B) {
	benchmarkTraced(b, 1)
}

func BenchmarkTraced10QPS(b *testing.B) {
	benchmarkTraced(b, 10)
}

func BenchmarkTraced100QPS(b *testing.B) {
	benchmarkTraced(b, 100)
}

func BenchmarkTraced1000QPS(b *testing.B) {
	benchmarkTraced(b, 1000)
}

func benchmarkTraced(b *testing.B, qps int) {
	conn, err := grpc.Dial(addr, grpc.WithStatsHandler(ocgrpc.NewClientStatsHandler()), grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

func benchmarkUntraced(b *testing.B, qps int) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating untraced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

func runWithConn(b *testing.B, conn *grpc.ClientConn, qps int) {
	client := defs.NewPingClient(conn)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qpsIt(10, func() {
			pong, err := client.Ping(context.Background(), &defs.Payload{"Ping"})
			if err != nil {
				b.Fatalf("Pong #%d err: %v", i, err)
			}
			if pong == nil {
				b.Fatalf("Pong #%d is nil", i)
			}
		})
	}
	b.ReportAllocs()
}

func qpsIt(qps int, fn func()) {
	// 1000 QPS ==> (1s/1000Q) * (1e9ns/1s) ==> 0.001s or 1e6ns/Q
	period := time.Duration(int64(1 / float64(qps) * 1e9))
	tick := time.NewTicker(period)
	defer tick.Stop()

	var wg sync.WaitGroup
	defer wg.Wait()

	for i := 0; i < qps; i++ {
		wg.Add(1)
		go func() {
			fn()
			wg.Done()
		}()
	}
}
