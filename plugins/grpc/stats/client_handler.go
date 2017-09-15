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

	istats "github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"
	"github.com/golang/glog"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

var (
	// grpcClientConnKey is the key used to store client instrumentation
	// connection related data into the context.
	grpcClientConnKey *grpcInstrumentationKey
	// grpcClientRPCKey is the key used to store client instrumentation RPC
	// related data into the context.
	grpcClientRPCKey *grpcInstrumentationKey
)

// ClientHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC client.
type ClientHandler struct{}

// TagConn adds connection related data to the given context and returns the
// new context.
func (ch ClientHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
	return ctx
}

// HandleConn processes the connection events.
func (ch ClientHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
}

// TagRPC gets the github.com/census-instrumentation/opencensus-go/tags.TagsSet
// populated by the application code, serializes its tags into the GRPC
// metadata in order to be sent to the server.
func (ch ClientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("ClientHandler.TagRPC called with nil context")
		}
		return ctx
	}

	if info == nil {
		if glog.V(2) {
			glog.Infof("ClientHandler.TagRPC called with nil info.", info.FullMethodName)
		}
		return ctx
	}
	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		if glog.V(2) {
			glog.Infof("ClientHandler.TagRPC called with info.FullMethodName bad format. got %v, want '/$service/$method/'", info.FullMethodName)
		}
		return ctx
	}
	serviceName := names[1]
	methodName := names[2]

	d := &rpcData{
		startTime: startTime,
	}

	ts := tags.FromContext(ctx)
	encoded := tags.EncodeToFullSignature(ts)

	// Join old metadata with new one, so the old metadata information is not lost.
	md, _ := metadata.FromOutgoingContext(ctx)
	ctx = metadata.NewOutgoingContext(ctx, metadata.Join(md, metadata.Pairs(tagsKey, string(encoded))))

	tsb := tags.NewTagSetBuilder(ts)
	tsb.UpsertString(keyService, serviceName)
	tsb.UpsertString(keyMethod, methodName)

	// TODO(acetechnologist): should we be recording this later? What is the
	// point of updating d.reqLen & d.reqCount if we update now?
	ctx = tags.NewContext(ctx, tsb.Build())
	istats.RecordInt64(ctx, RPCClientStartedCount, 1)

	return context.WithValue(ctx, grpcClientRPCKey, d)
}

// HandleRPC processes the RPC events.
func (ch ClientHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.Begin, *stats.OutHeader, *stats.InHeader, *stats.InTrailer, *stats.OutTrailer:
		// do nothing for client
	case *stats.OutPayload:
		ch.handleRPCOutPayload(ctx, st)
	case *stats.InPayload:
		ch.handleRPCInPayload(ctx, st)
	case *stats.End:
		ch.handleRPCEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

func (ch ClientHandler) handleRPCOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("ClientHandler.handleRPCOutPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.RecordInt64(ctx, RPCClientRequestBytes, int64(s.Length))
	atomic.AddUint64(&d.reqCount, 1)
}

func (ch ClientHandler) handleRPCInPayload(ctx context.Context, s *stats.InPayload) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("ClientHandler.handleRPCInPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.RecordInt64(ctx, RPCClientResponseBytes, int64(s.Length))
	atomic.AddUint64(&d.respCount, 1)
}

func (ch ClientHandler) handleRPCEnd(ctx context.Context, s *stats.End) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("ClientHandler.handleRPCEnd failed to retrieve *rpcData from context")
		}
		return
	}
	elapsedTime := time.Since(d.startTime)

	var measurements []istats.Measurement
	measurements = append(measurements, RPCClientRequestCount.Is(int64(d.reqCount)))
	measurements = append(measurements, RPCClientResponseCount.Is(int64(d.respCount)))
	measurements = append(measurements, RPCClientFinishedCount.Is(1))
	measurements = append(measurements, RPCClientRoundTripLatency.Is(float64(elapsedTime)/float64(time.Millisecond)))

	if s.Error != nil {
		errorCode := s.Error.Error()
		ts := tags.FromContext(ctx)
		tsb := tags.NewTagSetBuilder(ts)
		tsb.UpsertString(keyOpStatus, errorCode)

		ctx = tags.NewContext(ctx, tsb.Build())
		measurements = append(measurements, RPCClientErrorCount.Is(1))
	}

	istats.Record(ctx, measurements...)
}
