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
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/net/context"

	"github.com/golang/glog"
	istats "go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

// serverHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC server.
type serverHandler struct{}

// NewServerStatsHandler returns a grpc/stats.Handler implementation
// that collects stats for a gRPC server. Predefined
// measures and views can be used to access the collected data.
func NewServerStatsHandler() stats.Handler {
	return serverHandler{}
}

// TagConn adds connection related data to the given context and returns the
// new context.
func (sh serverHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
	return ctx
}

// HandleConn processes the connection events.
func (sh serverHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	// Do nothing. This is here to satisfy the interface "google.golang.org/grpc/stats.Handler"
}

// TagRPC gets the metadata from GRPC context, extracts the encoded tags from
// it, creates a new tag.Map,
// adds it to the local context using tagging.NewContextWithTagsSet and finally
// returns the new ctx.
func (sh serverHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if info == nil {
		if glog.V(2) {
			glog.Infof("serverHandler.TagRPC called with nil info.", info.FullMethodName)
		}
		return ctx
	}
	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		if glog.V(2) {
			glog.Infof("serverHandler.TagRPC called with info.FullMethodName bad format. got %v, want '/$service/$method/'", info.FullMethodName)
		}
		return ctx
	}
	serviceName := names[1]
	methodName := names[2]

	d := &rpcData{
		startTime: startTime,
	}

	ts, err := sh.createTagMap(ctx, serviceName, methodName)
	if err != nil {
		return ctx
	}
	ctx = tag.NewContext(ctx, ts)

	istats.Record(ctx, RPCServerStartedCount.M(1))
	return context.WithValue(ctx, grpcServerRPCKey, d)
}

// HandleRPC processes the RPC events.
func (sh serverHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.Begin, *stats.InHeader, *stats.InTrailer, *stats.OutHeader, *stats.OutTrailer:
		// Do nothing for server
	case *stats.InPayload:
		sh.handleRPCInPayload(ctx, st)
	case *stats.OutPayload:
		// For stream it can be called multiple times per RPC.
		sh.handleRPCOutPayload(ctx, st)
	case *stats.End:
		sh.handleRPCEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

// GenerateServerTrailer is intended to be called in server interceptor.
// TODO(acetechnologist): could eventually be used to record the elapsed time
// of the RPC on the server side and generate the server trailer metadata that
// needs to be sent to the client.
func (sh serverHandler) GenerateServerTrailer(ctx context.Context) (metadata.MD, error) {
	return nil, nil
}

func (sh serverHandler) handleRPCInPayload(ctx context.Context, s *stats.InPayload) {
	d, ok := ctx.Value(grpcServerRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("serverHandler.handleRPCInPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.Record(ctx, RPCServerRequestBytes.M(int64(s.Length)))
	atomic.AddInt64(&d.reqCount, 1)
}

func (sh serverHandler) handleRPCOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(grpcServerRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("serverHandler.handleRPCOutPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.Record(ctx, RPCServerResponseBytes.M(int64(s.Length)))
	atomic.AddInt64(&d.respCount, 1)
}

func (sh serverHandler) handleRPCEnd(ctx context.Context, s *stats.End) {
	d, ok := ctx.Value(grpcServerRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("serverHandler.handleRPCEnd failed to retrieve *rpcData from context")
		}
		return
	}
	elapsedTime := time.Since(d.startTime)

	reqCount := atomic.LoadInt64(&d.reqCount)
	respCount := atomic.LoadInt64(&d.respCount)

	var m []istats.Measurement
	m = append(m, RPCServerRequestCount.M(reqCount))
	m = append(m, RPCServerResponseCount.M(respCount))
	m = append(m, RPCServerFinishedCount.M(1))
	m = append(m, RPCServerServerElapsedTime.M(float64(elapsedTime)/float64(time.Millisecond)))
	if s.Error != nil {
		errorCode := s.Error.Error()
		tm, err := tag.NewMap(ctx,
			tag.Upsert(keyOpStatus, errorCode),
		)
		if err == nil {
			ctx = tag.NewContext(ctx, tm)
		}
		m = append(m, RPCServerErrorCount.M(1))
	}

	istats.Record(ctx, m...)
}

// createTagMap creates a new tag map containing the tags extracted from the
// gRPC metadata.
func (sh serverHandler) createTagMap(ctx context.Context, serviceName, methodName string) (*tag.Map, error) {
	mods := []tag.Mutator{
		tag.Upsert(keyService, serviceName),
		tag.Upsert(keyMethod, methodName),
	}
	if tagsBin := stats.Tags(ctx); tagsBin != nil {
		old, err := tag.Decode([]byte(tagsBin))
		if err != nil {
			return nil, fmt.Errorf("serverHandler.createTagMap failed to decode tagsBin %v: %v", tagsBin, err)
		}
		return tag.NewMap(tag.NewContext(ctx, old), mods...)
	}
	return tag.NewMap(ctx, mods...)
}
