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

// Package stats defines the two handlers (ClientHandler and ServerHandler) for
// processing GRPC lifecycle events and process statistics data. Both are
// different implementations of the "google.golang.org/grpc/stats.Handler"
// interface. This package also defines the default metrics (a.k.a.
// measurements) to be collected by the instrumentation package from GRPC. Once
// collected, they can be exported to any external package or system.
// "github.com/google/instrumentation-go/grpc-plugin/export"" defines a sample
// server for exporting these metrics.
package stats

import (
	"net"
	"sync"
	"time"
)

// statsKey is the metadata key used to identify both the stats tags in the
// GRPC context metadata, as well as the RpcServerStats info sent back from
// the server to the client in the GRPC context metadata.
const statsKey = "grpc-stats-bin"

type grpcRPCKey struct{}
type grpcConnKey struct{}

var (
	// grpcInstConnKey is the key used to store connection related data to
	// context.
	grpcInstConnKey grpcConnKey
	// grpcInstRPCKey is the key used to store RPC related data to context.
	grpcInstRPCKey grpcRPCKey
)

// rpcData holds the instrumentation RPC data that is needed between the start
// and end of an call. It holds the info that this package needs to keep track
// of between the various GRPC events.
type rpcData struct {
	isClient, isStream bool

	methodName, serviceName string
	localAddr, remoteAddr   net.Addr

	reqLen, respLen, wireReqLen, wireRespLen int32

	// sequence number if streaming RPC.
	sequenceNumber int32

	// deadline is the GRPC deadline for this call.
	deadline time.Time

	// startTime as observed on the server side. It represents the time at
	// which the server started processing the request. In reality it is the
	// time at which GRPC invoked stats.handleRPCServerContext.
	startTime time.Time

	// elapsedTime as observed on the server side. This cannot be populated
	// until after the call completes.
	serverElapsedTime time.Duration

	// elapsedTime as observed on the server side. This is computed after the
	// call completes and GRPC invokes stats.generateRPCServerTrailer
	totalElapsedTime time.Duration

	authProtocol string
	err          error
}

// connData holds the instrumentation connection data that is needed between
// the start and end of a connection.
type connData struct {
	mu                    sync.Mutex
	creationTime          time.Time
	localAddr, remoteAddr net.Addr
	// activeRequests tracks the number of active requests on this connection.
	activeRequests int32
}

// clientConnStatus contains the status of a client connection.
type clientConnStatus struct {
	*connData
	// false when health checking fails
	reachable bool

	// if connection is active but in lameduck (not in use for now).
	lameduck      bool
	draining      bool
	sockConnected bool
}

// serverConnStatus contains the status of a server connection.
type serverConnStatus struct {
	*connData
	requests       requestStats
	healthChecks   requestStats
	cancelRequets  requestStats
	streamMessages requestStats
}

type requestStats struct {
	count    int64
	numBytes int64
}
