package adapter

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/stats"
)

func handleConnClientContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("handleConnClientContext called with nil context")
	}

	if info.RemoteAddr == nil || info.LocalAddr == nil {
		return ctx, errors.New("handleConnClientContext called with nil info.RemoteAddr or nil info.LocalAddr")
	}

	ctx = context.WithValue(ctx, grpcInstConnKey, &clientConnStatus{
		connData: &connData{
			mu:           sync.Mutex{},
			creationTime: time.Now(),
			localAddr:    info.LocalAddr,
			remoteAddr:   info.RemoteAddr,
		},
	})
	return ctx, nil
}

func handleConnBeginClient(ctx context.Context, s *stats.ConnBegin) error {
	_, ok := ctx.Value(grpcInstConnKey).(*clientConnStatus)
	if !ok {
		return errors.New("handleConnBeginClient cannot retrieve *clientConnStatus from context")
	}
	// TODO(acetechnologist): use clientConnStatus
	return nil
}

func handleConnEndClient(ctx context.Context, s *stats.ConnEnd) error {
	_, ok := ctx.Value(grpcInstConnKey).(*clientConnStatus)
	if !ok {
		return errors.New("handleConnBeginClient cannot retrieve *clientConnStatus from context")
	}
	// TODO(acetechnologist): use clientConnStatus
	return nil
}
