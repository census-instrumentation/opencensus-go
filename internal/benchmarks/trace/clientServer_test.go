package trace

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"testing"

	"google.golang.org/grpc"

	"go.opencensus.io/internal/benchmarks/proto"
	"go.opencensus.io/plugin/ocgrpc"
)

var (
	noTraceNoStatsAddr, noTraceYesStatsAddr, yesTraceNoStatsAddr, yesTraceYesStatsAddr string
)

func MainTest(m *testing.M) {
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

func registerAndStartServer(srv *grpc.Server) string {
	ln, err := randPortListener()
	if err != nil {
		log.Fatalf("Listening on addr %q err: %v", addr, err)
	}
	proto.RegisterPingServer(srv, new(server))
	go srv.Serve(ln)
	return ln.Addr().String()
}

func randPortListener() (ln net.Listener, err error) {
	for i := 0; i < 1e3; i++ {
		p := 4000 + int(rand.Float64()*(65536-4000))
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", p))
log.Printf("p: %d\n", p)
		if err == nil {
			return ln, nil
		}
	}
	return
}

var (
	noTraceNoStatsClient   = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: true, NoStats: true})
	noTraceYesStatsClient  = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: true, NoStats: false})
	yesTraceNoStatsClient  = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: false, NoStats: true})
	yesTraceYesStatsClient = grpc.WithStatsHandler(&ocgrpc.ClientHandler{NoTrace: false, NoStats: false})

	noTraceNoStatsServer   = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: true, NoStats: true}))
	noTraceYesStatsServer  = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: true, NoStats: false}))
	yesTraceNoStatsServer  = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: false, NoStats: true}))
	yesTraceYesStatsServer = grpc.NewServer(grpc.StatsHandler(&ocgrpc.ServerHandler{NoTrace: false, NoStats: false}))
)
