package stats

import (
	"bytes"
	"fmt"
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
)

// IntervalAggViewDesc holds the parameters describing an interval aggregation.
type IntervalAggViewDesc struct {
	*ViewDescCommon

	// Number of internal sub-intervals to use when collecting stats for each
	// interval. The max error in interval measurements will be approximately
	// 1/SubIntervals (although in practice, this will only be approached in
	// the presence of very large and bursty workload changes), and underlying
	// memory usage will be roughly proportional to the value of this
	// field. Must be in the range [2, 20]. A value of 5 will be used if this
	// is unspecified.
	SubIntervals int

	// The size of each interval, as a time duration. Must have at least one
	// element.
	Intervals []time.Duration
}

func (id *IntervalAggViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newIntervalsAggregator(t, id.Intervals, id.SubIntervals), nil
}

func (id *IntervalAggViewDesc) retrieveView(now time.Time) (*View, error) {
	iav, err := id.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		AggregationViewDesc: id,
		ViewAgg:             iav,
	}, nil
}

func (id *IntervalAggViewDesc) viewDesc() *ViewDescCommon {
	return id.ViewDescCommon
}

func (id *IntervalAggViewDesc) isValid() error {
	if id.SubIntervals < 2 || id.SubIntervals < 20 {
		return fmt.Errorf("%v error. subIntervals is not in [2,20]", id)
	}
	return nil
}

func (id *IntervalAggViewDesc) retrieveAggreationView(now time.Time) (*IntervalAggView, error) {
	var aggs []*IntervalAgg

	for sig, a := range id.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), id.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature %v", sig)
		}
		aggregator, ok := a.(*intervalsAggregator)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.intervalsAggregator", a)
		}
		ia := &IntervalAgg{
			IntervalStats: aggregator.retrieveCollected(now),
			Tags:          tags,
		}
		aggs = append(aggs, ia)
	}

	return &IntervalAggView{
		Descriptor:   id,
		Aggregations: aggs,
	}, nil
}

// IntervalAggView is the set of collected IntervalAgg associated with
// ViewDesc.
type IntervalAggView struct {
	Descriptor   *IntervalAggViewDesc
	Aggregations []*IntervalAgg
}

// IntervalAgg is a statistical summary of measures associated with a unique
// tag set for a specific time interval.
type IntervalAgg struct {
	IntervalStats []*IntervalStats
	Tags          []tagging.Tag
}

func (id *IntervalAggViewDesc) String() string {
	if id == nil {
		return "nil"
	}
	vd := id.ViewDescCommon
	var buf bytes.Buffer
	buf.WriteString("  viewDesc{\n")
	fmt.Fprintf(&buf, "    Name: %v,\n", vd.Name)
	fmt.Fprintf(&buf, "    Description: %v,\n", vd.Description)
	fmt.Fprintf(&buf, "    MeasureDescName: %v,\n", vd.MeasureDescName)
	fmt.Fprintf(&buf, "    TagKeys: %v,\n", vd.TagKeys)
	fmt.Fprintf(&buf, "    Intervals: %v,\n", id.Intervals)
	buf.WriteString("    },\n")
	buf.WriteString("  }")
	return buf.String()
}

func (iv *IntervalAggView) String() string {
	if iv == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  viewAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", iv.Aggregations)
	buf.WriteString("  }")
	return buf.String()
}

func (ia *IntervalAgg) String() string {
	if ia == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  IntervalAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", ia.IntervalStats)
	fmt.Fprintf(&buf, "    Tags: %v,\n", ia.Tags)
	buf.WriteString("  }")
	return buf.String()
}
