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

package stats

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	istats "github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

var (
	// grpcServerConnKey is the key used to store client instrumentation
	// connection related data into the context.
	grpcServerConnKey *grpcInstrumentationKey
	// grpcServerRPCKey is the key used to store client instrumentation RPC
	// related data into the context.
	grpcServerRPCKey *grpcInstrumentationKey
)

// serverHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC server.
type serverHandler struct{}

// NewServerHandler returns the "google.golang.org/grpc/stats.Handler"
// implementation for the grpc server.
func NewServerHandler() stats.Handler {
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
// it, creates a new github.com/census-instrumentation/opencensus-go/tags.TagsSet,
// adds it to the local context using tagging.NewContextWithTagsSet and finally
// returns the new ctx.
func (sh serverHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	startTime := time.Now()
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("serverHandler.TagRPC called with nil context")
		}
		return ctx
	}
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

	ts, err := sh.createTagSet(ctx, serviceName, methodName)
	if err != nil {
		return ctx
	}
	ctx = tags.NewContext(ctx, ts)

	istats.RecordInt64(ctx, RPCServerStartedCount, 1)
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

	istats.RecordInt64(ctx, RPCServerRequestBytes, int64(s.Length))
	atomic.AddUint64(&d.reqCount, 1)
}

func (sh serverHandler) handleRPCOutPayload(ctx context.Context, s *stats.OutPayload) {
	d, ok := ctx.Value(grpcServerRPCKey).(*rpcData)
	if !ok {
		if glog.V(2) {
			glog.Infoln("serverHandler.handleRPCOutPayload failed to retrieve *rpcData from context")
		}
		return
	}

	istats.RecordInt64(ctx, RPCServerResponseBytes, int64(s.Length))
	atomic.AddUint64(&d.respCount, 1)
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

	var measurements []istats.Measurement
	measurements = append(measurements, RPCServerRequestCount.Is(int64(d.reqCount)))
	measurements = append(measurements, RPCServerResponseCount.Is(int64(d.respCount)))
	measurements = append(measurements, RPCServerFinishedCount.Is(1))
	measurements = append(measurements, RPCServerServerElapsedTime.Is(float64(elapsedTime)/float64(time.Millisecond)))
	if s.Error != nil {
		errorCode := s.Error.Error()
		ts := tags.FromContext(ctx)
		tsb := tags.NewTagSetBuilder(ts)
		tsb.UpsertString(keyOpStatus, errorCode)

		ts = tsb.Build()
		ctx = tags.NewContext(ctx, ts)
		measurements = append(measurements, RPCServerErrorCount.Is(1))
	}

	istats.Record(ctx, measurements...)
}

// createTagSet creates a new tagSet containing the tags extracted from the
// gRPC metadata.
func (sh serverHandler) createTagSet(ctx context.Context, serviceName, methodName string) (*tags.TagSet, error) {
	var tsb tags.TagSetBuilder

	if tagsBin := stats.Tags(ctx); tagsBin == nil {
		tsb = tags.NewTagSetBuilder(nil)
	} else {
		ts, err := tags.Decode([]byte(tagsBin))
		if err != nil {
			return nil, fmt.Errorf("serverHandler.createTagSet failed to decode tagsBin: %v. %v", tagsBin, err)
		}
		tsb = tags.NewTagSetBuilder(ts)
	}
	tsb.UpsertString(keyService, serviceName)
	tsb.UpsertString(keyMethod, methodName)
	return tsb.Build(), nil
}
