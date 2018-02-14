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

package trace

import (
	"testing"

	"google.golang.org/grpc"
)

// Start: 0. PlainClient-PlainServer
func BenchmarkPlainClientPlainServer1QPS(b *testing.B) {
	benchmarkPlainClientPlainServer(b, 1)
}

func BenchmarkPlainClientPlainServer10QPS(b *testing.B) {
	benchmarkPlainClientPlainServer(b, 10)
}

func BenchmarkPlainClientPlainServer100QPS(b *testing.B) {
	benchmarkPlainClientPlainServer(b, 100)
}

func BenchmarkPlainClientPlainServer1000QPS(b *testing.B) {
	benchmarkPlainClientPlainServer(b, 1000)
}

func benchmarkPlainClientPlainServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(plainServerAddr, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}
// End: 0. PlainClient-PlainServer

// Start: 1. NoTraceNoStatsClient-NoTraceNoStatsServer
func BenchmarkNoTraceNoStatsClientNoTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceNoStatsServer(b, 1)
}

func BenchmarkNoTraceNoStatsClientNoTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceNoStatsServer(b, 10)
}

func BenchmarkNoTraceNoStatsClientNoTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceNoStatsServer(b, 100)
}

func BenchmarkNoTraceNoStatsClientNoTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceNoStatsServer(b, 1000)
}

func benchmarkNoTraceNoStatsClientNoTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceNoStatsAddr, noTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 1. NoTraceNoStatsClient-NoTraceNoStatsServer

// Start: 2. NoTraceNoStatsClient-NoTraceYesStatsServer
func BenchmarkNoTraceNoStatsClientNoTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceYesStatsServer(b, 1)
}

func BenchmarkNoTraceNoStatsClientNoTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceYesStatsServer(b, 10)
}

func BenchmarkNoTraceNoStatsClientNoTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceYesStatsServer(b, 100)
}

func BenchmarkNoTraceNoStatsClientNoTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientNoTraceYesStatsServer(b, 1000)
}

func benchmarkNoTraceNoStatsClientNoTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, noTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 2. NoTraceNoStatsClient-NoTraceYesStatsServer

// Start: 3. NoTraceNoStatsClient-YesTraceNoStatsServer
func BenchmarkNoTraceNoStatsClientYesTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceNoStatsServer(b, 1)
}

func BenchmarkNoTraceNoStatsClientYesTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceNoStatsServer(b, 10)
}

func BenchmarkNoTraceNoStatsClientYesTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceNoStatsServer(b, 100)
}

func BenchmarkNoTraceNoStatsClientYesTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceNoStatsServer(b, 1000)
}

func benchmarkNoTraceNoStatsClientYesTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceNoStatsAddr, noTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 3. NoTraceNoStatsClient-YesTraceNoStatsServer

// Start: 4. NoTraceNoStatsClient-YesTraceYesStatsServer
func BenchmarkNoTraceNoStatsClientYesTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceYesStatsServer(b, 1)
}

func BenchmarkNoTraceNoStatsClientYesTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceYesStatsServer(b, 10)
}

func BenchmarkNoTraceNoStatsClientYesTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceYesStatsServer(b, 100)
}

func BenchmarkNoTraceNoStatsClientYesTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceNoStatsClientYesTraceYesStatsServer(b, 1000)
}

func benchmarkNoTraceNoStatsClientYesTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceYesStatsAddr, noTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 4. NoTraceNoStatsClient-YesTraceYesStatsServer

// Start: 5. NoTraceYesStatsClient-NoTraceNoStatsServer
func BenchmarkNoTraceYesStatsClientNoTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceNoStatsServer(b, 1)
}

func BenchmarkNoTraceYesStatsClientNoTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceNoStatsServer(b, 10)
}

func BenchmarkNoTraceYesStatsClientNoTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceNoStatsServer(b, 100)
}

func BenchmarkNoTraceYesStatsClientNoTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceNoStatsServer(b, 1000)
}

func benchmarkNoTraceYesStatsClientNoTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, noTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 5. NoTraceYesStatsClient-NoTraceNoStatsServer

// Start: 6. NoTraceYesStatsClient-NoTraceYesStatsServer
func BenchmarkNoTraceYesStatsClientNoTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceYesStatsServer(b, 1)
}

func BenchmarkNoTraceYesStatsClientNoTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceYesStatsServer(b, 10)
}

func BenchmarkNoTraceYesStatsClientNoTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceYesStatsServer(b, 100)
}

func BenchmarkNoTraceYesStatsClientNoTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientNoTraceYesStatsServer(b, 1000)
}

func benchmarkNoTraceYesStatsClientNoTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, noTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 6. NoTraceYesStatsClient-NoTraceYesStatsServer

// Start: 7. NoTraceYesStatsClient-YesTraceNoStatsServer
func BenchmarkNoTraceYesStatsClientYesTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceNoStatsServer(b, 1)
}

func BenchmarkNoTraceYesStatsClientYesTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceNoStatsServer(b, 10)
}
func BenchmarkNoTraceYesStatsClientYesTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceNoStatsServer(b, 100)
}
func BenchmarkNoTraceYesStatsClientYesTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceNoStatsServer(b, 1000)
}

func benchmarkNoTraceYesStatsClientYesTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, yesTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 7. NoTraceYesStatsClient-YesTraceNoStatsServer

// Start: 8. NoTraceYesStatsClient-YesTraceYesStatsServer
func BenchmarkNoTraceYesStatsClientYesTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceYesStatsServer(b, 1)
}

