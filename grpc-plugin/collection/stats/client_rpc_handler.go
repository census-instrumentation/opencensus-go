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
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tags"
	pb "github.com/google/instrumentation-proto/stats"
)

// handleRPCClientContext gets the github.com/google/instrumentation-go/stats/tagging.TagSet
// set by the application code, serializes its tags into the GRPC metadata in
// order to be sent to the server.
func handleRPCClientContext(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
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
		isClient:    true,
	}

	ts := tagging.FromContext(ctx)
	encoded := tagging.EncodeToFullSignature(ts)

	statsCtx := &pb.StatsContext{
		Tags: encoded,
	}
	statsBin, err := proto.Marshal(statsCtx)
	if err != nil {
		if glog.V(2) {
			glog.Infof("handleRPCClientContext(_) cannot marshal pb.StatsContext %v\n. %v", statsCtx, err)
		}
		return ctx
	}

	// Join old metadata with new one, so the old metadata information is not lost.
	md, _ := metadata.FromContext(ctx)
	ctx = metadata.NewContext(ctx, metadata.Join(md, metadata.Pairs(statsKey, string(statsBin))))

	return context.WithValue(ctx, grpcInstRPCKey, d)
}

func handleRPCClientBegin(ctx context.Context, s *stats.Begin) {
	// TODO(acetechnologist): record rpc started?
	// measurement := RPCclientRpcStarted.CreateMeasurement(float64(1))
	// istats.RecordMeasurements(ctx, measurement)
}

func handleRPCClientOutHeader(ctx context.Context, s *stats.OutHeader) {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCClientOutHeader(_) failed to retrieve *rpcData from context")
		}
		return
	}

	// TODO(acetechnologist): add these to context tags?
	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr
}

func handleRPCClientOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCClientOutPayload(_) failed to retrieve *rpcData from context")
		}
		return
	}

	// TODO(acetechnologist): record these or later?
	atomic.AddInt32(&d.reqLen, int32(s.Length))
	atomic.AddInt32(&d.wireReqLen, int32(s.WireLength))
}

func handleRPCClientInPayload(ctx context.Context, s *stats.InPayload) {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCClientInPayload(_) failed to retrieve *rpcData from context")
		}
		return
	}

	// TODO(acetechnologist): record these or later?
	atomic.AddInt32(&d.respLen, int32(s.Length))
	atomic.AddInt32(&d.wireRespLen, int32(s.WireLength))
	return
}

func handleRPCClientEnd(ctx context.Context, s *stats.End) {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("handleRPCClientEnd(_) failed to retrieve *rpcData from context")
		}
		return
	}

	d.err = s.Error
	d.totalElapsedTime = time.Since(d.startTime)

	var measurements []istats.Measurement
	measurements = append(measurements, RPCclientRequestBytes.CreateMeasurement(float64(d.reqLen)))
	measurements = append(measurements, RPCclientResponseBytes.CreateMeasurement(float64(d.respLen)))
	measurements = append(measurements, RPCclientServerElapsedTime.CreateMeasurement(float64(d.serverElapsedTime)/float64(time.Millisecond)))

	if d.err != nil {
		measurements = append(measurements, RPCclientErrorCount.CreateMeasurement(1))
	}

	istats.RecordMeasurements(ctx, measurements...)
}
