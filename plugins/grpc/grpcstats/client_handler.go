// Copyright 2017, OpenCensus Authors
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

package grpcstats

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	istats "go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
)

// clientHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC client.
type clientHandler struct{}

// ClientStatsHandler returns a grpc/stats.Handler implementation
// that collects stats for a gRPC client. Predefined
// measures and views can be used to access the collected data.
func ClientStatsHandler() stats.Handler {
	return clientHandler{}
}

// TagConn adds connection related data to the given context and returns the
// new context.
func (ch clientHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
	return ctx
}

// HandleConn processes the connection events.
func (ch clientHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
}

// TagRPC gets the tag.Map populated by the application code, serializes
// its tags into the GRPC metadata in order to be sent to the server.
func (ch clientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("clientHandler.TagRPC called with nil context")
		}
		return ctx
	}

	if info == nil {
		if glog.V(2) {
			glog.Infof("clientHandler.TagRPC called with nil info.", info.FullMethodName)
		}
		return ctx
	}
	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		if glog.V(2) {
			glog.Infof("clientHandler.TagRPC called with info.FullMethodName bad format. got %v, want '/$service/$method/'", info.FullMethodName)
		}
		return ctx
	}
	serviceName := names[1]
	methodName := names[2]

	d := &rpcData{
		startTime: startTime,
	}

	ts := tag.FromContext(ctx)
	encoded := tag.Encode(ts)
	ctx = stats.SetTags(ctx, encoded)

	tagMap, err := tag.NewMap(ts,
		tag.Upsert(keyService, serviceName),
		tag.Upsert(keyMethod, methodName),
	)
	if err == nil {
		ctx = tag.NewContext(ctx, tagMap)
	}
	// TODO(acetechnologist): should we be recording this later? What is the
	// point of updating d.reqLen & d.reqCount if we update now?
	istats.Record(ctx, RPCClientStartedCount.M(1))

	return context.WithValue(ctx, grpcClientRPCKey, d)
}

// HandleRPC processes the RPC events.
func (ch clientHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
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

func (ch clientHandler) handleRPCOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("clientHandler.handleRPCOutPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.Record(ctx, RPCClientRequestBytes.M(int64(s.Length)))
	atomic.AddUint64(&d.reqCount, 1)
}

func (ch clientHandler) handleRPCInPayload(ctx context.Context, s *stats.InPayload) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("clientHandler.handleRPCInPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.Record(ctx, RPCClientResponseBytes.M(int64(s.Length)))
	atomic.AddUint64(&d.respCount, 1)
}

func (ch clientHandler) handleRPCEnd(ctx context.Context, s *stats.End) {
	d, ok := ctx.Value(grpcClientRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("clientHandler.handleRPCEnd failed to retrieve *rpcData from context")
		}
		return
	}
	elapsedTime := time.Since(d.startTime)

	var measurements []istats.Measurement
	measurements = append(measurements, RPCClientRequestCount.M(int64(d.reqCount)))
	measurements = append(measurements, RPCClientResponseCount.M(int64(d.respCount)))
	measurements = append(measurements, RPCClientFinishedCount.M(1))
	measurements = append(measurements, RPCClientRoundTripLatency.M(float64(elapsedTime)/float64(time.Millisecond)))

	if s.Error != nil {
		errorCode := s.Error.Error()
		newTagMap, err := tag.NewMap(tag.FromContext(ctx),
			tag.Upsert(keyOpStatus, errorCode),
		)
		if err == nil {
			ctx = tag.NewContext(ctx, newTagMap)
		}
		measurements = append(measurements, RPCClientErrorCount.M(1))
	}

	istats.Record(ctx, measurements...)
}
