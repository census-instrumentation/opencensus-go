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
	"sort"
	"time"

	"github.com/golang/glog"
	pb "github.com/google/instrumentation-go/grpc-plugin-app-sample"
	"github.com/google/instrumentation-go/grpc-plugin/collection"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in the format of host:port")

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())

	{
		// insturmentaiton specific
		collection.RegisterClientDefaults()
		statsHandler := collection.ClientHandler{}
		opts = append(opts, grpc.WithStatsHandler(statsHandler))
	}

	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		glog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)

	start := time.Now()
	var durations []int64
	for i := 0; i < 10; i++ {
		rpcstart := time.Now()
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "unary"})
		if err != nil {
			glog.Fatalf("%v.SayHello(_) = _, %v: ", client, err)
		}
		if i%10 == 0 {
			if glog.V(3) {
				glog.Infof("%v", resp.GetMessage())
			}
		}
		durations = append(durations, int64(time.Now().Sub(rpcstart)))

		// stream, err := client.SayHelloStream(context.Background(), &pb.HelloRequest{Name: "stream packet"})
		// if err != nil {
		// 	glog.Fatalf("%v.SayHelloStream(_) = _, %v: ", client, err)
		// }
		// for {
		// 	resp, err := stream.Recv()
		// 	if err == io.EOF {
		// 		break
		// 	}
		// 	if err != nil {
		// 		glog.Fatalf("%v.SayHelloStream(_) = _, %v: ", client, err)
		// 	}
		// 	if i%1000 == 0 {
		// 		glog.Infof("%v", resp.GetMessage())
		// 	}
		// }
	}
	glog.Infof("Elapsed: %v", time.Now().Sub(start))
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	glog.Infof("Best: %v", time.Duration(durations[0]))
	glog.Infof("Median: %v", time.Duration(durations[len(durations)/2]))
	glog.Infof("Worst: %v", time.Duration(durations[len(durations)-1]))
}
