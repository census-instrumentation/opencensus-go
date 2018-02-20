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
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"go.opencensus.io/internal/benchmarks/proto"
	"go.opencensus.io/plugin/ocgrpc"
)

func TestMain(m *testing.M) {
	plainServerAddr = registerAndStartServer(plainServer)
	noTraceNoStatsAddr = registerAndStartServer(noTraceNoStatsServer)
	noTraceYesStatsAddr = registerAndStartServer(noTraceYesStatsServer)
	yesTraceNoStatsAddr = registerAndStartServer(yesTraceNoStatsServer)
	yesTraceYesStatsAddr = registerAndStartServer(yesTraceYesStatsServer)

	defer noTraceNoStatsServer.Stop()
	defer noTraceYesStatsServer.Stop()
	defer yesTraceNoStatsServer.Stop()
	defer yesTraceYesStatsServer.Stop()

	os.Exit(m.Run())
}

type server int

func (s *server) Ping(ctx context.Context, p *proto.Payload) (*proto.Payload, error) {
	return &proto.Payload{Body: "Pong"}, nil
}

func registerAndStartServer(srv *grpc.Server) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("registerAndStartServer Listen err: %v", err)
	}
	proto.RegisterPingServer(srv, new(server))
	go srv.Serve(ln)
	return ln.Addr().String()
}

var (
	plainServerAddr, noTraceNoStatsAddr, noTraceYesStatsAddr, yesTraceNoStatsAddr, yesTraceYesStatsAddr string
)

var (
	noTraceNoStatsClient   = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: true, NoStats: true})
	noTraceYesStatsClient  = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: true, NoStats: false})
	yesTraceNoStatsClient  = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: false, NoStats: true})
	yesTraceYesStatsClient = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: false, NoStats: false})

	plainServer            = grpc.NewServer()
	noTraceNoStatsServer   = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: true, NoStats: true}))
	noTraceYesStatsServer  = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: true, NoStats: false}))
	yesTraceNoStatsServer  = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: false, NoStats: true}))
	yesTraceYesStatsServer = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: false, NoStats: false}))
)

func runWithConn(b *testing.B, conn *grpc.ClientConn, qps int) {
	client := proto.NewPingClient(conn)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qpsIt(qps, func() {
			pong, err := client.Ping(context.Background(), &proto.Payload{Body: "Ping"})
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
		<-tick.C
	}
}
