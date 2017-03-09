package adapter

import (
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"

	pb "github.com/google/instrumentation-go/instrumentation-go-grpc/generated_proto"
	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

func handleRPCServerContext(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	startTime := time.Now()
	if ctx == nil {
		return nil, errors.New("handleRPCServerContext called with nil context")
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.New("handleRPCServerContext failed to extract metadata")
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("handleRPCServerContext failed to extract peer info")
	}

	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		return nil, fmt.Errorf("handleRPCServerContext called with info.FullMethodName bad format: %v", info.FullMethodName)
	}

	d := &rpcData{
		startTime:   startTime,
		serviceName: names[1],
		methodName:  names[2],
	}

	if peer.AuthInfo != nil {
		d.authProtocol = peer.AuthInfo.AuthType()
	}

	if deadline, ok := ctx.Deadline(); ok {
		d.deadline = deadline
	}

	// //TODO(acetechnologist): Creating copy of context with local tracing span
	// parentSpan, err := createParentSpan(md)
	// if err != nil {
	// 	return nil, err
	// }
	// span := &trace.Span{}
	// span.Start(parentSpan, trace.ServerType, methodName)
	// ctx = trace.NewContext(ctx, span)

	// Creating copy of context with stats tags
	ctx, err := createStatsContext(ctx, md, d.methodName)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, grpcInstRPCKey, d), nil
}

func handleRPCServerInHeader(ctx context.Context, s *stats.InHeader) error {
	// Increment count of active RPCs on the connection.
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleRPCServerInHeader failed to extract *serverConnStatus")
	}
	atomic.AddInt32(&scs.activeRequests, 1)

	// Set d.localAddr and d.remoteAddr
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCServerInHeader failed to extract *rpcData")
	}
	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr

	// TODO(acetechnologist):
	// If CPU profiler is enabled notify the stats package profiler of the
	// start of a new RPC. This cannot be invoked on handleRPCServerContext
	// because a single routine calls handleRPCServerContext for all RPCs.
	// if stats.ServerRPCStart != nil {
	// 	stats.ServerRPCStart(ctx)
	// }

	// reportStreamzServerDataStart(d)
	// RequestzStart(ctx, d)
	// RequestInfoUpdate(d)
	return nil
}

func handleRPCServerInPayload(ctx context.Context, s *stats.InPayload) error {
	// Record payload length received on this connection.
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleRPCServerInPayload failed to extract *serverConnStatus")
	}
	atomic.AddInt64(&scs.requests.count, 1)
	atomic.AddInt64(&scs.requests.numBytes, int64(s.Length))

	// Record payload length received on this rpc.
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCServerInPayload failed to extract *rpcData")
	}
	atomic.AddInt32(&d.reqLen, int32(s.Length))
	atomic.AddInt32(&d.wireReqLen, int32(s.WireLength))

	// TODO(acetechnologist):
	// argumentType, ok := s.Payload.(proto.Message)
	// if !ok {
	// 	return fmt.Errorf("handleRPCServerInPayload failed to extract argumentType. s.Payload is of type %T want type proto.Message", s.Payload)
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

func handleRPCServerOutPayload(ctx context.Context, s *stats.OutPayload) error {
	// Record payload length sent on this rpc.
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCServerOutPayload failed to extract *rpcData")
	}
	atomic.AddInt32(&d.respLen, int32(s.Length))
	atomic.AddInt32(&d.wireRespLen, int32(s.WireLength))

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

func generateRPCServerTrailer(ctx context.Context) (metadata.MD, error) {
	// Record payload length sent on this rpc.
	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return nil, errors.New("generateRPCServerTrailer failed to extract *rpcData")
	}
	d.serverElapsedTime = time.Since(d.startTime)

	// TODO(acetchnologist): generate proto statspb.RpcServerStats and create metadata.MD
	// elapsed := &statspb.RpcServerStats{
	// 	ServerElapsedTime: proto.Float64(float64(d.serverElapsedTime.Seconds())),
	// }
	// b, err := proto.Marshal(elapsed)
	// if err != nil {
	// 	log.Errorf("generateRPCServerTrailer cannot marshal %v to proto. %v", elapsed, err)
	// }
	// return metadata.Pairs(statsKey, string(b)), nil
	return nil, nil
}

func handleRPCServerEnd(ctx context.Context, s *stats.End) error {
	// Decrement count of active RPCs on the connection.
	scs, ok := ctx.Value(grpcInstConnKey).(*serverConnStatus)
	if !ok {
		return errors.New("handleRPCServerEnd failed to extract *serverConnStatus")
	}
	atomic.AddInt32(&scs.activeRequests, -1)

	d, ok := ctx.Value(grpcInstRPCKey).(*rpcData)
	if !ok {
		return errors.New("handleRPCServerEnd failed to extract *rpcData")
	}
	d.err = s.Error

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
	// reportStreamzServerDataEnd(d)
	// DapperRequestPayload(ctx, d)
	// DapperResponsePayload(ctx, d, status)
	// RpczServerFinish(d, status)
	// RequestzFinish(d, status)
	// For lb load reporting.
	// loadRecordEnd(ctx, d, status)
	return nil
}

// createStatsContext creates a census context from the gRPC context and tags
// received in metadata.
func createStatsContext(ctx context.Context, md metadata.MD, methodName string) (context.Context, error) {
	var cc pb.StatsContext

	if statsBin, ok := md[statsKey]; ok {
		if len(statsBin) != 1 {
			return nil, errors.New("createStatsContext failed to extract statsBin. Have a length different than 1 in the metadata received")
		}

		if err := proto.Unmarshal([]byte(statsBin[0]), &cc); err != nil {
			return nil, fmt.Errorf("createStatsContext failed to unmarshal statsBin[0]. Format is incorrect: %v. %v", statsBin[0], err)
		}
	}

	tagsSet, err := tagging.DecodeFromFullSignatureToTagsSet(cc.Tags)
	if err != nil {
		return nil, fmt.Errorf("createStatsContext failed to decode. %v", err)
	}

	tagsSet[keyMethodName] = keyMethodName.CreateTag(methodName)

	return tagging.NewContextWithTagsSet(ctx, tagsSet), nil
}
