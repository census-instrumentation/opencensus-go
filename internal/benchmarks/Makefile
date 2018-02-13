all: gen visual

gen:
	protoc -I defs defs/defs.proto --go_out=plugins=grpc:defs
visual:
	go run main.go
