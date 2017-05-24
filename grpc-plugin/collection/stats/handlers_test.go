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

	"google.golang.org/grpc/stats"
)

func TestServerDefaultCollections(t *testing.T) {
	h := ServerHandler{}

	conn1Ctx := context.Background()
	conn1Info := &stats.ConnTagInfo{}

	// connectConn1
	conn1CtxWithInfo := h.TagConn(conn1Ctx, conn1Info)

	// conn1Rpc1
	rpc1Ctx := context.Background()
	rpc1Info := &stats.RPCTagInfo{}
	rpc1Ctx = h.TagRPC(rpc1Ctx, rpc1Info)

	rpc1Begin := &stats.Begin{}
	h.HandleRPC(rpc1Ctx, rpc1Begin)

	rpc1InHdr := &stats.InHeader{}
	h.HandleRPC(rpc1Ctx, rpc1InHdr)

	rpc1InPay := &stats.InPayload{}
	h.HandleRPC(rpc1Ctx, rpc1InPay)

	rpc1InTrailer := &stats.InTrailer{}
	h.HandleRPC(rpc1Ctx, rpc1InTrailer)

	rpc1OutCtx := context.Background()
	rpc1OutHdr := &stats.OutHeader{}
	h.HandleRPC(rpc1OutCtx, rpc1OutHdr)

	rpc1OutPay := &stats.OutPayload{}
	h.HandleRPC(rpc1OutCtx, rpc1OutPay)

	md, err := h.GenerateServerTrailer(rpc1OutCtx)
	if err != nil {
	}
	if md != nil {
	}

	rpc1OutTrailer := &stats.OutTrailer{}
	h.HandleRPC(rpc1OutCtx, rpc1OutTrailer)
	if err != nil {

	}

	rpc1End := &stats.End{}
	h.HandleRPC(rpc1OutCtx, rpc1End)

	conn1End := &stats.ConnEnd{}
	h.HandleConn(conn1CtxWithInfo, conn1End)

	// connectConn2
	// conn2Rpc1
	// conn2Rpc2
	// conn2Rpc3
}

func TestClientDefaultCollections(t *testing.T) {
	c := ClientHandler{}
	conn1Ctx := context.Background()
	conn1Info := &stats.ConnTagInfo{}

	// connectConn1
	conn1CtxWithConnInfo := c.TagConn(conn1Ctx, conn1Info)

	// conn1Rpc1
	// conn1Rpc2
	// conn1Rpc3
	c1End := &stats.ConnEnd{}
	c.HandleConn(conn1CtxWithConnInfo, c1End)

	// connectConn2
	// conn2Rpc1
	// conn2Rpc2
	// conn2Rpc3
}
