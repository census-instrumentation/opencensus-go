package export

import (
	context "golang.org/x/net/context"

	"fmt"

	pb "github.com/golang/protobuf/ptypes/empty"
	"github.com/google/instrumentation-go/grpc-plugin/topb"
	istats "github.com/google/instrumentation-go/stats"
	statsPb "github.com/google/instrumentation-proto/stats"
	spb "github.com/grpc/grpc-proto/grpc/instrumentation/v1alpha"
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
func (s *server) GetCanonicalRpcStats(ctx context.Context, empty *pb.Empty) (*spb.CanonicalRpcStats, error) {

	/*
			RpcClientErrors
			RpcClientRequestBytes
			RpcClientResponseBytes
			RpcClientElapsedTime (RoundTripLatency)
			RpcClientServerElapsedTime

		// UncompressedRequestBytes defined in java but not in proto
		// UncompressedResponseBytes defined in java but not in proto
			RpcClientCompletedRpcs
			RpcClientStartedRpcs
			RpcClientRequestCount
			RpcClientResponseCount

			RpcServerErrors
			RpcServerRequestBytes
			RpcServerResponseBytes
			RpcServerElapsedTime	   // What is the difference between serverElapsedTime and ServerServerElapsedTime ?
			RpcServerServerElapsedTime // What is the difference between serverElapsedTime and ServerServerElapsedTime ?
			RpcServerCompletedRpcs
			RpcServerRequestCount	   // This is the started count.
		// UncompressedRequestBytes defined in java but not in proto
		// UncompressedResponseBytes defined in java but not in proto
			RpcServerResponseCount     // Why is this needed.
	*/
	return nil, nil
}

// Query the server for specific stats
func (s *server) GetStats(ctx context.Context, req *spb.StatsRequest) (*spb.StatsResponse, error) {
	views, err := istats.RetrieveViews(req.ViewNames, req.MeasurementNames)
	if err != nil {
		return nil, err
	}

	resp, err := buildStatsResponse(views)
	if err != nil {
		return nil, err
	}

	if req.DontIncludeDescriptorsInFirstResponse {
		return resp, nil
	}

	// TODO(mmoakil): where does req.DontIncludeDescriptorsInFirstResponse need
	// to be included? Only CanonicalRpcStats have these. And interestingly
	// GetCanonicalRpcStats(...) doesn't have a
	// DontIncludeDescriptorsInFirstResponse.
	return resp, nil
}

// Request the server to stream back snapshots of the requested stats
func (s *server) WatchStats(req *spb.StatsRequest, stream spb.Monitoring_WatchStatsServer) error {
	c2 := make(chan []*istats.View, 1024)

	err := istats.SubscribeToManyViews(req.GetViewNames(), req.GetMeasurementNames(), c2)
	if err != nil {
		return fmt.Errorf("WatchStats failed to subscribe view names %v and measurement names %v", req.GetViewNames(), req.GetMeasurementNames())
	}

	for {
		views := <-c2

		resp, err := buildStatsResponse(views)
		if err != nil {
			return err
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

// Return request traces.
func (s *server) GetRequestTraces(ctx context.Context, req *spb.TraceRequest) (*spb.TraceResponse, error) {
	return nil, nil
}

// Return application-defined groups of monitoring data.
// This is a low level facility to allow extension of the monitoring API to
// application-specific monitoring data. Frameworks may use this to define
// additional groups of monitoring data made available by servers.
func (s *server) GetCustomMonitoringData(ctx context.Context, req *spb.MonitoringDataGroup) (*spb.CustomMonitoringData, error) {
	return nil, nil
}

func buildStatsResponse(vws []*istats.View) (*spb.StatsResponse, error) {
	resp := &spb.StatsResponse{}

	for _, vw := range vws {
		vwpb, err := topb.View(vw)
		if err != nil {
			return nil, err
		}

		resp.ViewResponses = append(resp.ViewResponses, &spb.ViewResponse{
			MeasurementDescriptor: &statsPb.MeasurementDescriptor{},
			ViewDescriptor:        &statsPb.ViewDescriptor{},
			View:                  vwpb,
		})
	}

	return resp, nil
}
