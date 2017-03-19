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
	"google.golang.org/grpc/stats"
)

// ClientHandler is the type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from the GRPC client.
type ClientHandler struct{}

// TagConn can attach some information to the given context. For RPC stats
// handling the context used in HandleRPC is not derived from the context
// returned.
func (ch ClientHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	if ctx == nil {
		if glog.V(2) {
			glog.Infoln("ClientHandler.TagConn(_) called with nil context")
		}
		return ctx
	}

	if info.RemoteAddr == nil || info.LocalAddr == nil {
		if glog.V(2) {
			glog.Infoln("ClientHandler.TagConn(_) called with nil info.RemoteAddr or nil info.LocalAddr")
		}
		return ctx
	}

	ctx = context.WithValue(ctx, grpcInstConnKey, &clientConnStatus{
		connData: &connData{
			mu:           sync.Mutex{},
			creationTime: time.Now(),
			localAddr:    info.LocalAddr,
			remoteAddr:   info.RemoteAddr,
		},
	})
	return ctx
}

// HandleConn processes the Conn events.
func (ch ClientHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	_, ok := ctx.Value(grpcInstConnKey).(*clientConnStatus)
	if !ok {
		if glog.V(2) {
			glog.Infoln("ClientHandler.HandleConn(_) couldn't retrieve *clientConnStatus from context")
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
			glog.Infoln("ClientHandler.HandleConn(_) called with uenxpected stats.ConnStats %T", st)
		}
	}
	return
}

// TagRPC gets the tagging.TagsSet set by the application code, serializes its
// tags into the GRPC metadata in order to be sent to the server.
func (ch ClientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	return handleRPCClientContext(ctx, info)
}

// HandleRPC processes the RPC events.
func (ch ClientHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.InHeader, *stats.InTrailer, *stats.OutTrailer:
		// do nothing for client
	case *stats.Begin:
		handleRPCClientBegin(ctx, st)
	case *stats.InPayload:
		handleRPCClientInPayload(ctx, st)
	case *stats.OutHeader:
		handleRPCClientOutHeader(ctx, st)
	case *stats.OutPayload:
		handleRPCClientOutPayload(ctx, st)
	case *stats.End:
		handleRPCClientEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}
