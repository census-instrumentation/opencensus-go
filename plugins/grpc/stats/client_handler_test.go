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

package stats

import (
	"context"
	"errors"
	"testing"

	istats "github.com/census-instrumentation/opencensus-go/stats"
	"github.com/census-instrumentation/opencensus-go/tags"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

func TestClientDefaultCollections(t *testing.T) {
	k1, _ := tags.CreateKeyString("k1")
	k2, _ := tags.CreateKeyString("k2")

	type tagPair struct {
		k *tags.KeyString
		v string
	}

	type wantData struct {
		v    func() istats.View
		rows []*istats.Row
	}
	type rpc struct {
		tags        []tagPair
		tagInfo     *stats.RPCTagInfo
		inPayloads  []*stats.InPayload
		outPayloads []*stats.OutPayload
		end         *stats.End
	}

	type testCase struct {
		label string
		rpcs  []*rpc
		wants []*wantData
	}
	tcs := []testCase{
		{
			"1",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
					},
					&stats.End{Error: nil},
				},
			},
			[]*wantData{
				{
					func() istats.View { return RPCClientRequestCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 1, 1, 1, 0),
						},
					},
				},
				{
					func() istats.View { return RPCClientResponseCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 1, 1, 1, 0),
						},
					},
				},
				{
					func() istats.View { return RPCClientRequestBytesView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcBytesBucketBoundaries, []int64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 10, 10, 10, 0),
						},
					},
				},
				{
					func() istats.View { return RPCClientResponseBytesView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcBytesBucketBoundaries, []int64{0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 1, 10, 10, 10, 0),
						},
					},
				},
			},
		},
		{
			"2",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
						{Length: 10},
						{Length: 10},
					},
					&stats.End{Error: nil},
				},
				{
					[]tagPair{{k1, "v11"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 10},
						{Length: 10},
					},
					[]*stats.OutPayload{
						{Length: 10},
						{Length: 10},
					},
					&stats.End{Error: errors.New("someError")},
				},
			},
			[]*wantData{
				{
					func() istats.View { return RPCClientErrorCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyOpStatus, []byte("someError")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationCountValue(1),
						},
					},
				},
				{
					func() istats.View { return RPCClientRequestCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 2, 2, 3, 2.5, 0.5),
						},
					},
				},
				{
					func() istats.View { return RPCClientResponseCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 2, 1, 2, 1.5, 0.5),
						},
					},
				},
			},
		},
		{
			"3",
			[]*rpc{
				{
					[]tagPair{{k1, "v1"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 1},
					},
					[]*stats.OutPayload{
						{Length: 1},
						{Length: 1024},
						{Length: 65536},
					},
					&stats.End{Error: nil},
				},
				{
					[]tagPair{{k1, "v1"}, {k2, "v2"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 1024},
					},
					[]*stats.OutPayload{
						{Length: 4096},
						{Length: 16384},
					},
					&stats.End{Error: errors.New("someError1")},
				},
				{
					[]tagPair{{k1, "v11"}, {k2, "v22"}},
					&stats.RPCTagInfo{FullMethodName: "/package.service/method"},
					[]*stats.InPayload{
						{Length: 2048},
						{Length: 16384},
					},
					[]*stats.OutPayload{
						{Length: 2048},
						{Length: 4096},
						{Length: 16384},
					},
					&stats.End{Error: errors.New("someError2")},
				},
			},
			[]*wantData{
				{
					func() istats.View { return RPCClientErrorCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyOpStatus, []byte("someError1")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationCountValue(1),
						},
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyOpStatus, []byte("someError2")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationCountValue(1),
						},
					},
				},
				{
					func() istats.View { return RPCClientRequestCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 3, 2, 3, 2.666666666, 0.333333333*2),
						},
					},
				},
				{
					func() istats.View { return RPCClientResponseCountView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcCountBucketBoundaries, []int64{0, 0, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 3, 1, 2, 1.333333333, 0.333333333*2),
						},
					},
				},
				{
					func() istats.View { return RPCClientRequestBytesView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcBytesBucketBoundaries, []int64{0, 1, 1, 1, 2, 2, 1, 0, 0, 0, 0, 0, 0, 0, 0}, 8, 1, 65536, 13696.125, 481423542.982143*7),
						},
					},
				},
				{
					func() istats.View { return RPCClientResponseBytesView },
					[]*istats.Row{
						{
							[]tags.Tag{
								{keyMethod, []byte("method")},
								{keyService, []byte("package.service")},
							},
							istats.NewTestingAggregationDistributionValue(rpcBytesBucketBoundaries, []int64{0, 1, 1, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 4, 1, 16384, 4864.25, 59678208.25*3),
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		istats.RestartWorker()
		registerDefaultsClient()

		h := ClientHandler{}
		for _, rpc := range tc.rpcs {
			tsb := tags.NewTagSetBuilder(nil)
			for _, t := range rpc.tags {
				tsb.UpsertString(t.k, t.v)
			}
			ts := tsb.Build()
			encoded := tags.EncodeToFullSignature(ts)

			md := metadata.Pairs(tagsKey, string(encoded))
			ctx := metadata.NewIncomingContext(context.Background(), md)

			ctx = h.TagRPC(ctx, rpc.tagInfo)

			for _, out := range rpc.outPayloads {
				h.HandleRPC(ctx, out)
			}

			for _, in := range rpc.inPayloads {
				h.HandleRPC(ctx, in)
			}

			h.HandleRPC(ctx, rpc.end)
		}

		for _, wantData := range tc.wants {
			gotRows, err := istats.RetrieveData(wantData.v())
			if err != nil {
				t.Errorf("Test case '%v'. RetrieveData for %v failed. %v", tc.label, wantData.v().Name(), err)
				continue
			}

			for _, gotRow := range gotRows {
				if !istats.RowsContain(wantData.rows, gotRow) {
					t.Errorf("Test case '%v'. View '%v' got unexpected row '%v'", tc.label, wantData.v().Name(), gotRow)
					break
				}
			}

			for _, wantRow := range wantData.rows {
				if !istats.RowsContain(gotRows, wantRow) {
					t.Errorf("Test case '%v'. View '%v' want row '%v'. Not received", tc.label, wantData.v().Name(), wantRow)
					break
				}
			}
		}
	}
}
