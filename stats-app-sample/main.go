package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	cstats "github.com/google/instrumentation-go/grpc-plugin/collection/stats"
	"github.com/google/instrumentation-go/stats"
	spb "github.com/grpc/grpc-proto/grpc/instrumentation/v1alpha"
)

var c = make(chan *stats.View, 256)

func main() {
	//registerViews()
	fmt.Println("1")
	cstats.RegisterServerDefaults()
	//registerRpcCanonical()

	var measurements []stats.Measurement
	measurements = append(measurements, cstats.RPCserverServerElapsedTime.CreateMeasurement(float64(200000)/float64(time.Millisecond)))
	stats.RecordMeasurements(context.Background(), measurements...)

	// ---------------------------------------RETRIEVE USAGE---------------------------
	req := &spb.StatsRequest{
		MeasurementNames: []string{"/rpc/server/server_elapsed_time"},
		ViewNames:        []string{},
	}

	views := stats.RetrieveViews(req.ViewNames, req.MeasurementNames)

	buildStatsResponse(views)
	return

	done := make(chan bool)
	go func(c chan *stats.View) {
		i := 0
		for {
			i++
			v := <-c
			fmt.Printf("%v -->\n%v\n", i, v)
		}
	}(c)
	<-done
}

func buildStatsResponse(vws []*stats.View) {
	resp := &spb.StatsResponse{}

	fmt.Printf("%v\n", len(vws))
	for _, vw := range vws {
		fmt.Printf("%v", vw)
	}

	_ = resp

}
