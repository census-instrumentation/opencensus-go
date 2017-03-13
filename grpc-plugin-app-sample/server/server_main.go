package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/google/instrumentation-go/grpc-plugin-app-sample"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type server struct{}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "Hello " + req.GetName(),
	}, nil
}

func (s *server) SayHelloStream(req *pb.HelloRequest, stream pb.Greeter_SayHelloStreamServer) error {
	for i := 0; i < 5; i++ {
		err := stream.Send(&pb.HelloReply{
			Message: fmt.Sprintf("Hello %v %v", req.GetName(), i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(grpcServer, new(server))
	grpcServer.Serve(lis)
}
