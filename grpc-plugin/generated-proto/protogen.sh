// 1- download protobuf compiler from: https://github.com/google/protobuf/releases for example protoc-3.2.0-linux-x86_64
// 2- extract it. You should have a file called "...protoc-3.2.0-linux-x86_64/bin/protoc"
// 3- download and add protoc-gen-go (plugin for protoc). "https://github.com/golang/protobuf" https://github.com/golang/protobuf/protoc-gen-go/main.go
// 4- run protoc --go_out=output_directory input_directory/file.proto

cd $GOPATH/src
git clone https://github.com/google/instrumentation-proto.git
git clone https://github.com/grpc/grpc-proto.git

go_source="GOPATH/src/"
output_directory=$go_source"github.com/google/instrumentation-go/grpc-plugin/generated-proto"
input_directory=$go_source"github.com/google/instrumentation-proto"
~/Documents/protoc-3.2.0-linux-x86_64/bin/protoc --go_out=plugins=grpc:$output_directory --proto_path="${input_directory}" $input_directory/stats/*.proto
~/Documents/protoc-3.2.0-linux-x86_64/bin/protoc --go_out=plugins=grpc:$output_directory --proto_path="${input_directory}" $input_directory/service/monitoring.proto





// go build ./grpc-plugin/collection-plugin/
// go test -v -bench=. ./instrumentation-go-grpc/adapter/
// go test -v -bench=. ./stats/




// http://www.grpc.io/docs/quickstart/go.html
// go get google.golang.org/grpc
// download released protoc from https://github.com/google/protobuf/releases --> https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-osx-x86_64.zip
// add it to PATH (~/.profile; source profile OR /etc/environment; source /etc/environment)
// go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
// export PATH=$PATH:$GOPATH/bin
// protoc --go_out=plugins=grpc:. *.proto
// go get github.com/google/instrumentation-proto
// go get github.com/grpc/grpc-proto


go_source="GOPATH/src/"
output_directory=$go_source"github.com/google/instrumentation-go/grpc-plugin/generated-proto"
input_directory=$go_source"github.com/google/instrumentation-proto"
~/Documents/protoc-3.2.0-linux-x86_64/bin/protoc --go_out=$output_directory --proto_path="${input_directory}" $input_directory/stats/*.proto
~/Documents/protoc-3.2.0-linux-x86_64/bin/protoc --go_out=$output_directory --proto_path="${input_directory}" $input_directory/service/monitoring.proto
