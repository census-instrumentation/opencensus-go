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
	"sync"
	"time"

	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

// ServerHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC server.
type ServerHandler struct{}

// TagConn can attach some information to the given context. For RPC stats
// handling, the context used in HandleRPC for all RPCs on this connection will
// be derived from the context returned.
func (sh ServerHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("ServerHandler.TagConn(_) called with nil context")
		}
		return ctx
	}

	if info.RemoteAddr == nil || info.LocalAddr == nil {
		if glog.V(2) {
			glog.Infoln("ServerHandler.TagConn(_) called with nil info.RemoteAddr or nil info.LocalAddr")
		}
		return ctx
	}

	ctx = context.WithValue(ctx, grpcInstConnKey, &serverConnStatus{
		connData: &connData{
			mu:           sync.Mutex{},
			creationTime: time.Now(),
			localAddr:    info.LocalAddr,
			remoteAddr:   info.RemoteAddr,
		},
	})
	return ctx
}

// HandleConn processes the connection events.
func (sh ServerHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	_, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		if glog.V(2) {
			glog.Infoln("ServerHandler.HandleConn(_) couldn't retrieve *serverConnStatus from context")
		}
		return
	}

	switch st := s.(type) {
	case *stats.ConnBegin:
		// Do nothing with ConnBegin now.
	case *stats.ConnEnd:
		// Do nothing with ConnEnd now.
	default:
		if glog.V(2) {
			glog.Infoln("ServerHandler.HandleConn(_) called with uenxpected stats.ConnStats %T", st)
		}
	}
	return
}

// TagRPC just forwards the call to handleRPCServerContext. This is necessary
// because this package implementation of RPC handling methds is a lot more
// granular the interface requirements by "google.golang.org/grpc/stats.Handler".
func (sh ServerHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return handleRPCServerContext(ctx, info)
}

// HandleRPC processes the RPC events.
func (sh ServerHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.Begin, *stats.InTrailer, *stats.OutHeader, *stats.OutTrailer:
		// Do nothing for server
	case *stats.InHeader:
		handleRPCServerInHeader(ctx, st)
	case *stats.InPayload:
		handleRPCServerInPayload(ctx, st)
	case *stats.OutPayload:
		// For stream it can be called multiple times per RPC.
		handleRPCServerOutPayload(ctx, st)
	case *stats.End:
		handleRPCServerEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

// GenerateServerTrailer just forwards the call to handleRPCServerContext. It
// is intended to be called in server interceptor.
func (sh ServerHandler) GenerateServerTrailer(ctx context.Context) (metadata.MD, error) {
	return generateRPCServerTrailer(ctx)
}
