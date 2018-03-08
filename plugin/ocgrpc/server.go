// Copyright 2018, OpenCensus Authors
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

package ocgrpc

import (
	"go.opencensus.io/trace"
	"golang.org/x/net/context"

	"google.golang.org/grpc/stats"
)

// ServerHandler implements gRPC stats.Handler recording OpenCensus stats and
// traces. Use with gRPC servers.
//
// When installed (see Example), tracing metadata is read from inbound RPCs
// by default. If no tracing metadata is present, or if the tracing metadata is
// present but the SpanContext isn't sampled, then a new trace may be started
// (as determined by Sampler).
type ServerHandler struct {
	// NoTrace may be set to true to disable OpenCensus tracing integration.
	// If set to true, no trace metadata will be read from inbound RPCs and no
	// new Spans will be created.
	NoTrace bool

	// NoStats may be set to true to disable recording OpenCensus stats for RPCs.
	NoStats bool

	// StartNewTraces may be set to true to start a new trace, wrapping each
	// RPC with a root Span.
	//
	// You should set this if your gRPC server is a public-facing service.
	//
	// Be aware that if you leave this false (the default) on a public-facing
	// server, callers will be able to send tracing metadata in gRPC headers
	// and trigger traces in your backend.
	StartNewTraces bool

	// Sampler to use for RPCs handled by this server. This will be called for
	// each RPC that is not already sampled (assuming NoStats is not set to true).
	//
	// In particular, this will be called even if there is tracing metadata
	// present on the inbound RPC, but the SpanContext is not sampled. This
	// ensures that each service has some opportunity to be traced. If you would
	// like to not add any additional traces for this gRPC service, use:
	//   trace.ProbabilitySampler(0.0)
	//
	// If not set, the default sampler will be used (see trace.SetDefaultSampler).
	Sampler trace.Sampler
}

var _ stats.Handler = (*ServerHandler)(nil)

// HandleConn exists to satisfy gRPC stats.Handler.
func (s *ServerHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
	// no-op
}

// TagConn exists to satisfy gRPC stats.Handler.
func (s *ServerHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	// no-op
	return ctx
}

// HandleRPC implements per-RPC tracing and stats instrumentation.
func (s *ServerHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	if !s.NoTrace {
		s.traceHandleRPC(ctx, rs)
	}
	if !s.NoStats {
		s.statsHandleRPC(ctx, rs)
	}
}

// TagRPC implements per-RPC context management.
func (s *ServerHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	if !s.NoTrace {
		ctx = s.traceTagRPC(ctx, rti)
	}
	if !s.NoStats {
		ctx = s.statsTagRPC(ctx, rti)
	}
	return ctx
}
