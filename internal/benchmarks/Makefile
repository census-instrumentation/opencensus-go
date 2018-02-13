all: gen visual

gen:
	protoc -I proto proto/proto.proto --go_out=plugins=grpc:proto
visual:
	go run main.go
