FLAGS = -gcflags=all=-l -parallel=1
PROTO_DIR = internal/protobuf
PROTO_GEN_DIR = $(PROTO_DIR)/gen
PROTO = protoc \
		--proto_path=$(PROTO_DIR) \
		--go_opt=paths=source_relative

dep:
	go mod tidy

# TODO: make it work from docker run
proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

	mkdir -p ./$(PROTO_GEN_DIR)/securegcm
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/securegcm ./$(PROTO_DIR)/device_to_device_messages.proto ./$(PROTO_DIR)/securegcm.proto ./$(PROTO_DIR)/ukey.proto
	
	mkdir -p ./$(PROTO_GEN_DIR)/connections
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/connections ./$(PROTO_DIR)/offline_wire_formats.proto

	mkdir -p ./$(PROTO_GEN_DIR)/securemessage
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/securemessage ./$(PROTO_DIR)/securemessage.proto
	
	mkdir -p ./$(PROTO_GEN_DIR)/sharing
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/sharing ./$(PROTO_DIR)/wire_format.proto

# TODO: make it work from docker run
mock:
	go install go.uber.org/mock/mockgen@latest

fmt:
	gofmt ./...

lint:
	docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v2.1.6 golangci-lint run

clean:
	rm -rfv ./$(PROTO_GEN_DIR)
	rm -rvf ./mocks

test:
	go test $(FLAGS) ./... -coverprofile=coverage.out

ci-test:
	CI=true go test $(FLAGS) ./... -coverprofile=coverage.out

test-no-cache:
	go test $(FLAGS) -count=1 ./...

