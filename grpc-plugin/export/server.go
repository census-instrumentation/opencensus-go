package export

import pb "github.com/google/instrumentation-go/grpc-plugin/generated-proto/service"

// Return canonical RPC stats
func GetCanonicalRpcStats(ctx context.Context, out *pb.CanonicalRpcStats) error {

}

// Query the server for specific stats
func GetStats(ctx context.Context, in *pb.StatsRequest, out *pb.StatsResponse) error {

}

// Request the server to stream back snapshots of the requested stats
func WatchStats(ctx context.Context, in *pb.StatsRequest, s pb.Monitoring_WatchStats) error {
  // s.Send(*pb.StatsResponse)
  // s.Send(*pb.StatsResponse)
}


// Return request traces.
func GetRequestTraces(ctx context.Context, in *pb.TraceRequest, out *pb.TraceResponse) error{
  // TODO(aveitch): Please define the messages here
}

// Return application-defined groups of monitoring data.
// This is a low level facility to allow extension of the monitoring API to
// application-specific monitoring data. Frameworks may use this to define
// additional groups of monitoring data made available by servers.
func GetCustomMonitoringData(ctx context.Context, in *pb.MonitoringDataGroup, , out *pb.CustomMonitoringData) error {
}