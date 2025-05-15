FLAGS = -gcflags=all=-l -parallel=1
PROTO = protoc \
		--proto_path=protobuf \
		--go_opt=paths=source_relative

dep:
	go mod tidy

# TODO: make it work from docker run
proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

	mkdir -p ./protobuf/gen/securegcm
	$(PROTO) --go_out=protobuf/gen/securegcm ./protobuf/device_to_device_messages.proto ./protobuf/securegcm.proto ./protobuf/ukey.proto
	
	mkdir -p ./protobuf/gen/connections
	$(PROTO) --go_out=protobuf/gen/connections ./protobuf/offline_wire_formats.proto

	mkdir -p ./protobuf/gen/securemessage
	$(PROTO) --go_out=protobuf/gen/securemessage ./protobuf/securemessage.proto
	
	mkdir -p ./protobuf/gen/sharing
	$(PROTO) --go_out=protobuf/gen/sharing ./protobuf/wire_format.proto

# TODO: make it work from docker run
mock:
	go install go.uber.org/mock/mockgen@latest

clean:
	rm -rfv ./protobuf/gen
	rm -rvf ./mocks

test:
	go test $(FLAGS) ./...

test-no-cache:
	go test $(FLAGS) -count=1 ./...

