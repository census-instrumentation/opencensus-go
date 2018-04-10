package ocgrpc

import (
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func BenchmarkStatusCodeToString_OK(b *testing.B) {
	st := status.New(codes.OK, "OK")
	for i := 0; i < b.N; i++ {
		s := statusCodeToString(st)
		_ = s
	}
}

func BenchmarkStatusCodeToString_Unauthenticated(b *testing.B) {
	st := status.New(codes.Unauthenticated, "Unauthenticated")
	for i := 0; i < b.N; i++ {
		s := statusCodeToString(st)
		_ = s
	}
}

var codeToStringMap = map[codes.Code]string{
	codes.OK:                 "OK",
	codes.Canceled:           "CANCELLED",
	codes.Unknown:            "UNKNOWN",
	codes.InvalidArgument:    "INVALID_ARGUMENT",
	codes.DeadlineExceeded:   "DEADLINE_EXCEEDED",
	codes.NotFound:           "NOT_FOUND",
	codes.AlreadyExists:      "ALREADY_EXISTS",
	codes.PermissionDenied:   "PERMISSION_DENIED",
	codes.ResourceExhausted:  "RESOURCE_EXHAUSTED",
	codes.FailedPrecondition: "FAILED_PRECONDITION",
	codes.Aborted:            "ABORTED",
	codes.OutOfRange:         "OUT_OF_RANGE",
	codes.Unimplemented:      "UNIMPLEMENTED",
	codes.Internal:           "INTERNAL",
	codes.Unavailable:        "UNAVAILABLE",
	codes.DataLoss:           "DATA_LOSS",
	codes.Unauthenticated:    "UNAUTHENTICATED",
}

func BenchmarkMapAlternativeImpl_OK(b *testing.B) {
	st := status.New(codes.OK, "OK")
	for i := 0; i < b.N; i++ {
		s := codeToStringMap[st.Code()]
		_ = s
	}
}
