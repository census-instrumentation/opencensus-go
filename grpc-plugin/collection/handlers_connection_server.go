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

func handleConnServerBegin(ctx context.Context, s *stats.ConnBegin) error {
	_, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleConnServerBegin cannot retrieve *serverConnStatus from context")
	}
	// TODO(acetechnologist): use serverConnStatus
	return nil
}

func handleConnServerEnd(ctx context.Context, s *stats.ConnEnd) error {
	_, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleConnServerEnd cannot retrieve *serverConnStatus from context")
	}
	// TODO(acetechnologist): use serverConnStatus
	return nil
}
