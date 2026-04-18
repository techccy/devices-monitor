#!/bin/bash

echo "Building CCY Terminal Remote Monitoring and Management System..."

# Create bin directory if it doesn't exist
mkdir -p bin

# Build server
echo "Building server..."
go build -o bin/ccy-server cmd/server/main.go

# Build agent
echo "Building agent..."
go build -o bin/ccy-agent cmd/agent/main.go

# Build CLI
echo "Building CLI..."
go build -o bin/ccy cmd/cli/main.go

# Build Windows binaries
echo "Building Windows binaries..."
GOOS=windows GOARCH=amd64 go build -o bin/ccy-server.exe cmd/server/main.go
GOOS=windows GOARCH=amd64 go build -o bin/ccy-agent.exe cmd/agent/main.go
GOOS=windows GOARCH=amd64 go build -o bin/ccy.exe cmd/cli/main.go

echo "Build complete!"
echo ""
echo "Binaries:"
ls -lh bin/
