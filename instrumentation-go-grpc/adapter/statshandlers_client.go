package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	pb "github.com/google/instrumentation-go/instrumentation-go-grpc/generated_proto"
	//istats "github.com/google/instrumentation-go/stats/"
	"github.com/google/instrumentation-go/stats/tagging"
)

func handleClientConnContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("handleClientConnContext called with nil context")
	}

	if info.RemoteAddr == nil {
		return ctx, errors.New("failed to collect connection stats: info.RemoteAddr is nil")
	}

	c := &counter{}
	ccs := rpcstats.DefaultManager().NewClientConnStatus(info.RemoteAddr.String())
	ctx = context.WithValue(ctx, grpcInstConnKey, &connData{
		localAddr:        info.LocalAddr,
		remoteAddr:       info.RemoteAddr,
		c:                c,
		clientConnStatus: ccs,
	})
	return ctx, nil
}

// TODO(menghanl): handle signals about conn connectivity when they are available.

func handleConnEndClient(ctx context.Context, s *stats.ConnEnd) error {
	cd, ok := ctx.Value(grpcInstConnKey).(*connData)
	if !ok {
		return errors.New("*connData cannot be retrieved from context")
	}
	rpcstats.DefaultManager().RemoveClientConnStatus(cd.clientConnStatus)
	return nil
}

func handleClientContext(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	d := &Data{
		startTime: time.Now(),
	}

	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		return nil, fmt.Errorf("info.FullMethodName bad format: %v", info.FullMethodName)
	}
	serviceName := names[1]
	methodName := names[2]

	// TODO(acetchnologist): add tracing hooks here
	// parentSpan, ok := trace.FromContext(ctx)
	// if !ok {
	// 	// it is ok not having a trace in the context
	// 	parentSpan = nil
	// }
	// // Creating copy of context with local tracing span
	// span := &trace.Span{}
	// span.Start(parentSpan, trace.ClientType, methodName)
	// traceCtxPb := &ccpb.GoogleTraceContext{
	// 	TraceIdLo:                  span.TraceID,
	// 	SpanId:                     span.ID,
	// 	SpanOptions:                uint32(span.Mask),
	// 	ParentSpanId:               span.ParentID,
	// 	InverseSamplingProbability: span.SamplingProbability,
	// }
	// traceBin, err := proto.Marshal(traceCtxPb)
	// if err != nil {
	// 	return nil, fmt.Errorf("cannot marshal tracing context proto %v: %v", traceCtxPb, err)
	// }
	// d.span = span
	// d.family = "Sent." + serviceName

	ts := tagging.FromContext(ctx)
	encoded, err := encodeToGrpcFormat(ts)
	if err != nil {
		return nil, fmt.Errorf("while encodeToGrpcFormat in handleClientContext: %v", err)
	}
	d.methodName = methodName
	d.serviceName = serviceName

	statsCtx := &pb.StatsContext{
		Tags: encoded,
	}
	statsBin, err := proto.Marshal(statsCtx)
	if err != nil {
		return nil, fmt.Errorf("Cannot marshal stats context proto %v: %v", statsCtx, err)
	}
	// Join old metadata with new one, so the old metadata information is not lost.
	md, _ := metadata.FromContext(ctx)
	ctx = metadata.NewContext(ctx, metadata.Join(md, metadata.Pairs(statsKey, string(statsBin))))

	// TODO(acetchnologist): add tracing to metadata here
	//ctx = metadata.NewContext(ctx, metadata.Join(md, metadata.Pairs(traceKey, string(traceBin), censusTagsKey, string(censusBin))))

	ctx = context.WithValue(ctx, grpcInstKey, d)

	return ctx, nil
}

func handleBeginClient(ctx context.Context, s *stats.Begin) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	d.isClient = true
	d.failFastOption = s.FailFast

	RequestzStart(ctx, d)
	reportStreamzClientDataStart(d)
	return nil
}

func handleOutHeaderClient(ctx context.Context, s *stats.OutHeader) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}
	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr
	RequestInfoUpdate(d)
	return nil
}

func handleOutPayloadClient(ctx context.Context, s *stats.OutPayload) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	// TODO(menghanl): uncomment the following line if it's needed for client side lb load reporting.
	// atomic.AddUint32(&d.reqCount, 1)
	atomic.AddUint32(&d.reqLen, uint32(s.Length))
	atomic.AddUint32(&d.wireReqLen, uint32(s.WireLength))

	argumentType, ok := s.Payload.(proto.Message)
	if !ok {
		return fmt.Errorf("s.Payload is of type %T want type proto.Message", s.Payload)
	}

	payload := &rpctrace.Payload{
		Pay:        s.Data,
		PayLen:     s.Length,
		WirePayLen: s.WireLength,
	}
	d.payloadReq = payload

	RequestzPayload(d, payload, argumentType)
	return nil
}

func handleInPayloadClient(ctx context.Context, s *stats.InPayload) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	// TODO(menghanl): uncomment the following line if it's needed for client side lb load reporting.
	// atomic.AddUint32(&d.respCount, 1)
	atomic.AddUint32(&d.respLen, uint32(s.Length))
	atomic.AddUint32(&d.wireRespLen, uint32(s.WireLength))

	payload := &rpctrace.Payload{
		Pay:        s.Data,
		PayLen:     s.Length,
		WirePayLen: s.WireLength,
	}
	d.payloadResp = payload
	return nil
}

func handleEndClient(ctx context.Context, s *stats.End) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	status := util.ErrorToStatus(s.Error)

	// stubby in Go doesn't record any stats on the client side. To have
	// parity we only need to record census info in the server.
	d.totalElapsedTime = time.Since(d.startTime)

	d.span.Finish()
	reportStreamzClientDataEnd(d)
	DapperRequestPayload(ctx, d)
	DapperResponsePayload(ctx, d, status)
	RpczClientFinish(d, status)
	RequestzFinish(d, status)
	return nil
}
