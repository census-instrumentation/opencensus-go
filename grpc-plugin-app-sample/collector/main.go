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
	"bytes"
	"flag"
	"fmt"
	"io"

	"github.com/golang/glog"
	statsPb "github.com/google/instrumentation-proto/stats"
	pb "github.com/grpc/grpc-proto/grpc/instrumentation/v1alpha"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The instrumentation server address in the format of host:port")
var vname = flag.String("view_name", "", "The view name to extract. If empty (the default) will return all views")

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		glog.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()
	client := pb.NewMonitoringClient(conn)

	resp, err := client.GetStats(context.Background(), &pb.StatsRequest{
		MeasurementNames: []string{"/rpc/server/server_latency"},
		ViewNames:        []string{},
	})
	if err != nil {
		glog.Fatalf("%v.WatchStats(_) = _, %v: ", client, err)
	}
	processResponse(resp)
	return

	stream, err := client.WatchStats(context.Background(), &pb.StatsRequest{
		DontIncludeDescriptorsInFirstResponse: true,
		MeasurementNames:                      []string{},
		ViewNames:                             []string{},
	})
	if err != nil {
		glog.Fatalf("%v.WatchStats(_) = _, %v: ", client, err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			glog.Fatalf("%v.WatchStats(_) = _, %v: ", client, err)
		}
		processResponse(resp)
	}
}

func processResponse(resp *pb.StatsResponse) {
	for _, vr := range resp.GetViewResponses() {
		switch vt := vr.View.View.(type) {
		case *statsPb.View_DistributionView:
			glog.Infof("%v:\n\t%v", vr.View.ViewName, vt.DistributionView.Aggregations)
		case *statsPb.View_IntervalView:
			var b bytes.Buffer
			b.WriteString(fmt.Sprintf("%v:\n", vr.View.ViewName))
			for _, a := range vt.IntervalView.Aggregations {
				b.WriteString(fmt.Sprintf("\t%v\n", a))
			}
			glog.Infof("%v", b.String())
		default:
			glog.Infof("\tcannot print view %T", vr.View.View)
		}
	}
}
