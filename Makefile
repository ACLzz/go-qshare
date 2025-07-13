TEST_FLAGS = -coverpkg=./... -gcflags=all=-l -parallel=1
PROTO_DIR = internal/protobuf
PROTO_GEN_DIR = $(PROTO_DIR)/gen
PROTO = protoc \
		--proto_path=$(PROTO_DIR) \
		--go_opt=paths=source_relative

dep:
	go mod tidy

proto:
	mkdir -p ./$(PROTO_GEN_DIR)/securegcm
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/securegcm ./$(PROTO_DIR)/device_to_device_messages.proto ./$(PROTO_DIR)/securegcm.proto ./$(PROTO_DIR)/ukey.proto
	
	mkdir -p ./$(PROTO_GEN_DIR)/connections
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/connections ./$(PROTO_DIR)/offline_wire_formats.proto

	mkdir -p ./$(PROTO_GEN_DIR)/securemessage
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/securemessage ./$(PROTO_DIR)/securemessage.proto
	
	mkdir -p ./$(PROTO_GEN_DIR)/sharing
	$(PROTO) --go_out=$(PROTO_GEN_DIR)/sharing ./$(PROTO_DIR)/wire_format.proto

tools:
	go install go.uber.org/mock/mockgen@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.2.1
	go install github.com/wadey/gocovmerge@latest

fmt:
	go fmt ./...

lint:
	golangci-lint run

mock:
	mockgen -destination internal/mock/log.go -source ./log.go -package mock Logger

test:
	go test $(TEST_FLAGS) -coverprofile=coverage1.out ./qserver/...
	go test $(TEST_FLAGS) -coverprofile=coverage2.out `go list ./... | grep -v qserver`
	gocovmerge coverage1.out coverage2.out > coverage.out
	rm -f coverage1.out coverage2.out

test-no-cache:
	go test $(TEST_FLAGS) -count=1 ./...

ci-tools:
	go install go.uber.org/mock/mockgen@latest
	go install github.com/wadey/gocovmerge@latest

ci-test:
	export CI=true
	go test $(TEST_FLAGS) -coverprofile=coverage1.txt ./qserver/...
	go test $(TEST_FLAGS) -coverprofile=coverage2.txt `go list ./... | grep -v qserver`
	gocovmerge coverage1.txt coverage2.txt > coverage.txt
	rm -f coverage1.txt coverage2.txt

clean:
	rm -rfv ./$(PROTO_GEN_DIR)
	rm -rvf ./internal/mock
