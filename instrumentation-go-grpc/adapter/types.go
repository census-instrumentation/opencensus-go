package adapter

import (
	"net"
	"sync"
	"time"
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

// type counter32 int32

// func (c *counter32) incr(i int32) {
// 	atomic.AddInt32(c, i)
// }

// func (c *counter32) ActiveCount() int32 {
// 	return int32(atomic.LoadInt32(c))
// }

// type counter64 int64

// func (c *counter64) incr(i int64) {
// 	atomic.AddInt64(c, i)
// }

// func (c *counter64) ActiveCount() int64 {
// 	return int64(atomic.LoadInt64(c))
// }
