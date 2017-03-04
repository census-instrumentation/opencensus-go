// download protobuf compiler from: https://github.com/google/protobuf/releases for example protoc-3.2.0-linux-x86_64
// extract it. Yo ushould have a file called ".../bin/protoc"

// go build ./instrumentation-go-grpc/adapter/
// go test -v -bench=. ./instrumentation-go-grpc/adapter/

go get github.com/google3/instrumentation-proto

go_source="GOPATH/src/"
output_directory=$go_source"github.com/google/instrumentation-go-grpc/generated_proto"
input_directory=$go_source"github.com/google/instrumentation-proto/stats"
protoc --go_out=$output_directory --proto_path="${input_directory}" "${input_directory}/census.proto"
protoc --go_out=$output_directory input_directory/stats_context.proto
