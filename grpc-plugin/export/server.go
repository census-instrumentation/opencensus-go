package export

import (
	context "golang.org/x/net/context"

	pb "github.com/golang/protobuf/ptypes/empty"
	spb "github.com/google/instrumentation-go/grpc-plugin/generated-proto/service"
	statsPb "github.com/google/instrumentation-go/grpc-plugin/generated-proto/stats"
	istats "github.com/google/instrumentation-go/stats"
)

type server struct {
	c chan *istats.View
}

func NewServer() spb.MonitoringServer {
	s := &server{
		c: make(chan *istats.View, 1024),
	}
	return s
}

// Return canonical RPC stats
func (s *server) GetCanonicalRpcStats(ctx context.Context, empty *pb.Empty) (resp *spb.CanonicalRpcStats, err error) {
	return nil, nil
}

// Query the server for specific stats
func (s *server) GetStats(ctx context.Context, req *spb.StatsRequest) (resp *spb.StatsResponse, err error) {
	return nil, nil
}

// Request the server to stream back snapshots of the requested stats
func (s *server) WatchStats(req *spb.StatsRequest, stream spb.Monitoring_WatchStatsServer) error {
	c2 := make(chan []*istats.View, 1024)

	err := istats.SubscribeToManyViews(req.GetViewNames(), req.GetMeasurementNames(), c2)
	for {
		vws := <-c2

		var vwResps []*spb.ViewResponse
		for _, vw := range vws {
			vwResps = append(vwResps, &spb.ViewResponse{
				View: &statsPb.View{},
			})
		}

		resp := &spb.StatsResponse{
			ViewResponses: vwResps,
		}
		if err := stream.Send(resp); err != nil {
			// handle
		}
	}
}

// Return request traces.
func (s *server) GetRequestTraces(ctx context.Context, req *spb.TraceRequest) (resp *spb.TraceResponse, err error) {
	return nil, nil
}

// Return application-defined groups of monitoring data.
// This is a low level facility to allow extension of the monitoring API to
// application-specific monitoring data. Frameworks may use this to define
// additional groups of monitoring data made available by servers.
func (s *server) GetCustomMonitoringData(ctx context.Context, req *spb.MonitoringDataGroup) (resp *spb.CustomMonitoringData, err error) {
	return nil, nil
}
