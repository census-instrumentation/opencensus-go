package stats

import (
	"bytes"
	"fmt"
	"time"
)

// AggregationViewDesc is the interface that all aggregations are expected to
// implement.
type AggregationViewDesc interface {
	// creates an aggregator instance for a unique tags signature.
	createAggregator(t time.Time) (aggregator, error)
	// retrieves the collected *View holding collected data by all the
	// aggregator instances..
	retrieveView(now time.Time) (*View, error)
	// returns the *ViewDesc associated with this AggregationViewDesc
	viewDesc() *ViewDescCommon
	// validates the input recieved as requested by the client code.
	isValid() error
}

// aggregator is the interface that the aggregators created by an aggregation
// are expected to implement.
type aggregator interface {
	addSample(v Measurement, t time.Time)
}

// ViewDescCommon is a helper data structure that holds common fields to all
// ViewAggregationDesc. It should never be used standalone but always as part
// of a ViewAggregationDesc.
type ViewDescCommon struct {
	// Name of ViewDesc. Must be unique.
	// TODO(iamm2): provide examples for Name.
	Name string
	// TODO(iamm2): provide an example for description.
	Description string

	// MeasureDescName is the name of a Measure. Examples are cpu:tickCount,
	// diskio:time...
	MeasureDescName string

	// Keys to perform the aggregation on.
	TagKeys []string

	// start is time when ViewDesc was registered.
	start time.Time

	// vChans are the channels through which the collected views for this ViewDesc
	// are sent to the consumers of this view.
	vChans map[chan *View]struct{}

	// signatures holds the aggregations for each unique tag signature (values
	// for all keys) to its *stats.Aggregator.
	signatures map[string]aggregator
}

// A View is a set of Aggregations about usage of the single resource
// associated with the given view during a particular time interval. Each
// Aggregation is specific to a unique set of tags. The Census infrastructure
// reports a stream of View events to the application for further processing
// such as further aggregations, logging and export to other services.
type View struct {
	AggregationViewDesc AggregationViewDesc
	// ViewAgg is expected to be a *DistributionAggView or a
	// *IntervalAggView
	ViewAgg interface{}
}

func (vw *View) String() string {
	if vw == nil {
		return "nil"
	}
	var buf bytes.Buffer
	buf.WriteString("View{\n")
	fmt.Fprintf(&buf, "%v,\n", vw.AggregationViewDesc)
	fmt.Fprintf(&buf, "%v,\n", vw.ViewAgg)
	buf.WriteString("}")
	return buf.String()
}
