.PHONY: fmt test vet run build

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

run:
	go run ./cmd/server

build:
	go build ./...
