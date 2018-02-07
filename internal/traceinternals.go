package internal

import "time"

// TraceInternal allows internal access to some trace functionality.
// TODO(#412): remove this
var TraceInternal interface{}


// BucketConfiguration stores the number of samples to store for span buckets
// for successful and failed spans for a particular span name.
type BucketConfiguration struct {
	Name                 string
	MaxRequestsSucceeded int
	MaxRequestsErrors    int
}

// PerMethodSummary is a summary of the spans stored for a single span name.
type PerMethodSummary struct {
	Active         int
	LatencyBuckets []LatencyBucketSummary
	ErrorBuckets   []ErrorBucketSummary
}

// LatencyBucketSummary is a summary of a latency bucket.
type LatencyBucketSummary struct {
	MinLatency, MaxLatency time.Duration
	Size                   int
}

// ErrorBucketSummary is a summary of an error bucket.
type ErrorBucketSummary struct {
	ErrorCode int32
	Size      int
}
