REPO_ROOT := $(shell pwd)
BINARY := $(REPO_ROOT)/bin/processor

.PHONY: up build test fmt clean

up:
	docker-compose -f docker-compose.yml up -d

build:
	go build -o $(BINARY) ./cmd/processor

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -rf bin/
