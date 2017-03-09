// download protobuf compiler from: https://github.com/google/protobuf/releases for example protoc-3.2.0-linux-x86_64
// extract it. You should have a file called ".../bin/protoc"

git clone https://github.com/google/instrumentation-proto.git
go_source="GOPATH/src/"
output_directory=$go_source"github.com/google/instrumentation-go/instrumentation-go-grpc/generated_proto"
input_directory=$go_source"github.com/google/instrumentation-proto/stats"
protoc --go_out=$output_directory --proto_path="${input_directory}" "${input_directory}/census.proto"
protoc --go_out=$output_directory input_directory/stats_context.proto








// go build ./instrumentation-go-grpc/adapter/
// go test -v -bench=. ./instrumentation-go-grpc/adapter/
