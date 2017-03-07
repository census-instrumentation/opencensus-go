package adapter

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Data holds the instrumentation data that is mutated between the start/before
// and end/after the call. It holds all tracing and stats collection/census
// info. All its fields are only accessible by the instrumentation package and
// do not need to be written/modified/read by external packages.
type Data struct {
	isClient, isStream bool

	caller                  string
	methodName, serviceName string
	localAddr, remoteAddr   net.Addr

	reqLen, respLen, wireReqLen, wireRespLen uint32

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

	err error
}

type counter struct {
	c int32
}

func (c *counter) incr() {
	atomic.AddInt32(&c.c, 1)
}

func (c *counter) decr() {
	atomic.AddInt32(&c.c, -1)
}

func (c *counter) ActiveCount() int {
	return int(atomic.LoadInt32(&c.c))
}

// connStatus contains status information for a single conn.
type connStatus struct {
	mu    sync.Mutex
	bns   string // BNS name of the other end
	laddr string
	raddr string
	qos   string // The string form of the QoS (GToS) for the connection.
	// activeCounter returns the number of active requests on this conn.
	activeCounter counter
}

// clientConnStatus is the status of a client connection.
type clientConnStatus struct {
	connStatus
	// Protected by connStatus.mu
	reachable     bool // false when health checking fails
	lameduck      bool
	draining      bool
	sockConnected bool
}

type requestStats struct {
	// These values are uintptrs so we don't have to worry about alignment
	// issues on 32-bit systems.
	atomicCount    uintptr
	atomicNumBytes uintptr
	// Seconds since startTime at which the latest request was received. Not a
	// time.Time so it can be atomic.
	atomicLast uintptr
}

// serverConnStatus is the status of a server connection.
type serverConnStatus struct {
	connStatus

	creationTime time.Time // Used to provide stable temporal ordering on /rpcz/server.
	// Stats for requests, aggregated per type.
	// Currently has space for REQUEST, HEALTH_CHECK, CANCEL_REQUEST,
	// CHANNEL_OPTIONS and STREAM_MESSAGE.
	requestStats [5]requestStats
}

// connData holds connection related data for instrumentation.
type connData struct {
	localAddr, remoteAddr net.Addr
	c                     *counter
	serverConnStatus      *serverConnStatus
	clientConnStatus      *clientConnStatus
}
