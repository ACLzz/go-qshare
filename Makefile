FLAGS = -gcflags=all=-l -parallel=1

dep:
	go mod tidy

test:
	go test $(FLAGS) ./...

test-no-cache:
	go test $(FLAGS) -count=1 ./...