func BenchmarkNoTraceYesStatsClientYesTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceYesStatsServer(b, 10)
}

func BenchmarkNoTraceYesStatsClientYesTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceYesStatsServer(b, 100)
}

func BenchmarkNoTraceYesStatsClientYesTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkNoTraceYesStatsClientYesTraceYesStatsServer(b, 1000)
}

func benchmarkNoTraceYesStatsClientYesTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceYesStatsAddr, noTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 8. NoTraceYesStatsClient-YesTraceYesStatsServer

// Start: 9. YesTraceNoStatsClient-NoTraceNoStatsServer
func BenchmarkYesTraceNoStatsClientNoTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceNoStatsServer(b, 1)
}

func BenchmarkYesTraceNoStatsClientNoTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceNoStatsServer(b, 10)
}

func BenchmarkYesTraceNoStatsClientNoTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceNoStatsServer(b, 100)
}

func BenchmarkYesTraceNoStatsClientNoTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceNoStatsServer(b, 1000)
}

func benchmarkYesTraceNoStatsClientNoTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceNoStatsAddr, yesTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 9. YesTraceNoStatsClient-NoTraceNoStatsServer

// Start: 10. YesTraceNoStatsClient-NoTraceYesStatsServer
func BenchmarkYesTraceNoStatsClientNoTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceYesStatsServer(b, 1)
}

func BenchmarkYesTraceNoStatsClientNoTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceYesStatsServer(b, 10)
}

func BenchmarkYesTraceNoStatsClientNoTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceYesStatsServer(b, 100)
}

func BenchmarkYesTraceNoStatsClientNoTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientNoTraceYesStatsServer(b, 1000)
}

func benchmarkYesTraceNoStatsClientNoTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, yesTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 10. YesTraceNoStatsClient-NoTraceYesStatsServer

// Start: 11. YesTraceNoStatsClient-YesTraceNoStatsServer
func BenchmarkYesTraceNoStatsClientYesTraceNoStatsServer1QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceNoStatsServer(b, 1)
}

func BenchmarkYesTraceNoStatsClientYesTraceNoStatsServer10QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceNoStatsServer(b, 10)
}

func BenchmarkYesTraceNoStatsClientYesTraceNoStatsServer100QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceNoStatsServer(b, 100)
}

func BenchmarkYesTraceNoStatsClientYesTraceNoStatsServer1000QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceNoStatsServer(b, 1000)
}

func benchmarkYesTraceNoStatsClientYesTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceNoStatsAddr, yesTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 11. YesTraceNoStatsClient-YesTraceNoStatsServer

// Start: 12. YesTraceNoStatsClient-YesTraceYesStatsServer
func BenchmarkYesTraceNoStatsClientYesTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceYesStatsServer(b, 1)
}

func BenchmarkYesTraceNoStatsClientYesTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceYesStatsServer(b, 10)
}

func BenchmarkYesTraceNoStatsClientYesTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceYesStatsServer(b, 100)
}

func BenchmarkYesTraceNoStatsClientYesTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkYesTraceNoStatsClientYesTraceYesStatsServer(b, 1000)
}

func benchmarkYesTraceNoStatsClientYesTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceYesStatsAddr, yesTraceNoStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 12. YesTraceNoStatsClient-YesTraceYesStatsServer

// Start: 13. YesTraceYesStatsClient-NoTraceNoStatsServer
func benchmarkYesTraceYesStatsClientNoTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceNoStatsAddr, yesTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 13. YesTraceYesStatsClient-NoTraceNoStatsServer

// Start: 14. YesTraceYesStatsClient-NoTraceYesStatsServer
func benchmarkYesTraceYesStatsClientNoTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(noTraceYesStatsAddr, yesTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 14. YesTraceYesStatsClient-NoTraceYesStatsServer

// Start: 15. YesTraceYesStatsClient-YesTraceNoStatsServer
func benchmarkYesTraceYesStatsClientYesTraceNoStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceNoStatsAddr, yesTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 15. YesTraceYesStatsClient-YesTraceNoStatsServer

// Start: 16. YesTraceYesStatsClient-YesTraceYesStatsServer
func BenchmarkYesTraceYesStatsClientYesTraceYesStatsServer1QPS(b *testing.B) {
	benchmarkYesTraceYesStatsClientYesTraceYesStatsServer(b, 1)
}

func BenchmarkYesTraceYesStatsClientYesTraceYesStatsServer10QPS(b *testing.B) {
	benchmarkYesTraceYesStatsClientYesTraceYesStatsServer(b, 10)
}

func BenchmarkYesTraceYesStatsClientYesTraceYesStatsServer100QPS(b *testing.B) {
	benchmarkYesTraceYesStatsClientYesTraceYesStatsServer(b, 100)
}

func BenchmarkYesTraceYesStatsClientYesTraceYesStatsServer1000QPS(b *testing.B) {
	benchmarkYesTraceYesStatsClientYesTraceYesStatsServer(b, 1000)
}

func benchmarkYesTraceYesStatsClientYesTraceYesStatsServer(b *testing.B, qps int) {
	conn, err := grpc.Dial(yesTraceYesStatsAddr, yesTraceYesStatsClient, grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Creating traced gRPC connection: %v", err)
	}
	defer conn.Close()
	runWithConn(b, conn, qps)
}

// End: 16. YesTraceYesStatsClient-YesTraceYesStatsServer
