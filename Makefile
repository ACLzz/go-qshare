FLAGS = -gcflags=all=-l -parallel=1

dep:
	go mod tidy

proto:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc --proto_path=protobuf --go_out=protobuf protobuf/*.proto --go_opt=paths=import

clean:
	rm -rfv ./protobuf/gen

test:
	go test $(FLAGS) ./...

test-no-cache:
	go test $(FLAGS) -count=1 ./...

