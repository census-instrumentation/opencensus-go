package adapter

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

const (
	// traceKey is the metadata key used to identify the tracing info in the
	// gRPC metadata context.
	traceKey = "grpc-tracing-bin"

	// statsKey is the metadata key used to identify both the census tags in
	// the gRPC metadata context as well as RpcServerStats info sent back from
	// the server to the client in the gRPC metadata context.
	statsKey = "grpc-stats-bin"
)

type grpcRPCKey struct{}
type grpcConnKey struct{}

var (
	// grpcInstConnKey is the key used to store connection related data to context.
	grpcInstConnKey grpcConnKey
	// grpcInstRPCKey is the key used to store RPC related data to context.
	grpcInstRPCKey grpcRPCKey
)

// HandleConnServerContext adds connection related data to the context and returns
// the new context.
func HandleConnServerContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	return handleConnServerContext(ctx, info)
}

// HandleConnClientContext adds connection related data to the context and returns
// the new context.
func HandleConnClientContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	return handleConnClientContext(ctx, info)
}

// HandleConnEnd records measurements for a completed connection.
func HandleConnEnd(ctx context.Context, s *stats.ConnEnd) error {
	if s.IsClient() {
		return handleConnClientEnd(ctx, s)
	}
	return handleConnServerEnd(ctx, s)
}

// HandleRPCServerContext gets the metadata from context and extracts census tags
// and tracing span from it. Then it creates the local trace span and the
// census handle context.Handle, it adds them to the local context using the
// keys census.Key and tracekey.Key, starts the span and finally returns the
// new ctx.
func HandleRPCServerContext(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	return handleRPCServerContext(ctx, info)
}

// HandleRPCClientContext gets the application code census tags and tracing info
// and serializes them into the gRPC metadata in order to be sent to the
// server. This is intended to be used as stats.RPCTagger.
func HandleRPCClientContext(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	return handleRPCClientContext(ctx, info)
}

// HandleBegin processes the beginning of an RPC.
func HandleBegin(ctx context.Context, s *stats.Begin) error {
	if s.IsClient() {
		return handleRPCClientBegin(ctx, s)
	}
	return nil

}

// HandleInHeader processes the inbound header of an RPC. On the server side it
// is called before HandleBegin.
func HandleInHeader(ctx context.Context, s *stats.InHeader) error {
	if s.IsClient() {
		return nil
	}

	return handleRPCServerInHeader(ctx, s)
}

// HandleInPayload processes the inbound payload of an RPC. For stream it can
// be called multiple times.
func HandleInPayload(ctx context.Context, s *stats.InPayload) error {
	if s.IsClient() {
		return handleRPCClientInPayload(ctx, s)
	}
	return handleRPCServerInPayload(ctx, s)
}

// HandleInTrailer processes the trailer of an RPC after it is received.
func HandleInTrailer(ctx context.Context, s *stats.InTrailer) error {
	return nil
}

// HandleOutHeader processes the outbound header of an RPC.
func HandleOutHeader(ctx context.Context, s *stats.OutHeader) error {
	if s.IsClient() {
		return handleRPCClientOutHeader(ctx, s)
	}
	return nil
}

// HandleOutPayload processes the outbound payload of an RPC. For stream it can
// be called multiple times.
func HandleOutPayload(ctx context.Context, s *stats.OutPayload) error {
	if s.IsClient() {
		return handleRPCClientOutPayload(ctx, s)
	}
	return handleRPCServerOutPayload(ctx, s)
}

// GenerateServerTrailer records the elapsed time of the RPC in Data,
// and generates the server trailer metadata that needs to be sent
// to the client.
// It's intended to be called in server interceptor.
func GenerateServerTrailer(ctx context.Context) (metadata.MD, error) {
	return generateRPCServerTrailer(ctx)
}

// HandleOutTrailer processes the trailer of an RPC after it is sent.
func HandleOutTrailer(ctx context.Context, s *stats.OutTrailer) error {
	return nil
}

// HandleEnd records measurements for a completed gRPC call for the
// ResourceId_RPC_SERVER resource. It is called whenever an RPC is finished.
func HandleEnd(ctx context.Context, s *stats.End) error {
	if s.IsClient() {
		return handleRPCClientEnd(ctx, s)
	}
	return handleRPCServerEnd(ctx, s)
}
