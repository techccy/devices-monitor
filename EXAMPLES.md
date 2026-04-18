# CCY Device Monitor - Usage Examples

## Example 1: Basic Setup

### Step 1: Start the Server
```bash
# Terminal 1
./bin/ccy-server -addr :8080 -secret my-secret-key
```

### Step 2: Start an Agent on your monitored device
```bash
# Terminal 2
./bin/ccy-agent -server ws://localhost:8080/api/ws -id device-001
```

### Step 3: Use the CLI to login and manage devices
```bash
# Terminal 3
./bin/ccy login -u test@example.com -p testpass
./bin/ccy ls
./bin/ccy status device-001
./bin/ccy logout
```

## Example 2: Multiple Devices

### Monitor multiple devices by running multiple agents:
```bash
# Device 1
./bin/ccy-agent -server ws://localhost:8080/api/ws -id laptop-home

# Device 2 (in another terminal)
./bin/ccy-agent -server ws://localhost:8080/api/ws -id macbook-office
```

### View all devices:
```bash
./bin/ccy login -u test@example.com -p testpass
./bin/ccy ls
```

## Example 3: Remote Command Execution

### Query network information:
```bash
./bin/ccy net device-001
```

### SSH to device:
```bash
./bin/ccy ssh device-001
```

## Example 4: Using TLS

### Generate certificates:
```bash
./generate-certs.sh
```

### Start server with TLS:
```bash
./bin/ccy-server -tls-addr :8443 -cert certs/server.crt -key certs/server.key -secret my-secret-key
```

### Connect agent to TLS server:
```bash
./bin/ccy-agent -server wss://localhost:8443/api/ws -id device-001
```

### Use CLI with TLS server:
```bash
./bin/ccy login -u test@example.com -p testpass -server https://localhost:8443
```

## Example 5: Offline Device Query

### Query device that is offline:
```bash
./bin/ccy status device-001
# Output: Device is offline (showing last snapshot)
```

## Example 6: Building Windows Binaries

### Build for Windows:
```bash
./build.sh
# Or use Makefile:
make build-windows
```

### Copy to Windows machine and run:
```cmd
ccy-server.exe -addr :8080 -secret my-secret-key
ccy-agent.exe -server ws://localhost:8080/api/ws -id device-001
ccy.exe login -u test@example.com -p testpass
```

## Example 7: Development Workflow

### Run server in development mode:
```bash
make run-server
```

### Run agent in development mode:
```bash
make run-agent
```

### Build and test:
```bash
make build
make test
```

## Tips

1. **Device Registration**: Currently, devices need to be manually created in the server's storage. Future versions will support device registration.

2. **Security**: Change the default JWT secret key in production. Use TLS for all production deployments.

3. **Monitoring**: The agent sends heartbeats every 5 minutes. Check server logs for connection status.

4. **Troubleshooting**:
   - Check firewall settings if agent cannot connect
   - Verify WebSocket URL is correct
   - Check server logs for authentication errors
   - Ensure device ID is unique for each agent

5. **Cross-platform**: The CLI works on Windows, macOS, and Linux. The agent can run on any platform supported by Go.
