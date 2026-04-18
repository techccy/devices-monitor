# CCY Terminal Remote Monitoring and Management System

A command-line based remote device management tool for monitoring and controlling personal devices from restricted environments like school computers.

## Architecture

The system consists of three main components:

- **Agent**: Background daemon that runs on monitored devices, collecting system metrics and executing commands
- **Server**: Central node that handles authentication, device management, and command routing
- **CLI**: Command-line interface for users to monitor and control devices

## Features

- **Zero-dependency**: Single-file executables for easy deployment
- **Real-time monitoring**: 5-minute heartbeat intervals with system snapshots
- **Command execution**: Execute remote commands on monitored devices
- **SSH tunneling**: Interactive terminal access to devices
- **Offline support**: View last known status when device is offline
- **Auto-reconnection**: Agent automatically reconnects with exponential backoff
- **Secure**: JWT-based authentication, optional TLS support

## Quick Start

### 1. Start the Server

```bash
./bin/ccy-server -addr :8080 -secret your-secret-key
```

With TLS:

```bash
./bin/ccy-server -tls-addr :8443 -cert server.crt -key server.key -secret your-secret-key
```

### 2. Start the Agent (on monitored device)

```bash
./bin/ccy-agent -server ws://localhost:8080/api/ws -id your-device-id
```

### 3. Use the CLI

Login:
```bash
./bin/ccy login -u user@example.com -p password
```

List devices:
```bash
./bin/ccy ls
```

Check device status:
```bash
./bin/ccy status <device-id>
```

Query network info:
```bash
./bin/ccy net <device-id>
```

SSH to device:
```bash
./bin/ccy ssh <device-id>
```

Logout:
```bash
./bin/ccy logout
```

## Building

Build all components:

```bash
go build -o bin/ccy-server cmd/server/main.go
go build -o bin/ccy-agent cmd/agent/main.go
go build -o bin/ccy cmd/cli/main.go
```

## Development

### Project Structure

```
devices-monitor/
├── cmd/
│   ├── agent/      # Agent entry point
│   ├── server/     # Server entry point
│   └── cli/        # CLI entry point
├── internal/
│   ├── agent/      # Agent implementation
│   ├── server/     # Server implementation
│   └── common/     # Shared data models
├── pkg/
│   ├── storage/    # In-memory storage
│   ├── auth/       # JWT authentication
│   └── ssh/        # SSH tunneling
└── bin/            # Compiled binaries
```

### Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/gorilla/websocket` - WebSocket connections

## API Endpoints

### Authentication

- `POST /api/login` - Login and receive JWT token

### Devices

- `GET /api/devices` - List all devices (requires auth)
- `GET /api/devices/{id}` - Get device status (requires auth)
- `POST /api/devices/{id}` - Send command to device (requires auth)

### WebSocket

- `GET /api/ws?device_id={id}` - WebSocket connection for agents

## Security Considerations

1. Change the default JWT secret in production
2. Use TLS for production deployments
3. Implement proper password hashing (currently using plain text for simplicity)
4. Add rate limiting to prevent brute force attacks
5. Consider adding device registration/authorization flow

## TODO

- [x] Implement proper password hashing
- [x] Add TLS certificate generation script
- [x] Add proper error handling and logging
- [x] Implement device registration flow
- [x] Add multi-user support with proper device authorization
- [x] Add configuration file support
- [x] Implement systemd/supervisord integration for agent
- [x] Add database persistence (PostgreSQL/MongoDB)
- [x] Add monitoring and metrics
- [x] Add unit tests and integration tests

## License

MIT