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

package collection

import (
	"github.com/golang/glog"

	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

type ServerHandler struct{}

// TagConn can attach some information to the given context.
// The returned context will be used for stats handling.
// For conn stats handling, the context used in HandleConn for this
// connection will be derived from the context returned.
// For RPC stats handling,
//  - On server side, the context used in HandleRPC for all RPCs on this
// connection will be derived from the context returned.
//  - On client side, the context is not derived from the context returned.
func (sh ServerHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	c, err := handleConnServerContext(ctx, info)
	if err != nil {
		return ctx
	}
	return c
}

// HandleConn processes the Conn stats.
func (sh ServerHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch st := s.(type) {
	case *stats.ConnBegin:
		// Do nothing with ConnBegin now.
		return
	case *stats.ConnEnd:
		handleConnServerEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

// TagRPC can attach some information to the given context.
// The returned context is used in the rest lifetime of the RPC.
// HandleRPCServerContext gets the metadata from context and extracts census tags
// and tracing span from it. Then it creates the local trace span and the
// census handle context.Handle, it adds them to the local context using the
// keys census.Key and tracekey.Key, starts the span and finally returns the
// new ctx.
func (sh ServerHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	c, err := handleRPCServerContext(ctx, info)
	if err != nil {
		return ctx
	}
	return c

}

// HandleRPC processes the RPC stats.
func (sh ServerHandler) HandleRPC(ctx context.Context, s stats.RPCStats) {
	switch st := s.(type) {
	case *stats.Begin, *stats.InTrailer, *stats.OutHeader, *stats.OutTrailer:
		// Do nothing for server
	case *stats.InHeader:
		handleRPCServerInHeader(ctx, st)
	case *stats.InPayload:
		handleRPCServerInPayload(ctx, st)
	case *stats.OutPayload:
		handleRPCServerOutPayload(ctx, st) // For stream it can be called multiple times.
	case *stats.End:
		handleRPCServerEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

// GenerateServerTrailer records the elapsed time of the RPC in Data,
// and generates the server trailer metadata that needs to be sent
// to the client.
// It's intended to be called in server interceptor.
func (sh ServerHandler) GenerateServerTrailer(ctx context.Context) (metadata.MD, error) {
	return generateRPCServerTrailer(ctx)
}
