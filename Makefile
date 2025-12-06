REPO_ROOT := $(shell pwd)
BINARY := $(REPO_ROOT)/bin/processor

.PHONY: up down build test test-integration test-verbose fmt clean deps lint

up:
	docker-compose -f docker-compose.yml up -d

down:
	docker-compose -f docker-compose.yml down

build:
	go build -o $(BINARY) ./cmd/processor

test:
	go test ./...

test-integration:
	KAFKA_TEST=1 go test -v ./test/kafka_test/...

test-verbose:
	go test -v ./...

test-cover:
	go test -cover ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/

deps:
	go mod download
	go mod tidy

lint:
	go vet ./...
