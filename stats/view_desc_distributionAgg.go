package stats

import (
	"bytes"
	"fmt"
	"time"

	"github.com/google/instrumentation-go/stats/tagging"
)

// DistributionAggViewDesc holds the parameters describing an aggregation
// distribution..
type DistributionAggViewDesc struct {
	*ViewDescCommon

	// An aggregation distribution may contain a histogram of the values in the
	// population. The bucket boundaries for that histogram are described
	// by Bounds. This defines len(Bounds)+1 buckets.
	//
	// if len(Bounds) >= 2 then the boundaries for bucket index i are:
	// [-infinity, bounds[i]) for i = 0
	// [bounds[i-1], bounds[i]) for 0 < i < len(Bounds)
	// [bounds[i-1], +infinity) for i = len(Bounds)
	//
	// if len(Bounds) == 0 then there is no histogram associated with the
	// distribution. There will be a single bucket with boundaries
	// (-infinity, +infinity).
	//
	// if len(Bounds) == 1 then there is no finite buckets, and that single
	// element is the common boundary of the overflow and underflow buckets.
	Bounds []float64
}

func (dd *DistributionAggViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newDistributionAggregator(dd.Bounds), nil
}

func (dd *DistributionAggViewDesc) retrieveView(now time.Time) (*View, error) {
	dav, err := dd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		AggregationViewDesc: dd,
		ViewAgg:             dav,
	}, nil
}

func (dd *DistributionAggViewDesc) viewDesc() *ViewDescCommon {
	return dd.ViewDescCommon
}

func (dd *DistributionAggViewDesc) isValid() error {
	for i := 1; i < len(dd.Bounds); i++ {
		if dd.Bounds[i-1] >= dd.Bounds[i] {
			return fmt.Errorf("%v error. bounds are not increasing", dd)
		}
	}
	return nil
}

func (dd *DistributionAggViewDesc) retrieveAggreationView(t time.Time) (*DistributionAggView, error) {
	var aggs []*DistributionAgg

	for sig, a := range dd.signatures {
		tags, err := tagging.TagsFromValuesSignature([]byte(sig), dd.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature %v", sig)
		}
		aggregator, ok := a.(*distributionAggregator)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.distributionAggregator", a)
		}
		da := &DistributionAgg{
			DistributionStats: aggregator.retrieveCollected(),
			Tags:              tags,
		}
		aggs = append(aggs, da)
	}

	return &DistributionAggView{
		Descriptor:   dd,
		Aggregations: aggs,
		Start:        dd.start,
		End:          t,
	}, nil
}

// DistributionAggView is the set of collected DistributionAgg associated with
// ViewDesc.
type DistributionAggView struct {
	Descriptor   *DistributionAggViewDesc
	Aggregations []*DistributionAgg
	Start, End   time.Time // start is time when ViewDesc was registered.
}

// An DistributionAgg is a statistical summary of measures associated with a
// unique tag set.
type DistributionAgg struct {
	*DistributionStats
	Tags []tagging.Tag
}

func (dd *DistributionAggViewDesc) String() string {
	if dd == nil {
		return "nil"
	}
	vd := dd.ViewDescCommon
	var buf bytes.Buffer
	buf.WriteString("  viewDesc{\n")
	fmt.Fprintf(&buf, "    Name: %v,\n", vd.Name)
	fmt.Fprintf(&buf, "    Description: %v,\n", vd.Description)
	fmt.Fprintf(&buf, "    MeasureDescName: %v,\n", vd.MeasureDescName)
	fmt.Fprintf(&buf, "    TagKeys: %v,\n", vd.TagKeys)
	fmt.Fprintf(&buf, "    Bound: %v,\n", dd.Bounds)
	buf.WriteString("    },\n")
	buf.WriteString("  }")
	return buf.String()
}

func (dv *DistributionAggView) String() string {
	if dv == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  viewAgg{\n")
	fmt.Fprintf(&buf, "    Start: %v,\n", dv.Start)
	fmt.Fprintf(&buf, "    End: %v,\n", dv.End)
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", dv.Aggregations)
	buf.WriteString("  }")
	return buf.String()
}

func (da *DistributionAgg) String() string {
	if da == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", da.DistributionStats)
	fmt.Fprintf(&buf, "    Tags: %v,\n", da.Tags)
	buf.WriteString("  }")
	return buf.String()
}
