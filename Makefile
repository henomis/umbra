BINARY ?= umbra
CMD ?= ./cmd
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: help build run test fmt vet tidy clean install

help:
	@echo "Available targets:"
	@echo "  make build   - Compile the CLI into ./bin/$(BINARY)"
	@echo "  make run     - Build and run the CLI (pass ARGS=...)"
	@echo "  make test    - Execute go test ./..."
	@echo "  make fmt     - Format all Go sources"
	@echo "  make vet     - Run go vet ./..."
	@echo "  make tidy    - go mod tidy"
	@echo "  make clean   - Remove build artifacts"
	@echo "  make install - Install the CLI into GOPATH/bin"

build:
	@mkdir -p bin
	GO111MODULE=on go build -o bin/$(BINARY) $(CMD)

run: build
	./bin/$(BINARY) $(ARGS)

test:
	GO111MODULE=on go test ./...

fmt:
	gofmt -s -w $(GOFILES)

vet:
	GO111MODULE=on go vet ./...

tidy:
	GO111MODULE=on go mod tidy

clean:
	rm -rf bin

install:
	GO111MODULE=on go install $(CMD)
