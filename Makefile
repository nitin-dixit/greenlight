fmt:
	go fmt ./...

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build -o bin/api ./cmd/api

run:
	go run ./cmd/api

check: fmt vet lint test

clean:
	rm -rf bin

