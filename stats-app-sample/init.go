package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tags"
	"golang.org/x/net/context"
)

// Create/register view description of type gauge float
var (
	RPCserverServerElapsedTime     stats.MeasureDescFloat64
	RPCserverServerElapsedTimeDist *stats.DistributionViewDesc
)

func registerViews() {
	// ---------------------------------------MEASURES---------------------------
	// Create/register measure descriptions of type string
	muString := &stats.MeasurementUnit{}
	mDescString := stats.NewMeasureDescString("SystemLoad", "At any time its value can be: HIGH, MEDIUM, LOW", muString)
	if err := stats.RegisterMeasureDesc(mDescString); err != nil {
		fmt.Printf("Error measure registration mDescString: %v\n", err)
	}

	// Create/register measure descriptions of type bool
	muBool := &stats.MeasurementUnit{}
	mDescBool := stats.NewMeasureDescBool("SystemHot", "At any time its value can be: TRUE, FALSE", muBool)
	if err := stats.RegisterMeasureDesc(mDescBool); err != nil {
		fmt.Printf("Error measure registration mDescmuBool: %v\n", err)
	}

	// Create/register measure descriptions of type float64
	muFloat64 := &stats.MeasurementUnit{
		Power10: 6,
		Numerators: []stats.BasicUnit{
			stats.BytesUnit,
		},
	}
	mDescFloat64 := stats.NewMeasureDescFloat64("DiskRead", "Read MBs", muFloat64)
	if err := stats.RegisterMeasureDesc(mDescFloat64); err != nil {
		fmt.Printf("Error measure registration mDescFloat64: %v\n", err)
	}

	// Create/register measure descriptions of type Int64
	muInt64 := &stats.MeasurementUnit{
		Power10: 6,
		Numerators: []stats.BasicUnit{
			stats.BytesUnit,
		},
	}
	mDescInt64 := stats.NewMeasureDescInt64("ConnectedUsers", "Connected users. It is for a snapshot", muInt64)
	if err := stats.RegisterMeasureDesc(mDescInt64); err != nil {
		fmt.Printf("Error measure registration mDescInt64: %v\n", err)
	}

	// ---------------------------------------VIEWS---------------------------
	// Creates keys/views
	mgr := tagging.DefaultKeyManager()
	k1, err := mgr.CreateKeyInt64("key1")
	if err != nil {
		panic(fmt.Sprintf("Key k1 not created %v", err))
	}
	k2, err := mgr.CreateKeyInt64("key2")
	if err != nil {
		panic(fmt.Sprintf("Key k2 not created %v", err))
	}
	k3, err := mgr.CreateKeyStringUTF8("key3")
	if err != nil {
		panic(fmt.Sprintf("Key k3 not created %v", err))
	}

	// Create/register view description of type gauge float
	c := make(chan *stats.View, 256)

	vwGaugeS := &stats.GaugeStringViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View1",
			Description:     "",
			MeasureDescName: "SystemLoad",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
	}
	if err := stats.RegisterViewDesc(vwGaugeS); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwGaugeS.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}

	vwGaugeB := &stats.GaugeBoolViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View2",
			Description:     "",
			MeasureDescName: "SystemHot",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
	}
	if err := stats.RegisterViewDesc(vwGaugeB); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwGaugeB.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}

	vwGaugeF := &stats.GaugeFloat64ViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View3",
			Description:     "",
			MeasureDescName: "DiskRead",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
	}
	if err := stats.RegisterViewDesc(vwGaugeF); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwGaugeF.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}

	vwGaugeI := &stats.GaugeInt64ViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View4",
			Description:     "",
			MeasureDescName: "ConnectedUsers",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
	}
	if err := stats.RegisterViewDesc(vwGaugeI); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwGaugeI.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}

	vwDist := &stats.DistributionViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View5",
			Description:     "",
			MeasureDescName: "DiskRead",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
		Bounds: []float64{0, 10, 100},
	}
	if err := stats.RegisterViewDesc(vwDist); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwDist.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}

	vwInterval := &stats.IntervalViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "View6",
			Description:     "",
			MeasureDescName: "DiskRead",
			TagKeys:         []tagging.Key{k1, k2, k3},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Second * 5, time.Second * 30, time.Second * 60},
	}
	if err := stats.RegisterViewDesc(vwInterval); err != nil {
		fmt.Printf("Error view registration: %v\n", err)
	}
	err = stats.Subscribe(&stats.SingleSubscription{
		C:        c,
		ViewName: vwInterval.ViewDescCommon().Name,
	})
	if err != nil {
		fmt.Printf("Error view subscription: %v\n", err)
	}
	// ---------------------------------------CREATES TAGS---------------------------
	// Set tags values in mutations
	mut1 := k1.CreateMutation(1, tagging.BehaviorAdd)
	mut2 := k2.CreateMutation(1, tagging.BehaviorAdd)
	mut3 := k3.CreateMutation("s1", tagging.BehaviorAdd)
	// ---------------------------------------RECORD USAGE---------------------------
	ctx := tagging.ContextWithDerivedTagSet(context.Background(), mut1, mut2, mut3)

	//...
	// DoStuff()
	//...
	// Creates few measurements
	m1 := mDescString.CreateMeasurement("HIGH")
	m2 := mDescBool.CreateMeasurement(true)
	m3 := mDescFloat64.CreateMeasurement(209)
	m4 := mDescFloat64.CreateMeasurement(9)
	m5 := mDescInt64.CreateMeasurement(20)
	m6 := mDescInt64.CreateMeasurement(40)
	stats.RecordMeasurements(ctx, m1, m2, m3, m4, m5, m6)

	// ---------------------------------------CREATES TAGS---------------------------
	// Set tags values in mutations
	mut1 = k1.CreateMutation(2, tagging.BehaviorAdd)
	mut2 = k2.CreateMutation(2, tagging.BehaviorAdd)
	mut3 = k3.CreateMutation("s2", tagging.BehaviorAdd)
	// ---------------------------------------RECORD USAGE---------------------------
	ctx = tagging.ContextWithDerivedTagSet(context.Background(), mut1, mut2, mut3)

	//...
	// DoStuff()
	//...
	// Create context with mutations
	// Creates few measurements
	m1 = mDescString.CreateMeasurement("LOW")
	m2 = mDescBool.CreateMeasurement(false)
	m3 = mDescFloat64.CreateMeasurement(218)
	m4 = mDescFloat64.CreateMeasurement(18)
	m5 = mDescInt64.CreateMeasurement(5)
	m6 = mDescInt64.CreateMeasurement(10)
	stats.RecordMeasurements(ctx, m1, m2, m3, m4, m5, m6)
}

