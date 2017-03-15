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

package collection

import (
	"context"
	"testing"

	"google.golang.org/grpc/stats"
)

func TestServerDefaultCollections(t *testing.T) {
	conn1Ctx := context.Background()
	conn1Info := &stats.ConnTagInfo{}

	// connectConn1
	conn1CtxWithInfo, err := HandleConnServerContext(conn1Ctx, conn1Info)
	if err != nil {

	}

	// conn1Rpc1
	rpc1Ctx := context.Background()
	rpc1Info := &stats.RPCTagInfo{}
	_, err = HandleRPCServerContext(rpc1Ctx, rpc1Info)
	if err != nil {

	}

	rpc1Begin := &stats.Begin{}
	err = HandleBegin(rpc1Ctx, rpc1Begin)
	if err != nil {

	}

	rpc1InHdr := &stats.InHeader{}
	err = HandleInHeader(rpc1Ctx, rpc1InHdr)
	if err != nil {

	}

	rpc1InPay := &stats.InPayload{}
	err = HandleInPayload(rpc1Ctx, rpc1InPay)
	if err != nil {

	}

	rpc1InTrailer := &stats.InTrailer{}
	err = HandleInTrailer(rpc1Ctx, rpc1InTrailer)
	if err != nil {

	}

	rpc1OutCtx := context.Background()
	rpc1OutHdr := &stats.OutHeader{}
	err = HandleOutHeader(rpc1OutCtx, rpc1OutHdr)
	if err != nil {

	}

	rpc1OutPay := &stats.OutPayload{}
	err = HandleOutPayload(rpc1OutCtx, rpc1OutPay)
	if err != nil {

	}

	md, err := GenerateServerTrailer(rpc1OutCtx)
	if err != nil {

	}
	if md != nil {

	}

	rpc1OutTrailer := &stats.OutTrailer{}
	err = HandleOutTrailer(rpc1OutCtx, rpc1OutTrailer)
	if err != nil {

	}

	rpc1End := &stats.End{}
	err = HandleEnd(rpc1OutCtx, rpc1End)
	if err != nil {

	}

	conn1End := &stats.ConnEnd{}
	err = HandleConnEnd(conn1CtxWithInfo, conn1End)
	if err != nil {

	}

	// connectConn2
	// conn2Rpc1
	// conn2Rpc2
	// conn2Rpc3
}

func TestClientDefaultCollections(t *testing.T) {
	conn1Ctx := context.Background()
	conn1Info := &stats.ConnTagInfo{}

	// connectConn1
	conn1CtxWithConnInfo, err := HandleConnClientContext(conn1Ctx, conn1Info)
	if err != nil {

	}
	// conn1Rpc1
	// conn1Rpc2
	// conn1Rpc3
	c1End := &stats.ConnEnd{}
	err = HandleConnEnd(conn1CtxWithConnInfo, c1End)
	if err != nil {

	}
	// connectConn2
	// conn2Rpc1
	// conn2Rpc2
	// conn2Rpc3
}
