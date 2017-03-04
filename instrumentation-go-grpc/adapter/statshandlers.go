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

	// lbTokenKey is the metadata key used for lb token.
	lbTokenKey = "lb-token"

	// statsKey is the metadata key used to identify both the census tags in
	// the gRPC metadata context as well as RpcServerStats info sent back from
	// the server to the client in the gRPC metadata context.
	statsKey = "grpc-stats-bin"
)

type grpcRPCKey struct{}
type grpcConnKey struct{}

var (
	// grpcInstKey is the key used to store RPC related data to context.
	grpcInstKey grpcRPCKey
	// grpcInstConnKey is the key used to store connection related data to context.
	grpcInstConnKey grpcConnKey
)

// ServerConnContextHandler adds connection related data to the context and returns
// the new context.
func ServerConnContextHandler(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	return handleServerConnContext(ctx, info)
}

// ClientConnContextHandler adds connection related data to the context and returns
// the new context.
func ClientConnContextHandler(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	return handleClientConnContext(ctx, info)
}

// HandleConnEnd records measurements for a completed connection.
func HandleConnEnd(ctx context.Context, s *stats.ConnEnd) error {
	if s.IsClient() {
		return handleConnEndClient(ctx, s)
	}
	return handleConnEndServer(ctx, s)
}

// ServerRPCContextHandler gets the metadata from context and extracts census tags
// and tracing span from it. Then it creates the local trace span and the
// census handle context.Handle, it adds them to the local context using the
// keys census.Key and tracekey.Key, starts the span and finally returns the
// new ctx.
func ServerRPCContextHandler(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	return handleServerRPCContext(ctx, info)
}

// ClientRPCContextHandler gets the application code census tags and tracing info
// and serializes them into the gRPC metadata in order to be sent to the
// server. This is intended to be used as stats.RPCTagger.
func ClientRPCContextHandler(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	return handleClientContext(ctx, info)
}

// HandleBegin processes the beginning of an RPC.
func HandleBegin(ctx context.Context, s *stats.Begin) error {
	if !s.IsClient() {
		return nil
	}
	return handleBeginClient(ctx, s)

}

// HandleInHeader processes the inbound header of an RPC. On the server side it
// is called before HandleBegin.
func HandleInHeader(ctx context.Context, s *stats.InHeader) error {
	if s.IsClient() {
		return nil
	}

	return handleInHeaderServer(ctx, s)
}

// HandleInPayload processes the inbound payload of an RPC. For stream it can
// be called multiple times.
func HandleInPayload(ctx context.Context, s *stats.InPayload) error {
	if s.IsClient() {
		return handleInPayloadClient(ctx, s)
	}
	return handleInPayloadServer(ctx, s)
}

// HandleInTrailer processes the trailer of an RPC after it is received.
func HandleInTrailer(ctx context.Context, s *stats.InTrailer) error {
	return nil
}

// HandleOutHeader processes the outbound header of an RPC.
func HandleOutHeader(ctx context.Context, s *stats.OutHeader) error {
	if !s.IsClient() {
		return nil
	}
	return handleOutHeaderClient(ctx, s)
}

// HandleOutPayload processes the outbound payload of an RPC. For stream it can
// be called multiple times.
func HandleOutPayload(ctx context.Context, s *stats.OutPayload) error {
	if s.IsClient() {
		return handleOutPayloadClient(ctx, s)
	}
	return handleOutPayloadServer(ctx, s)
}

// GenerateServerTrailer records the elapsed time of the RPC in Data,
// and generates the server trailer metadata that needs to be sent
// to the client.
// It's intended to be called in server interceptor.
func GenerateServerTrailer(ctx context.Context) (metadata.MD, error) {
	return generateServerTrailer(ctx)
}

// HandleOutTrailer processes the trailer of an RPC after it is sent.
func HandleOutTrailer(ctx context.Context, s *stats.OutTrailer) error {
	return nil
}

// HandleEnd records measurements for a completed gRPC call for the
// ResourceId_RPC_SERVER resource. It is called whenever an RPC is finished.
func HandleEnd(ctx context.Context, s *stats.End) error {
	if s.IsClient() {
		return handleEndClient(ctx, s)
	}
	return handleEndServer(ctx, s)
}
