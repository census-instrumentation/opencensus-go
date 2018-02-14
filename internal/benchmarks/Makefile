all: gen test

benchmark:
	go test -v -run="^6" -bench=. ./...
gen:
	protoc -I proto proto/proto.proto --go_out=plugins=grpc:proto
visual:
	go run main.go
