package api

import (
	"bytes"
	"fmt"
	"time"
)

// GaugeInt64ViewDesc defines an int64 gauge view.
type GaugeInt64ViewDesc struct {
	*ViewDescCommon
}

func (gd *GaugeInt64ViewDesc) createAggregator(t time.Time) (aggregator, error) {
	return newGaugeAggregatorInt64(), nil
}

func (gd *GaugeInt64ViewDesc) retrieveView(now time.Time) (*View, error) {
	gav, err := gd.retrieveAggreationView(now)
	if err != nil {
		return nil, err
	}
	return &View{
		AggregationViewDesc: gd,
		ViewAgg:             gav,
	}, nil
}

func (gd *GaugeInt64ViewDesc) viewDesc() *ViewDescCommon {
	return gd.ViewDescCommon
}

func (gd *GaugeInt64ViewDesc) isValid() error {
	return nil
}

func (gd *GaugeInt64ViewDesc) retrieveAggreationView(t time.Time) (*GaugeInt64AggView, error) {
	var aggs []*GaugeInt64Agg

	for sig, a := range gd.signatures {
		tags, err := tagsFromSignature([]byte(sig), gd.TagKeys)
		if err != nil {
			return nil, fmt.Errorf("malformed signature %v", sig)
		}
		aggregator, ok := a.(*gaugeAggregatorInt64)
		if !ok {
			return nil, fmt.Errorf("unexpected aggregator type. got %T, want stats.gaugeAggregatorInt64", a)
		}
		ga := &GaugeInt64Agg{
			GaugeInt64Stats: aggregator.retrieveCollected(),
			Tags:            tags,
		}
		aggs = append(aggs, ga)
	}

	return &GaugeInt64AggView{
		Descriptor:   gd,
		Aggregations: aggs,
	}, nil
}

// GaugeInt64AggView is the set of collected GaugeInt64Agg associated with
// ViewDesc.
type GaugeInt64AggView struct {
	Descriptor   *GaugeInt64ViewDesc
	Aggregations []*GaugeInt64Agg
}

// A GaugeInt64Agg is a statistical summary of measures associated with a
// unique tag set.
type GaugeInt64Agg struct {
	*GaugeInt64Stats
	Tags []Tag
}

// DistributionStats records a distribution of float64 sample values.
// It is the result of a DistributionAgg aggregation.
type GaugeInt64Stats struct {
	Value     int64
	TimeStamp time.Time
}

func (gd *GaugeInt64ViewDesc) String() string {
	if gd == nil {
		return "nil"
	}
	vd := gd.ViewDescCommon
	var buf bytes.Buffer
	buf.WriteString("  viewDesc{\n")
	fmt.Fprintf(&buf, "    Name: %v,\n", vd.Name)
	fmt.Fprintf(&buf, "    Description: %v,\n", vd.Description)
	fmt.Fprintf(&buf, "    MeasureDescName: %v,\n", vd.MeasureDescName)
	fmt.Fprintf(&buf, "    TagKeys: %v,\n", vd.TagKeys)
	buf.WriteString("    },\n")
	buf.WriteString("  }")
	return buf.String()
}

func (gv *GaugeInt64AggView) String() string {
	if gv == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  viewAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", gv.Aggregations)
	buf.WriteString("  }")
	return buf.String()
}

func (ga *GaugeInt64Agg) String() string {
	if ga == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionAgg{\n")
	fmt.Fprintf(&buf, "    Aggregations: %v,\n", ga.GaugeInt64Stats)
	fmt.Fprintf(&buf, "    Tags: %v,\n", ga.Tags)
	buf.WriteString("  }")
	return buf.String()
}

func (gs *GaugeInt64Stats) String() string {
	if gs == nil {
		return "nil"
	}

	var buf bytes.Buffer
	buf.WriteString("  DistributionStats{\n")
	fmt.Fprintf(&buf, "    Value: %v,\n", gs.Value)
	fmt.Fprintf(&buf, "    TimeStamp: %v,\n", gs.TimeStamp)
	buf.WriteString("  }")
	return buf.String()
}
