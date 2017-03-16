// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/golang/glog"
	pb "github.com/google/instrumentation-go/grpc-plugin-app-sample"
	"github.com/google/instrumentation-go/grpc-plugin/export"
	instPb "github.com/grpc/grpc-proto/grpc/instrumentation/v1alpha"
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
	registerCanonical()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		glog.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterGreeterServer(grpcServer, new(server))
	instPb.RegisterMonitoringServer(grpcServer, export.NewServer())

	grpcServer.Serve(lis)
}
