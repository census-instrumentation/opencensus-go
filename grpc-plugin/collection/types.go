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
	"net"
	"sync"
	"time"
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

// rpcData holds the instrumentation data that is mutated between the start and
// end the call. It holds all tracing and stats collection/census info. All its
// fields are only accessible by the instrumentation package and do not need to
// be written/modified/read by external packages.
type rpcData struct {
	isClient, isStream bool

	methodName, serviceName string
	localAddr, remoteAddr   net.Addr

	reqLen, respLen, wireReqLen, wireRespLen int32

	// sequence number if streaming RPC.
	sequenceNumber int32

	// deadline is the grpc deadline for this call.
	deadline time.Time

	// startTime as observed on the server side. This is the time at which the
	// server started processing the request.
	startTime time.Time

	// elapsedTime as observed on the server side. This cannot be populated
	// until after the call completes.
	serverElapsedTime time.Duration

	// elapsedTime as observed on the server side. This cannot be populated
	// until after the call completes.
	totalElapsedTime time.Duration

	authProtocol string
	err          error
}

// connData holds connection related data for instrumentation.
type connData struct {
	mu                    sync.Mutex
	creationTime          time.Time
	localAddr, remoteAddr net.Addr
	activeRequests        int32 // activeRequests returns the number of active requests on this conn.
}

// clientConnStatus contains the status of a client connection.
type clientConnStatus struct {
	*connData
	reachable     bool // false when health checking fails
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
