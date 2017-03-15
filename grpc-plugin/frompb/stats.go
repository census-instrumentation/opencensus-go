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

package frompb

import (
	"fmt"
	"time"

	istats "github.com/google/instrumentation-go/stats"
	"github.com/google/instrumentation-go/stats/tagging"
	pb "github.com/google/instrumentation-proto/stats"
)

// Desc converts a protocol buffer ViewDescriptor to a pure Go struct
// implementing github.com/google/instrumentation-go/stats.ViewDesc.
func Desc(d *pb.ViewDescriptor) (istats.ViewDesc, error) {
	var vd istats.ViewDesc
	switch t := d.Aggregation.(type) {
	case *pb.ViewDescriptor_DistributionAggregation:
		vd = distributionDesc(t)
	case *pb.ViewDescriptor_IntervalAggregation:
		vd = intervalDesc(t)
	//case *pb.ViewDescriptor_GaugeAggregation:
	//case *pb.ViewDescriptor_CounterAggregation:
	default:
		return nil, fmt.Errorf("%T no supported in frompb.Desc()", t)
	}

	vdc := vd.ViewDescCommon()
	vdc.Description = d.Description
	vdc.MeasureDescName = d.MeasurementDescriptorName
	vdc.Name = d.Name
	vdc.TagKeys = []tagging.Key{}
	for _, k := range d.TagKeys {
		sk, err := tagging.DefaultKeyManager().CreateKeyStringUTF8(k)
		if err != nil {
			return nil, fmt.Errorf("ToDesc failed. %v", err)
		}
		vdc.TagKeys = append(vdc.TagKeys, sk)
	}
	return vd, nil
}

func distributionDesc(vd *pb.ViewDescriptor_DistributionAggregation) *istats.DistributionViewDesc {
	tmp := &istats.DistributionViewDesc{
		Vdc:    &istats.ViewDescCommon{},
		Bounds: make([]float64, len(vd.DistributionAggregation.BucketBounds)),
	}
	copy(tmp.Bounds, vd.DistributionAggregation.BucketBounds)
	return tmp
}

func intervalDesc(vd *pb.ViewDescriptor_IntervalAggregation) *istats.IntervalViewDesc {
	tmp := &istats.IntervalViewDesc{
		Vdc:          &istats.ViewDescCommon{},
		SubIntervals: int(vd.IntervalAggregation.NSubIntervals),
	}
	for _, d := range vd.IntervalAggregation.IntervalSizes {
		sd := time.Duration(d.Nanos) + time.Second*time.Duration(d.Seconds)
		tmp.Intervals = append(tmp.Intervals, sd)
	}
	return tmp
}

// func ToGaugeDesc(dv *pb.GaugeDesc) *istats.ViewDesc {
// 	return *istats.GaugeBoolViewDesc
// 	return *istats.GaugeFloat64ViewDesc
// 	return *istats.GaugeInt64ViewDesc
// 	return *istats.GaugeStringViewDesc
// }

//func ToCounterDesc(*pb.CounterDesc) *istats.ViewDesc{
// 	return *istats.CounterInt64ViewDesc
// }
