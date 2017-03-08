package adapter

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/stats"
)

func handleConnServerContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("handleConnServerContext called with nil context")
	}

	if info.RemoteAddr == nil || info.LocalAddr == nil {
		return ctx, errors.New("handleConnServerContext called with nil info.RemoteAddr or nil info.LocalAddr")
	}

	ctx = context.WithValue(ctx, grpcInstConnKey, &serverConnStatus{
		connData: &connData{
			mu:           sync.Mutex{},
			creationTime: time.Now(),
			localAddr:    info.LocalAddr,
			remoteAddr:   info.RemoteAddr,
		},
	})
	return ctx, nil
}

func handleConnBeginServer(ctx context.Context, s *stats.ConnBegin) error {
	_, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleConnBeginServer cannot retrieve *serverConnStatus from context")
	}
	// TODO(acetechnologist): use serverConnStatus
	return nil
}

func handleConnEndServer(ctx context.Context, s *stats.ConnEnd) error {
	_, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleConnEndServer cannot retrieve *serverConnStatus from context")
	}
	// TODO(acetechnologist): use serverConnStatus
	return nil
}
