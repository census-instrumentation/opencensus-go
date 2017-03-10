package collection

import (
	"log"
	"time"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
)

var (
	// C is the channel where the client code can access the collected views.
	C chan *istats.View

	keyServiceName tagging.KeyStringUTF8
	keyMethodName  tagging.KeyStringUTF8

	kiloBs       *istats.MeasurementUnit
	milliseconds *istats.MeasurementUnit
	count        *istats.MeasurementUnit

	measureRPCReqLen  istats.MeasureDescFloat64
	measureRPCRespLen istats.MeasureDescFloat64
	measureRPCElapsed istats.MeasureDescFloat64
	measureRPCError   istats.MeasureDescInt64

	vwDistributionRPCReqLen  *istats.DistributionViewDesc
	vwDistributionRPCRespLen *istats.DistributionViewDesc
	vwIntervalRPCElapsed     *istats.IntervalViewDesc
	vwGaugeRPCError          *istats.GaugeInt64ViewDesc
)

func init() {
	var err error

	// Initializing keys
	keyMethodName, err = tagging.DefaultKeyManager().CreateKeyStringUTF8("methodName")
	if err != nil {
		log.Fatalf("init() failed to create/retrieve keyStringUTF8. %v", err)
	}
	keyMethodName, err = tagging.DefaultKeyManager().CreateKeyStringUTF8("serviceName")
	if err != nil {
		log.Fatalf("init() failed to create/retrieve keyStringUTF8. %v", err)
	}

	// Creating units
	kiloBs = &istats.MeasurementUnit{
		Power10:    3,
		Numerators: []istats.BasicUnit{istats.BytesUnit},
	}
	milliseconds = &istats.MeasurementUnit{
		Power10:    -3,
		Numerators: []istats.BasicUnit{istats.SecsUnit},
	}
	count = &istats.MeasurementUnit{
		Power10:    1,
		Numerators: []istats.BasicUnit{istats.ScalarUnit},
	}

	// Creating/Registering measures
	measureRPCReqLen = istats.NewMeasureDescFloat64("RPCReqLen", "", kiloBs)
	measureRPCRespLen = istats.NewMeasureDescFloat64("RPCRespLen", "", kiloBs)
	measureRPCElapsed = istats.NewMeasureDescFloat64("RPCElapsed", "", milliseconds)
	measureRPCError = istats.NewMeasureDescInt64("RPCError", "", count)

	if err := istats.RegisterMeasureDesc(measureRPCReqLen); err != nil {
		log.Fatalf("init() failed to register measureRPCReqLen.\n %v", err)
	}
	if err := istats.RegisterMeasureDesc(measureRPCRespLen); err != nil {
		log.Fatalf("init() failed to register measureRPCRespLen.\n %v", err)
	}
	if err := istats.RegisterMeasureDesc(measureRPCElapsed); err != nil {
		log.Fatalf("init() failed to register measureRPCElapsed.\n %v", err)
	}
	if err := istats.RegisterMeasureDesc(measureRPCError); err != nil {
		log.Fatalf("init() failed to register measureRPCError.\n %v", err)
	}

	// Creating/Registering views
	C = make(chan *istats.View, 1024)
	vwDistributionRPCReqLen = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "RPCReqLen",
			Description:     "",
			MeasureDescName: "RPCReqLen",
			TagKeys:         []tagging.Key{keyServiceName, keyMethodName},
		},
		Bounds: []float64{0, 1, 10, 100, 1000, 10000},
	}
	vwDistributionRPCRespLen = &istats.DistributionViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "RPCRespLen",
			Description:     "",
			MeasureDescName: "RPCRespLen",
			TagKeys:         []tagging.Key{keyServiceName, keyMethodName},
		},
		Bounds: []float64{0, 1, 10, 100, 1000, 10000},
	}
	vwIntervalRPCElapsed = &istats.IntervalViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "RPCElapsed",
			Description:     "",
			MeasureDescName: "RPCElapsed",
			TagKeys:         []tagging.Key{keyServiceName, keyMethodName},
		},
		SubIntervals: 5,
		Intervals:    []time.Duration{time.Second * 10, time.Second * 30, time.Minute * 1, time.Minute * 10},
	}
	vwGaugeRPCError = &istats.GaugeInt64ViewDesc{
		Vdc: &istats.ViewDescCommon{
			Name:            "RPCError",
			Description:     "",
			MeasureDescName: "RPCError",
			TagKeys:         []tagging.Key{keyServiceName, keyMethodName},
		},
	}

	if err := istats.RegisterViewDesc(vwDistributionRPCReqLen, C); err != nil {
		log.Fatalf("init() failed to register vwDistributionRPCReqLen.\n %v", err)
	}
	if err := istats.RegisterViewDesc(vwDistributionRPCRespLen, C); err != nil {
		log.Fatalf("init() failed to register vwDistributionRPCRespLen.\n %v", err)
	}
	if err := istats.RegisterViewDesc(vwIntervalRPCElapsed, C); err != nil {
		log.Fatalf("init() failed to register vwIntervalRPCElapsed.\n %v", err)
	}
	if err := istats.RegisterViewDesc(vwGaugeRPCError, C); err != nil {
		log.Fatalf("init() failed to register vwGaugeRPCError.\n %v", err)
	}
}
