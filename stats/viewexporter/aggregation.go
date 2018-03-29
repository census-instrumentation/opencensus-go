// Copyright 2018, OpenCensus Authors
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

package viewexporter

import "go.opencensus.io/stats"

//go:generate stringer -type AggType

// AggType represents the type of aggregation function used on a View.
type AggType int

// All available aggregation types.
const (
	AggTypeNone         AggType = iota // no aggregation; reserved for future use.
	AggTypeCount                       // the count aggregation, see Count.
	AggTypeSum                         // the sum aggregation, see Sum.
	AggTypeDistribution                // the distribution aggregation, see Distribution.
	AggTypeLastValue                   // the last value aggregation, see LastValue.
)

func (t AggType) String() string {
	return aggTypeName[t]
}

var aggTypeName = map[AggType]string{
	AggTypeNone:         "None",
	AggTypeCount:        "Count",
	AggTypeSum:          "Sum",
	AggTypeDistribution: "Distribution",
	AggTypeLastValue:    "LastValue",
}

// Aggregation represents a data aggregation method. Use one of the functions:
// Count, Sum, Mean, or Distribution to construct an Aggregation.
type Aggregation struct {
	Type    AggType   // Type is the Aggregation of this Aggregation.
	Buckets []float64 // Buckets are the bucket endpoints if this Aggregation represents a distribution, see: Distribution().
}

// AggregatedUnit computes the unit of measure for the result of an aggregation of this type.
func (aggType AggType) AggregatedUnit(measuredUnit string) string {
	if aggType == AggTypeCount {
		return stats.UnitNone
	}
	return measuredUnit
}
