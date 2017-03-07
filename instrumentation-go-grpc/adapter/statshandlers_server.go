package adapter

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/stats"
)

// This funciton extracts auth info out of LOASAuthInfo type object which is wrapped inside of
// an util.AuthInfo type object. If the object is not the expected type it extracts only the AuthType.
func extractAuthInfo(d *Data, authInfo credentials.AuthInfo) {
	// d.authProtocol = authInfo.AuthType()
	// uauth, ok := authInfo.(*util.AuthInfo)
	// if !ok {
	// 	return
	// }
	// if auth, ok := uauth.AuthInfo.(*l2handshaker.LOASAuthInfo); ok {
	// 	d.caller = auth.MDBUser()
	// 	d.encrypted = auth.SecurityLevel() == commonpb.SecurityLevel_INTEGRITY_AND_PRIVACY
	// 	d.privacyBoosted = false // is not supported in gRPC as of 2016-11-17, so it should always be false.
	// 	d.signed = d.encrypted || auth.SecurityLevel() == commonpb.SecurityLevel_INTEGRITY
	// 	if ui := auth.UnauthInfo(); ui != nil {
	// 		d.borgCell = ui.UnauthBorgCell
	// 	}
	// }
}

func handleServerConnContext(ctx context.Context, info *stats.ConnTagInfo) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("handleServerConnContext called with nil context")
	}

	if info.RemoteAddr == nil || info.LocalAddr == nil {
		return ctx, errors.New("failed to collect connection stats: info.RemoteAddr or info.LocalAddr is nil")
	}

	c := &counter{}
	scs := DefaultManager().NewServerConnStatus(c, info.LocalAddr, info.RemoteAddr)
	ctx = context.WithValue(ctx, grpcInstConnKey, &connData{
		localAddr:        info.LocalAddr,
		remoteAddr:       info.RemoteAddr,
		c:                c,
		serverConnStatus: scs,
	})
	return ctx, nil
}

// TODO(menghanl): handle ConnBegin if necessary.
// func handleConnBeginServer(ctx context.Context, s *stats.ConnBegin) {}

func handleConnEndServer(ctx context.Context, s *stats.ConnEnd) error {
	cd, ok := ctx.Value(grpcInstConnKey).(*connData)
	if !ok {
		return errors.New("*connData cannot be retrieved from context")
	}
	DefaultManager().RemoveServerConnStatus(cd.serverConnStatus)
	return nil
}

func handleServerRPCContext(ctx context.Context, info *stats.RPCTagInfo) (context.Context, error) {
	if ctx == nil {
		return nil, errors.New("called with nil context.Context")
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get metadata")
	}

	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, errors.New("failed to get peer info")
	}

	d := &Data{
		startTime: time.Now(),
	}

	extractAuthInfo(d, peer.AuthInfo)

	parentSpan, err := createParentSpan(md)
	if err != nil {
		return nil, err
	}

	names := strings.Split(info.FullMethodName, "/")
	if len(names) != 3 {
		return nil, fmt.Errorf("info.FullMethodName bad format: %v", info.FullMethodName)
	}
	serviceName := names[1]
	methodName := names[2]

	// //TODO(acetechnologist): Creating copy of context with local tracing span
	// span := &trace.Span{}
	// span.Start(parentSpan, trace.ServerType, methodName)
	// ctx = trace.NewContext(ctx, span)

	// Creating copy of context with census tags
	ctx, err = createCensusContext(ctx, md, d.caller, methodName)
	if err != nil {
		return nil, err
	}

	if deadline, ok := ctx.Deadline(); ok {
		d.deadline = deadline
	}
	// d.family = "Recv." + serviceName // for client, family = "Sent." + call.Service
	d.methodName = methodName
	d.serviceName = serviceName

	ctx = context.WithValue(ctx, grpcInstKey, d)

	return ctx, nil
}

func handleInHeaderServer(ctx context.Context, s *stats.InHeader) error {
	// Increment count of active RPCs on the connection.
	if cd, ok := ctx.Value(grpcInstConnKey).(*connData); ok {
		cd.c.incr()
	} else {
		log.Info("*connData cannot be retrieved from context, rpcz page may not work")
	}

	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	d.localAddr = s.LocalAddr
	d.remoteAddr = s.RemoteAddr

	// Notifiy the census package profiler of the start of a new RPC.
	if census.ServerRPCStart != nil {
		census.ServerRPCStart(ctx)
	}

	reportStreamzServerDataStart(d)
	RequestzStart(ctx, d)
	RequestInfoUpdate(d)
	return nil
}

