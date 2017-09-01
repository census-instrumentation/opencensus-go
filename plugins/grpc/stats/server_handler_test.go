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
	"testing"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

func TestServerDefaultCollections(t *testing.T) {
	type rpc struct {
		encodedTagSet string
		tagInfo       *stats.RPCTagInfo
		inPayloads    []*stats.InPayload
		outPayloads   []*stats.OutPayload
		end           *stats.End
	}

	type testCase struct {
		rpcs []*rpc
	}
	tcs := []testCase{
		{
			[]*rpc{
				{
					"",
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
		},
	}

	for _, tc := range tcs {
		h := ServerHandler{}
		for _, rpc := range tc.rpcs {
			md := metadata.Pairs(tagsKey, string(rpc.encodedTagSet))
			ctx := metadata.NewOutgoingContext(context.Background(), md)

			ctx = h.TagRPC(ctx, rpc.tagInfo)

			for _, in := range rpc.inPayloads {
				h.HandleRPC(ctx, in)
			}

			for _, out := range rpc.outPayloads {
				h.HandleRPC(ctx, out)
			}

			h.HandleRPC(ctx, rpc.end)
		}
	}
}

func TestClientDefaultCollections(t *testing.T) {
	c := ClientHandler{}

	rpc1Info := &stats.RPCTagInfo{}
	ctx := c.TagRPC(context.Background(), rpc1Info)

	rpc1OuPay := &stats.OutPayload{}
	c.HandleRPC(ctx, rpc1OuPay)

	rpc1InPay := &stats.InPayload{}
	c.HandleRPC(ctx, rpc1InPay)

	rpc1End := &stats.End{}
	c.HandleRPC(ctx, rpc1End)
}
