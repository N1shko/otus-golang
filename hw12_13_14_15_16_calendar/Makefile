BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build: generate
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) -config ./configs/config.yaml

build-img: generate
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run $(DOCKER_IMG)

version: build
	$(BIN) version

generate:
	protoc --go_out=internal/server/grpc/pb --go_opt=paths=source_relative \
       	 --go-grpc_out=internal/server/grpc/pb --go-grpc_opt=paths=source_relative \
			   ./EventService.proto

test:
	go test -race ./internal/...

migrate:
	goose postgres "user=calendar dbname=calendar sslmode=disable password=calendar host=127.0.0.1" up -dir migrations

install-lint-deps: build
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2

lint: install-lint-deps
	golangci-lint run --fix ./...

.PHONY: build run build-img run-img version test lint
