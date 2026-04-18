# Quick Start Guide

Get up and running with CCY Device Monitor in 5 minutes.

## Prerequisites

- Go 1.19 or higher
- Basic terminal knowledge

## Installation

### Option 1: Build from Source

```bash
# Clone the repository (if not already done)
cd devices-monitor

# Build all components
make build

# Or use the build script
./build.sh
```

### Option 2: Use Pre-built Binaries

If binaries are available, you can download them directly from the `bin/` directory.

## Setup

### 1. Start the Server

In one terminal:

```bash
./bin/ccy-server -addr :8080 -secret my-super-secret-key
```

You should see:
```
2024/04/18 15:19:00 Starting server on :8080
```

### 2. Start an Agent

In another terminal:

```bash
./bin/ccy-agent -server ws://localhost:8080/api/ws -id my-laptop
```

You should see:
```
2024/04/18 15:19:05 Starting agent for device my-laptop-mac
2024/04/18 15:19:05 Connected to server
```

### 3. Use the CLI

In a third terminal:

```bash
# Login
./bin/ccy login -u test@example.com -p testpass
# Output: Login successful

# List devices
./bin/ccy ls
# Output:
# Devices:
# Name                 ID                              Status
# ------------------------------------------------------------
# my-laptop-mac        <device-id>                     Online

# Check device status
./bin/ccy status <device-id>

# Query network info
./bin/ccy net <device-id>

# SSH to device
./bin/ccy ssh <device-id>

# Logout when done
./bin/ccy logout
```

## Common Tasks

### Monitoring Multiple Devices

```bash
# Terminal 1: Start server
./bin/ccy-server -addr :8080 -secret my-key

# Terminal 2: Start agent on device 1
./bin/ccy-agent -server ws://localhost:8080/api/ws -id laptop-home

# Terminal 3: Start agent on device 2
./bin/ccy-agent -server ws://localhost:8080/api/ws -id macbook-office

# Terminal 4: Use CLI to monitor all devices
./bin/ccy login -u test@example.com -p testpass
./bin/ccy ls
```

### Using TLS (Production Setup)

```bash
# Generate certificates
./generate-certs.sh

# Start server with TLS
./bin/ccy-server -tls-addr :8443 -cert certs/server.crt -key certs/server.key -secret my-key

# Connect agent to TLS server
./bin/ccy-agent -server wss://localhost:8443/api/ws -id my-laptop

# Use CLI with TLS server
./bin/ccy login -u test@example.com -p testpass -server https://localhost:8443
```

### Cross-Platform Use

```bash
# Build for all platforms
make build-windows

# Copy binaries to Windows machine
cp bin/ccy.exe /path/to/windows/machine/
cp bin/ccy-agent.exe /path/to/windows/machine/
cp bin/ccy-server.exe /path/to/windows/machine/

# Run on Windows
ccy-server.exe -addr :8080 -secret my-key
```

## Troubleshooting

### Agent cannot connect

1. Check if server is running: `./bin/ccy-server -addr :8080 -secret my-key`
2. Verify WebSocket URL: `ws://localhost:8080/api/ws`
3. Check firewall settings
4. Verify device ID is correct

### Authentication fails

1. Check JWT secret matches between server and CLI
2. Verify login credentials (email/password)
3. Clear old tokens: `rm -rf ~/.ccy/`

### Device shows offline

1. Check if agent is running and connected
2. Verify network connectivity
3. Check server logs for connection errors
4. Ensure device ID matches

### SSH connection issues

1. Verify device is online
2. Check WebSocket connection is active
3. Ensure agent has proper permissions
4. Try reconnecting the agent

## Next Steps

- Read the [full documentation](README.md) for advanced features
- Check [usage examples](EXAMPLES.md) for more scenarios
- Review the [implementation details](IMPLEMENTATION.md) for technical information

## Support

For issues or questions, please refer to the project documentation or check the build requirements in `buildfiles/PRD.md`.