func registerRpcCanonical() {
	var measures []stats.MeasureDesc
	var views []stats.ViewDesc

	rpcMillisBucketBoundaries := []float64{0, 1, 2, 3, 4, 5, 6, 8, 10, 13, 16, 20, 25, 30, 40, 50, 65, 80, 100, 130, 160, 200, 250, 300, 400, 500, 650, 800, 1000, 2000, 5000, 10000, 20000, 50000, 100000}

	keyMethod, err := tagging.DefaultKeyManager().CreateKeyStringUTF8("grpc.method")
	if err != nil {
		log.Fatalf("init() failed to create/retrieve keyStringUTF8. %v", err)
	}

	milliseconds := &stats.MeasurementUnit{
		Power10:    -3,
		Numerators: []stats.BasicUnit{stats.SecsUnit},
	}

	RPCserverServerElapsedTime = stats.NewMeasureDescFloat64("/rpc/server/server_elapsed_time", "Server elapsed time in msecs", milliseconds)
	measures = append(measures, RPCserverServerElapsedTime)
	for _, m := range measures {
		if err := stats.RegisterMeasureDesc(m); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", m, err)
		}
	}

	RPCserverServerElapsedTimeDist = &stats.DistributionViewDesc{
		Vdc: &stats.ViewDescCommon{
			Name:            "/rpc/server/server_elapsed_time/distribution_cumulative",
			Description:     "Server elapsed time in msecs",
			MeasureDescName: RPCserverServerElapsedTime.Meta().Name(),
			TagKeys:         []tagging.Key{keyMethod},
		},
		Bounds: rpcMillisBucketBoundaries,
	}
	views = append(views, RPCserverServerElapsedTimeDist)

	for _, v := range views {
		if err := stats.RegisterViewDesc(v); err != nil {
			log.Fatalf("init() failed to register %v.\n %v", v, err)
		}
	}
}
