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
	"golang.org/x/net/context"

	"github.com/golang/glog"
	"google.golang.org/grpc/stats"
)

type ClientHandler struct{}

func (ch ClientHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	c, err := handleConnClientContext(ctx, info)
	if err != nil {
		return ctx
	}
	return c
}

func (ch ClientHandler) HandleConn(ctx context.Context, s stats.ConnStats) {
	switch st := s.(type) {
	case *stats.ConnBegin:
		// Do nothing with ConnBegin now.
	case *stats.ConnEnd:
		handleConnClientEnd(ctx, st)
	default:
		glog.Infof("unexpected stats: %T", st)
	}
}

// TagRPC gets the application code census tags and tracing info
// and serializes them into the gRPC metadata in order to be sent to the
// server. This is intended to be used as stats.RPCTagger.
func (ch ClientHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	c, err := handleRPCClientContext(ctx, info)
	if err != nil {
		return ctx
	}
	return c
}

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
