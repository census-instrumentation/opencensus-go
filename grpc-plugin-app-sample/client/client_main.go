package main

import (
	"flag"
	"io"
	"log"

	pb "github.com/google/instrumentation-go/grpc-plugin-app-sample"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "unary"})
	if err != nil {
		log.Fatalf("%v.SayHello(_) = _, %v: ", client, err)
	}
	log.Printf("%v", resp.GetMessage())

	stream, err := client.SayHelloStream(context.Background(), &pb.HelloRequest{Name: "stream packet"})
	if err != nil {
		log.Fatalf("%v.SayHelloStream(_) = _, %v: ", client, err)
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("%v.SayHelloStream(_) = _, %v: ", client, err)
		}
		log.Printf("%v", resp.GetMessage())
	}
}