func handleInPayloadServer(ctx context.Context, s *stats.InPayload) error {
	// Record payload length received on this connection.
	if cd, ok := ctx.Value(grpcInstConnKey).(*connData); ok {
		cd.serverConnStatus.RecordRequest(rpcpb.RPC_Request_REQUEST, uintptr(s.Length))
	} else {
		log.Info("*connData cannot be retrieved from context, rpcz page may not work")
	}

	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	atomic.AddUint32(&d.reqCount, 1)
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

func handleOutPayloadServer(ctx context.Context, s *stats.OutPayload) error {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	atomic.AddUint32(&d.respCount, 1)
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

func generateServerTrailer(ctx context.Context) (metadata.MD, error) {
	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return nil, errors.New("*Data cannot be retrieved from context")
	}

	d.serverElapsedTime = time.Since(d.startTime)

	elapsed := &statspb.RpcServerStats{
		ServerElapsedTime: proto.Float64(float64(d.serverElapsedTime.Seconds())),
	}

	b, err := proto.Marshal(elapsed)
	if err != nil {
		log.Errorf("Cannot marshal %v\n: %v", elapsed, err)
		return nil, fmt.Errorf("cannot marshal %v\n: %v", elapsed, err)
	}

	return metadata.Pairs(statsKey, string(b)), nil
}

func handleEndServer(ctx context.Context, s *stats.End) error {
	// Decrement count of active RPCs on the connection.
	if cd, ok := ctx.Value(grpcInstConnKey).(*connData); ok {
		cd.c.decr()
	} else {
		log.Info("*connData cannot be retrieved from context, rpcz page may not work")
	}

	d, ok := ctx.Value(grpcInstKey).(*Data)
	if !ok {
		return errors.New("*Data cannot be retrieved from context")
	}

	status := util.ErrorToStatus(s.Error)
	d.status = status

	recordCensusInServer(ctx, d, status, s)

	d.span.Finish()
	reportStreamzServerDataEnd(d)
	DapperRequestPayload(ctx, d)
	DapperResponsePayload(ctx, d, status)
	RpczServerFinish(d, status)
	RequestzFinish(d, status)
	// For lb load reporting.
	loadRecordEnd(ctx, d, status)
	return nil
}

func recordCensusInServer(ctx context.Context, d *Data, status *status.Status, s *stats.End) {
	// This replicates what happens in census-rpc.cc:RecordRpcStats.
	tags := []census.Tag{{"PeerJob", "unknown"}, {"TrafficClass", "unknown"}}

	if d.borgCell != "" {
		tags = append(tags, census.Tag{"PeerCell", d.borgCell})
	} else {
		tags = append(tags, census.Tag{"PeerCell", "unknown"})
	}

	if d.caller != "" {
		tags = append(tags, census.Tag{"PeerUser", d.caller})
	}

	switch {
	case d.signed && d.encrypted:
		tags = append(tags, census.Tag{"PeerSecurityLevel", "SSL_PRIVACY_AND_INTEGRITY"})
	case d.signed:
		tags = append(tags, census.Tag{"PeerSecurityLevel", "SSL_INTEGRITY"})
	default:
		tags = append(tags, census.Tag{"PeerSecurityLevel", "SSL_NONE"})
	}

	legacyStatus := rpcstatus.ToLegacy(status)

	var appError, rpcError float64
	if legacyStatus == rpcstatus.LegacyApplicationError {
		appError = 1
	}

	if legacyStatus != rpcstatus.LegacyOK && appError == 0 {
		rpcError = 1
	}
	if rpcError == 1 || appError == 1 {
		tags = append(tags, census.Tag{"OpStatus", legacyStatus.String()})
	}

	ctx, err := census.NewContext(ctx, tags...)
	if err != nil {
		log.Error(err)
		return
	}
	census.RecordUsage(ctx, cpb.ResourceId_RPC_SERVER,
		float64(d.serverElapsedTime)/float64(time.Millisecond),
		float64(d.reqLen),
		float64(d.respLen),
		rpcError,
		appError,
		0, // uncompressedReq. TODO(mmoakil): is this used anywhere?
		0, // uncompressedRes. TODO(mmoakil): is this used anywhere?
		float64(d.serverElapsedTime)/float64(time.Millisecond),
	)
}

func createParentSpan(md metadata.MD) (*trace.Span, error) {
	traceBin, ok := md[traceKey]
	if !ok {
		return nil, nil // it is not an error to return a nil parentSpan.
	}

	if len(traceBin) != 1 {
		return nil, errors.New("traceBin have a length different than 1 in the metadata received")
	}

	// parent Span was started on the client side. Try to decode it to be used
	// as the parent in the server.
	var tc ipb.GoogleTraceContext
	if err := proto.Unmarshal([]byte(traceBin[0]), &tc); err != nil {
		return nil, fmt.Errorf("decoded trace format is incorrect. %v %v", traceBin[0], err)
	}

	return &trace.Span{
		TraceID:             tc.GetTraceIdLo(),
		ID:                  tc.GetSpanId(),
		Mask:                trace.Mask(tc.GetSpanOptions()),
		SamplingProbability: tc.GetInverseSamplingProbability(),
		Type:                trace.ClientType,
	}, nil
}

// createCensusContext creates a census context from the gRPC context and tags
// received in metadata.
func createCensusContext(ctx context.Context, md metadata.MD, caller, methodName string) (context.Context, error) {
	var tags []byte

	if censusBin, ok := md[statsKey]; ok {
		if len(censusBin) != 1 {
			return nil, nil, errors.New("censusBin have a length different than 1 in the metadata received")
		}

		var cc ipb.StatsContext
		if err := proto.Unmarshal([]byte(censusBin[0]), &cc); err != nil {
			return nil, nil, fmt.Errorf("decoded census format is incorrect. %v %v", censusBin[0], err)
		}
		tags = cc.GetTags()
	}

	// creates a handle. If censusc package was not imported h.native will be 0.
	tagsSet, err := census.FromGRPCBytesRequest(caller, methodName, tags)
	if err != nil {
		return nil, nil, fmt.Errorf("constructing census handler from %v failed. %v", tags, err)
	}

	// adds census handle to local context with census.ContextKey():
	ctx = context.WithValue(ctx, census.ContextKey(), tagsSet)
	return ctx, nil
}
