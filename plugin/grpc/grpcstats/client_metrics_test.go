// Copyright 2017, OpenCensus Authors
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

package grpcstats

import (
	"reflect"
	"runtime"
	"testing"

	"go.opencensus.io/stats"
)

func TestViewsAggregationsConform(t *testing.T) {
	// See Issue https://github.com/census-instrumentation/opencensus-go/issues/210.
	// This test ensures that the types of our Views match up with those
	// from the Java reference at
	// https://github.com/census-instrumentation/opencensus-java/blob/2b464864e3dd3f80e8e4c9dc72fccc225444a939/contrib/grpc_metrics/src/main/java/io/opencensus/contrib/grpc/metrics/RpcViewConstants.java#L113-L658
	// Add any other defined views to be type checked during tests to ensure we don't regress.

	assertTypeOf := func(v *stats.View, wantSample stats.Aggregation) {
		aggregation := v.Aggregation()
		gotValue := reflect.ValueOf(aggregation)
		wantValue := reflect.ValueOf(wantSample)
		if gotValue.Type() != wantValue.Type() {
			_, _, line, _ := runtime.Caller(1)
			t.Errorf("Item on line: %d got %T want %T", line, aggregation, wantSample)
		}
	}

	assertTypeOf(RPCClientErrorCountView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRoundTripLatencyView, stats.DistributionAggregation{})
	assertTypeOf(RPCClientRequestBytesView, stats.DistributionAggregation{})
	assertTypeOf(RPCClientResponseBytesView, stats.DistributionAggregation{})
	assertTypeOf(RPCClientRequestCountView, stats.DistributionAggregation{})
	assertTypeOf(RPCClientResponseCountView, stats.DistributionAggregation{})
	assertTypeOf(RPCClientRoundTripLatencyMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRequestBytesMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientResponseBytesMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientErrorCountMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientStartedCountMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientFinishedCountMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRequestCountMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientResponseCountMinuteView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRoundTripLatencyHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRequestBytesHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientResponseBytesHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientErrorCountHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientStartedCountHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientFinishedCountHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientRequestCountHourView, stats.MeanAggregation{})
	assertTypeOf(RPCClientResponseCountHourView, stats.MeanAggregation{})
}
