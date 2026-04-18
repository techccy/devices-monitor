.PHONY: all clean build test run-server run-agent

# Variables
BINARY_DIR=bin
SERVER_BINARY=$(BINARY_DIR)/ccy-server
AGENT_BINARY=$(BINARY_DIR)/ccy-agent
CLI_BINARY=$(BINARY_DIR)/ccy

# Go flags
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

all: build

build: clean
	@echo "Building all components..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(LDFLAGS) -o $(SERVER_BINARY) cmd/server/main.go
	$(GO) build $(LDFLAGS) -o $(AGENT_BINARY) cmd/agent/main.go
	$(GO) build $(LDFLAGS) -o $(CLI_BINARY) cmd/cli/main.go
	@echo "Build complete!"

build-windows:
	@echo "Building Windows binaries..."
	@mkdir -p $(BINARY_DIR)
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_DIR)/ccy-server.exe cmd/server/main.go
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_DIR)/ccy-agent.exe cmd/agent/main.go
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY_DIR)/ccy.exe cmd/cli/main.go
	@echo "Windows build complete!"

clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@echo "Clean complete!"

test:
	@echo "Running tests..."
	$(GO) test ./...

run-server:
	@echo "Starting server..."
	$(GO) run cmd/server/main.go

run-agent:
	@echo "Starting agent..."
	$(GO) run cmd/agent/main.go

tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

deps:
	@echo "Installing dependencies..."
	$(GO) mod download

fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

vet:
	@echo "Vetting code..."
	$(GO) vet ./...
