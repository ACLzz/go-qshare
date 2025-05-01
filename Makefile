dep:
	go mod tidy

test:
	go test -gcflags=all=-l -parallel=1 ./...