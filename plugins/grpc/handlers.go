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
	"context"

	"google.golang.org/grpc/stats"
)

// Handlers is a composite type implementing the "google.golang.org/grpc/stats.Handler"
// interface to process lifecycle events from a GRPC client or server. Its only
// purpose is to allow for chaining others types implementing the "google.golang.org/grpc/stats.Handler"
// interface. In this package it allows both stats and tracing subHandlers to be chained together.
type Handlers struct {
	subHandlers []stats.Handler
}

// TagConn can attach some information to the given context. The returned
// context will be used for stats handling. For conn stats handling, the
// context used in HandleConn for this connection will be derived from the
// context returned.
// For client RPC stats handling, the context used in HandleRPC is not derived
// from the context returned.
// For server RPC stats handling, the context used in HandleRPC for all RPCs on
// this connection will be derived from this returned context.
func (h Handlers) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	for _, sh := range h.subHandlers {
		ctx = sh.TagConn(ctx, info)
	}
	return ctx
}

// HandleConn calls all registered connection subHandlers.
func (h Handlers) HandleConn(ctx context.Context, s stats.ConnStats) {
	for _, sh := range h.subHandlers {
		sh.HandleConn(ctx, s)
	}
}

// TagRPC can attach some information to the given context. The returned
// context is used in the rest lifetime of the RPC.
func (h Handlers) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	for _, sh := range h.subHandlers {
		ctx = sh.TagRPC(ctx, info)
	}
	return ctx
}

// HandleRPC calls all registered RPC subHandlers.
func (h Handlers) HandleRPC(ctx context.Context, s stats.RPCStats) {
	for _, sh := range h.subHandlers {
		sh.HandleRPC(ctx, s)
	}
}

// AddHandler adds a handler to the list of registered subHandlers. This list
// contains all the subhandlers that will be called during HandleConn and
// HandleRPC.
func (h Handlers) AddHandler(sh stats.Handler) {
	h.subHandlers = append(h.subHandlers, sh)
}
