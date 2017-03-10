package export

import pb "github.com/google/instrumentation-go/grpc-plugin/generated-proto/service"

// Return canonical RPC stats
func GetCanonicalRpcStats(ctx context.Context, out *pb.CanonicalRpcStats) error {

}

// Query the server for specific stats
func GetStats(ctx context.Context, in *pb.StatsRequest, out *pb.StatsResponse) error {

}

// Request the server to stream back snapshots of the requested stats
func WatchStats(ctx context.Context, in *pb.StatsRequest) returns (stream StatsResponse) {
  }


// Return request traces.
func GetRequestTraces(TraceRequest) returns(TraceResponse) {
  // TODO(aveitch): Please define the messages here
}

  // Return application-defined groups of monitoring data.
  // This is a low level facility to allow extension of the monitoring API to
  // application-specific monitoring data. Frameworks may use this to define
  // additional groups of monitoring data made available by servers.
  rpc GetCustomMonitoringData(MonitoringDataGroup)
    returns (CustomMonitoringData) {
  }

/*
func (s *Service) AcceptForSiriusBrownfieldMigration(ctx context.Context, in *spb.AcceptForSiriusBrownfieldMigrationRequest, out *spb.AcceptForSiriusBrownfieldMigrationResponse) error {
}
*/
