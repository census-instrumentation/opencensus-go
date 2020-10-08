package ocgrpc

import (
	"context"
	"log"
	"net"
	"testing"

	"go.opencensus.io/plugin/ocgrpc/helloworld"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type Recorder struct {
	spans []*trace.SpanData
}

func (r *Recorder) ExportSpan(s *trace.SpanData) {
	r.spans = append(r.spans, s)
}

func (r *Recorder) Flush() []*trace.SpanData {
	spans := r.spans
	r.spans = nil
	return spans
}

func initTracer() func() []*trace.SpanData {
	recorder := &Recorder{}
	trace.RegisterExporter(recorder)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})
	return recorder.Flush
}

var _ helloworld.GreeterServer = server{}

type server struct {
	*helloworld.UnimplementedGreeterServer
}

func (s server) SayHello(ctx context.Context, _ *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: "Hallo"}, nil
}

// createDialer creates a connection to be used as context dialer in GRPC
// communication.
func createDialer(s *grpc.Server) func(context.Context, string) (net.Conn, error) {
	const bufSize = 1024 * 1024

	listener := bufconn.Listen(bufSize)
	conn := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	go func() {
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return conn
}

func TestOutterSpanIsPassedToInterceptor(t *testing.T) {
	flusher := initTracer()

	s := grpc.NewServer()
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, &server{})

	dialer := createDialer(s)
	ctx, outterSpan := trace.StartSpan(context.Background(), "outter_span", trace.WithSpanKind(1))
	ctxSpanID := new(string)
	conn, err := grpc.Dial(
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithStatsHandler(&ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler: trace.AlwaysSample(),
			},
		}),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			*ctxSpanID = trace.FromContext(ctx).SpanContext().SpanID.String()
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)

	_, err = client.SayHello(
		ctx,
		&helloworld.HelloRequest{
			Name: "Pupo",
		},
	)
	if err != nil {
		t.Fatalf("call to Register failed: %v", err)
	}
	outterSpan.End()

	reportedSpans := flusher()

	if want, have := 2, len(reportedSpans); want != have {
		t.Errorf("unexpected number of spans")
	}

	if want, have := "helloworld.Greeter.SayHello", reportedSpans[0].Name; want != have {
		t.Fatalf("unexpected first span reported")
	}

	statsHandlerSpanID := reportedSpans[0].SpanID.String()
	if *ctxSpanID != statsHandlerSpanID {
		t.Errorf("span from the interceptor %q is different from the one created by the stats handler %q", *ctxSpanID, statsHandlerSpanID)
	}
}

func TestClientInterceptorCannotAccessClientHandler(t *testing.T) {
	flusher := initTracer()

	s := grpc.NewServer()
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, &server{})

	dialer := createDialer(s)
	ctx := context.Background()
	ctxSpan := new(trace.Span)
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithStatsHandler(&ClientHandler{
			StartOptions: trace.StartOptions{
				Sampler: trace.AlwaysSample(),
			},
		}),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctxSpan = trace.FromContext(ctx)
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)

	_, err = client.SayHello(
		context.Background(),
		&helloworld.HelloRequest{
			Name: "Pupo",
		},
	)
	if err != nil {
		t.Fatalf("call to Register failed: %v", err)
	}

	reportedSpans := flusher()

	// We check that there is a span being reported
	if want, have := 1, len(reportedSpans); want != have {
		t.Errorf("unexpected number of spans")
	}

	// we check that the span being reported is for the client
	if want, have := "helloworld.Greeter.SayHello", reportedSpans[0].Name; want != have {
		t.Fatalf("unexpected first span reported")
	}

	if ctxSpan == nil {
		t.Fatalf("no span in context passed to interceptor")
	}
}

func TestServerRegisterPersonSuccess(t *testing.T) {
	flusher := initTracer()

	s := grpc.NewServer(
		grpc.StatsHandler(&ServerHandler{}),
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			ctxSpan := trace.FromContext(ctx)
			if ctxSpan == nil {
				t.Fatalf("no span in context passed to interceptor")
			}
			return handler(ctx, req)
		}),
	)
	defer s.Stop()

	helloworld.RegisterGreeterServer(s, &server{})

	dialer := createDialer(s)

	ctx := context.Background()
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := helloworld.NewGreeterClient(conn)

	_, err = client.SayHello(ctx, &helloworld.HelloRequest{
		Name: "Pupo",
	})
	if err != nil {
		t.Fatalf("call to Register failed: %v", err)
	}

	reportedSpans := flusher()

	if want, have := 1, len(reportedSpans); want != have {
		t.Errorf("unexpected number of spans")
	}

	if want, have := "helloworld.Greeter.SayHello", reportedSpans[0].Name; want != have {
		t.Fatalf("unexpected first span reported")
	}
}
