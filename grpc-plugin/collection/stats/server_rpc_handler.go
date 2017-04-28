// Copyright 2017 Google Inc.
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

package stats

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tags"
	pb "github.com/google/instrumentation-proto/stats"
)

// handleRPCServerContext gets the metadata from GRPC context, extracts the
// encoded tags from it, creates a new github.com/google/instrumentation-go/stats/tagging.TagSet,
// adds it to the local context using tagging.NewContextWithTagSet and finally
// returns the new ctx.
func handleRPCServerContext(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("handleRPCServerContext(_) called with nil context")
		}
		return ctx
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerContext(_) failed to retrieve metadata from context")
		}
		return ctx
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerContext(_) failed to retrieve peer from context")
		}
		return ctx
	}

	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		if glog.V(2) {
			glog.Infof("handleRPCClientContext(_) called with info.FullMethodName bad format. got %v, want '/$service/$method/'", info.FullMethodName)
		}
		return ctx
	}

	d := &rpcData{
		startTime:   startTime,
		serviceName: names[1],
		methodName:  names[2],
	}

	if peer.AuthInfo != nil {
		d.authProtocol = peer.AuthInfo.AuthType()
	}

	if deadline, ok := ctx.Deadline(); ok {
		d.deadline = deadline
	}

	ctx, err := createStatsContext(ctx, md, d.methodName)
	if err != nil {
		return ctx
	}

	return context.WithValue(ctx, grpcInstRPCKey, d)
}

func handleRPCServerInHeader(ctx context.Context, s *stats.InHeader) {
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerInHeader(_) couldn't retrieve *serverConnStatus from context")
		}
		return
	}
	// TODO(acetechnologist): record rpc started?
	// measurement := RPCclientRpcStarted.CreateMeasurement(float64(1))
	// istats.RecordMeasurements(ctx, measurement)
	atomic.AddInt32(&scs.activeRequests, 1)

	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerInHeader(_) failed to retrieve *rpcData from context")
		}
		return
	}

	// TODO(acetechnologist): add these to context tags?
	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr

	// TODO(acetechnologist):
	// If CPU profiler is enabled notify the stats package profiler of the
	// start of a new RPC. This cannot be invoked on handleRPCServerContext
	// because a single routine calls handleRPCServerContext for all RPCs.
	// if stats.ServerRPCStart != nil {
	// 	stats.ServerRPCStart(ctx)
	// }
}

func handleRPCServerInPayload(ctx context.Context, s *stats.InPayload) {
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerInPayload(_) couldn't retrieve *serverConnStatus from context")
		}
		return
	}

	// TODO(acetechnologist): record these or later?
	// Record payload length received on this connection.
	atomic.AddInt64(&scs.requests.count, 1)
	atomic.AddInt64(&scs.requests.numBytes, int64(s.Length))

	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerInPayload(_) failed to retrieve *rpcData from context")
		}
		return
	}
	// TODO(acetechnologist): record these or later?
	// Record payload length received on this rpc.
	atomic.AddInt32(&d.reqLen, int32(s.Length))
	atomic.AddInt32(&d.wireReqLen, int32(s.WireLength))
}

func handleRPCServerOutPayload(ctx context.Context, s *stats.OutPayload) error {
	// Record payload length sent on this rpc.
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCServerOutPayload failed to extract *rpcData")
	}
	atomic.AddInt32(&d.respLen, int32(s.Length))
	atomic.AddInt32(&d.wireRespLen, int32(s.WireLength))

	// TODO(acetechnologist):
	// argumentType, ok := s.Payload.(proto.Message)
	// if !ok {
	// 	return fmt.Errorf("handleRPCInPayloadServer failed to extract argumentType. s.Payload is of type %T want type proto.Message", s.Payload)
	// }
	// payload := &rpctrace.Payload{
	// 	Pay:        s.Data,
	// 	PayLen:     s.Length,
	// 	WirePayLen: s.WireLength,
	// }
	// d.payloadResp = payload
	return nil
}

// GenerateServerTrailer records the elapsed time of the RPC, and generates the
// server trailer metadata that needs to be sent to the client.
func generateRPCServerTrailer(ctx context.Context) (metadata.MD, error) {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return nil, errors.New("generateRPCServerTrailer(_) failed to retrieve *rpcData from context")
	}

	// TODO(acetchnologist): generate proto statspb.RpcServerStats and create metadata.MD
	// Record payload length sent on this rpc.
	d.serverElapsedTime = time.Since(d.startTime)
	// elapsed := &statspb.RpcServerStats{
	// 	ServerElapsedTime: proto.Float64(float64(d.serverElapsedTime.Seconds())),
	// }
	// b, err := proto.Marshal(elapsed)
	// if err != nil {
	// 	log.Errorf("generateRPCServerTrailer cannot marshal %v to proto. %v", elapsed, err)
	// }
	// return metadata.Pairs(statsKey, string(b)), nil
	return nil, nil
}

func handleRPCServerEnd(ctx context.Context, s *stats.End) {
	// Decrement count of active RPCs on the connection.
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerEnd(_) failed to retrieve *serverConnStatus from context")
		}
		return
	}
	atomic.AddInt32(&scs.activeRequests, -1)

	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCServerEnd(_) failed to retrieve *rpcData from context")
		}
		return
	}

	// TODO(acetchnologist): Add more measurement here?
	var measurements []istats.Measurement
	measurements = append(measurements, RPCserverRequestBytes.CreateMeasurement(float64(d.reqLen)))
	measurements = append(measurements, RPCserverResponseBytes.CreateMeasurement(float64(d.respLen)))
	measurements = append(measurements, RPCserverServerElapsedTime.CreateMeasurement(float64(d.serverElapsedTime)/float64(time.Millisecond)))

	d.err = s.Error
	if d.err != nil {
		measurements = append(measurements, RPCserverErrorCount.CreateMeasurement(1))
	}

	istats.RecordMeasurements(ctx, measurements...)
}

// createStatsContext creates a census context from the gRPC context and tags
// received in metadata.
func createStatsContext(ctx context.Context, md metadata.MD, methodName string) (context.Context, error) {
	var cc pb.StatsContext

	if statsBin, ok := md[statsKey]; ok {
		if len(statsBin) != 1 {
			return nil, errors.New("createStatsContext(_) failed to retrieve statsBin from metadata. Have a length different than 1 in the metadata received")
		}

		if err := proto.Unmarshal([]byte(statsBin[0]), &cc); err != nil {
			return nil, fmt.Errorf("createStatsContext(_) failed to unmarshal statsBin[0]. Format is incorrect: %v. %v", statsBin[0], err)
		}
	}

	mut := keyMethod.CreateMutation(methodName, tagging.BehaviorAddOrReplace)

	builder := &tagging.TagSetBuilder{}
	builder.StartFromEncoded(cc.Tags)
	builder.AddMutations(mut)
	return tagging.ContextWithNewTagSet(ctx, builder.Build())
}
