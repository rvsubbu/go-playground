GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo "0.0.1")
GIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
ENV ?= dev

CGO_ENABLED := 0
LDFLAGS = -X main.version=$(VERSION) \
          -X main.gitSHA=$(GIT_SHA) \
          -X main.buildTime=$(BUILD_TIME) \
          -X main.env=$(ENV)

.PHONY: all build_chi_server build_mux_server run_chi_server run_mux_server clean tidy test info

all: build_chi_server build_mux_server

build_chi_server:
	mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "$(LDFLAGS)" \
		-o bin/basic_chi_server-$(GOOS)-$(GOARCH) ./basic_chi_server

build_mux_server:
	mkdir -p bin
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build -ldflags "$(LDFLAGS)" \
		-o bin/basic_mux_server-$(GOOS)-$(GOARCH) ./basic_mux_server

run_chi_server:
	go run -ldflags "$(LDFLAGS)" ./basic_chi_server

run_mux_server:
	go run -ldflags "$(LDFLAGS)" ./basic_mux_server

tidy:
	go mod tidy

test:
	go test -race -coverpkg=./... -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

# Display build configuration
info:
	@echo "Version:    $(VERSION)"
	@echo "Git SHA:    $(GIT_SHA)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Environment: $(ENV)"
	@echo "GOOS:       $(GOOS)"
	@echo "GOARCH:     $(GOARCH)"
	@echo "CGO:        $(CGO_ENABLED)"

clean:
	rm -rf bin
	go clean -testcache
