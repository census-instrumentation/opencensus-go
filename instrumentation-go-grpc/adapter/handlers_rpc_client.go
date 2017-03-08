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
	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

func handleRPCContextClient(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	startTime := time.Now()
	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		return nil, fmt.Errorf("handleRPCClientContext called with info.FullMethodName bad format: %v", info.FullMethodName)
	}

	d := &rpcData{
		startTime:   startTime,
		serviceName: names[1],
		methodName:  names[2],
	}

	ts := tagging.FromContext(ctx)
	encoded := tagging.EncodeToFullSignature(ts)

	statsCtx := &pb.StatsContext{
		Tags: encoded,
	}
	statsBin, err := proto.Marshal(statsCtx)
	if err != nil {
		return nil, fmt.Errorf("handleRPCClientContext cannot marshal pb.StatsContext %v\n. %v", statsCtx, err)
	}

	// Join old metadata with new one, so the old metadata information is not lost.
	md, _ := metadata.FromContext(ctx)
	ctx = metadata.NewContext(ctx, metadata.Join(md, metadata.Pairs(statsKey, string(statsBin))))

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

	// TODO(acetchnologist): add tracing to metadata here
	//ctx = metadata.NewContext(ctx, metadata.Join(md, metadata.Pairs(traceKey, string(traceBin), censusTagsKey, string(censusBin))))

	return context.WithValue(ctx, grpcInstRPCKey, d), nil
}

func handleRPCBeginClient(ctx context.Context, s *stats.Begin) error {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCBeginClient failed to extract *rpcData")
	}

	d.isClient = true

	// TODO(acetechnologist): call requestz started
	// TODO(acetechnologist): streamz.
	return nil
}

func handleRPCOutHeaderClient(ctx context.Context, s *stats.OutHeader) error {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleOutHeaderClient failed to extract *rpcData")
	}

	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr
	// TODO(acetechnologist): RequestInfoUpdate(d)
	return nil
}

func handleRPCOutPayloadClient(ctx context.Context, s *stats.OutPayload) error {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleOutPayloadClient failed to extract *rpcData")
	}
	atomic.AddInt32(&d.reqLen, int32(s.Length))
	atomic.AddInt32(&d.wireReqLen, int32(s.WireLength))

	// TODO(menghanl): uncomment the following line if it's needed for client side lb load reporting.
	// atomic.AddUint32(&d.reqCount, 1)

	// argumentType, ok := s.Payload.(proto.Message)
	// if !ok {
	// 	return fmt.Errorf("handleRPCOutPayloadClient failed to extract argumentType. s.Payload is of type %T want type proto.Message", s.Payload)
	// }
	// payload := &rpctrace.Payload{
	// 	Pay:        s.Data,
	// 	PayLen:     s.Length,
	// 	WirePayLen: s.WireLength,
	// }
	// d.payloadReq = payload

	// RequestzPayload(d, payload, argumentType)
	return nil
}

func handleRPCInPayloadClient(ctx context.Context, s *stats.InPayload) error {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleInPayloadClient failed to extract *rpcData")
	}
	atomic.AddInt32(&d.respLen, int32(s.Length))
	atomic.AddInt32(&d.wireRespLen, int32(s.WireLength))

	// TODO(menghanl): uncomment the following line if it's needed for client side lb load reporting.
	// atomic.AddUint32(&d.respCount, 1)

	// TODO(acetechnologist):
	// argumentType, ok := s.Payload.(proto.Message)
	// if !ok {
	// 	return fmt.Errorf("handleRPCInPayloadServer failed to extract argumentType. s.Payload is of type %T want type proto.Message", s.Payload)
	// }
	// payload := &rpctrace.Payload{
	// 	Pay:        s.Data,
	// 	PayLen:     s.Length,
	// 	WirePayLen: s.WireLength,
	// }
	// d.payloadResp = payload
	return nil
}

func handleRPCEndClient(ctx context.Context, s *stats.End) error {
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleEndClient failed to extract *rpcData")
	}
	d.err = s.Error

	d.totalElapsedTime = time.Since(d.startTime)

	var measurements []istats.Measurement
	measurements = append(measurements, measureRPCReqLen.CreateMeasurement(float64(d.reqLen)))
	measurements = append(measurements, measureRPCRespLen.CreateMeasurement(float64(d.respLen)))
	measurements = append(measurements, measureRPCElapsed.CreateMeasurement(float64(d.serverElapsedTime)/float64(time.Millisecond)))

	if d.err != nil {
		measurements = append(measurements, measureRPCError.CreateMeasurement(1))
	}

	istats.RecordMeasurements(ctx, measurements...)

	// TODO(acetechnologist):
	// d.span.Finish()
	// reportStreamzClientDataEnd(d)
	// DapperRequestPayload(ctx, d)
	// status := util.ErrorToStatus(s.Error)
	// DapperResponsePayload(ctx, d, status)
	// RpczClientFinish(d, status)
	// RequestzFinish(d, status)
	return nil
}
