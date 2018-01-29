// Package grpc is a package to assist with tracing incoming and outgoing gRPC requests.
//
// TODO: more documentation, when finalized.
package grpc

import (
	"strings"

	"github.com/census-instrumentation/opencensus-go/trace"
	"github.com/census-instrumentation/opencensus-go/trace/propagation"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

// ClientStatsHandler is a an implementation of grpc.StatsHandler.
type ClientStatsHandler struct {
}

var _ stats.Handler = &ClientStatsHandler{}

// ServerStatsHandler is a an implementation of grpc.StatsHandler.
type ServerStatsHandler struct {
}

var _ stats.Handler = &ServerStatsHandler{}

// NewClientStatsHandler returns a StatsHandler that can be passed to grpc.Dial
// using grpc.WithStatsHandler to enable trace context propagation and
// automatic span creation for outgoing gRPC requests.
func NewClientStatsHandler() *ClientStatsHandler {
	return &ClientStatsHandler{}
}

// NewServerStatsHandler returns a StatsHandler that can be passed to
// grpc.NewServer using grpc.StatsHandler to enable trace context propagation
// and automatic span creation for incoming gRPC requests.
func NewServerStatsHandler() *ServerStatsHandler {
	return &ServerStatsHandler{}
}

// TODO: test interoperability with other languages
const traceContextKey = "grpc-trace-bin"

// TagRPC creates a new trace span for the client side of the RPC.
//
// It returns ctx with the new trace span added and a serialization of the
// SpanContext added to the outgoing gRPC metadata.
func (c *ClientStatsHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	name := "Sent" + strings.Replace(rti.FullMethodName, "/", ".", -1)
	ctx = trace.StartSpanWithOptions(ctx, name, trace.StartSpanOptions{RecordEvents: true, RegisterNameForLocalSpanStore: true})
	traceContextBinary := propagation.Binary(trace.FromContext(ctx).SpanContext())
	if len(traceContextBinary) == 0 {
		return ctx
	}
	md := metadata.Pairs(traceContextKey, string(traceContextBinary))
	if oldMD, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(oldMD, md)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// TagRPC creates a new trace span for the server side of the RPC.
//
// It checks the incoming gRPC metadata in ctx for a SpanContext, and if
// it finds one, uses that SpanContext as the parent context of the new span.
//
// It returns ctx, with the new trace span added.
func (s *ServerStatsHandler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	name := "Recv" + strings.Replace(rti.FullMethodName, "/", ".", -1)
	opt := trace.StartSpanOptions{RecordEvents: true, RegisterNameForLocalSpanStore: true}
	if s := md[traceContextKey]; len(s) > 0 {
		// TODO: add any additional values as parent links?
		if parent, ok := propagation.FromBinary([]byte(s[0])); ok {
			return trace.StartSpanWithRemoteParent(ctx, name, parent, opt)
		}
	}
	return trace.StartSpanWithOptions(ctx, name, opt)
}

// HandleRPC processes the RPC stats, adding information to the current trace span.
func (c *ClientStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	handleRPC(ctx, rs)
}

// HandleRPC processes the RPC stats, adding information to the current trace span.
func (s *ServerStatsHandler) HandleRPC(ctx context.Context, rs stats.RPCStats) {
	handleRPC(ctx, rs)
}

func handleRPC(ctx context.Context, rs stats.RPCStats) {
	// TODO: compressed and uncompressed sizes are not populated in every message.
	switch rs := rs.(type) {
	case *stats.Begin:
		trace.SetSpanAttributes(ctx,
			trace.BoolAttribute{Key: "Client", Value: rs.Client},
			trace.BoolAttribute{Key: "FailFast", Value: rs.FailFast})
	case *stats.InPayload:
		trace.AddMessageReceiveEvent(ctx, 0 /* TODO: messageID */, int64(rs.Length), int64(rs.WireLength))
	case *stats.InHeader:
		trace.AddMessageReceiveEvent(ctx, 0, int64(rs.WireLength), int64(rs.WireLength))
	case *stats.InTrailer:
		trace.AddMessageReceiveEvent(ctx, 0, int64(rs.WireLength), int64(rs.WireLength))
	case *stats.OutPayload:
		trace.AddMessageSendEvent(ctx, 0, int64(rs.Length), int64(rs.WireLength))
	case *stats.OutHeader:
		trace.AddMessageSendEvent(ctx, 0, 0, 0)
	case *stats.OutTrailer:
		trace.AddMessageSendEvent(ctx, 0, int64(rs.WireLength), int64(rs.WireLength))
	case *stats.End:
		if rs.Error != nil {
			code, desc := grpc.Code(rs.Error), grpc.ErrorDesc(rs.Error)
			trace.SetSpanStatus(ctx, trace.Status{Code: int32(code), Message: desc})
		}
		trace.EndSpan(ctx)
	}
}

// TagConn is a no-op for this StatsHandler.
func (c *ClientStatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return ctx
}

// TagConn is a no-op for this StatsHandler.
func (s *ServerStatsHandler) TagConn(ctx context.Context, cti *stats.ConnTagInfo) context.Context {
	return ctx
}

// HandleConn is a no-op for this StatsHandler.
func (c *ClientStatsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
}

// HandleConn is a no-op for this StatsHandler.
func (s *ServerStatsHandler) HandleConn(ctx context.Context, cs stats.ConnStats) {
}
