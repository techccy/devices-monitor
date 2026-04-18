# CCY Terminal Remote Monitoring and Management System

A command-line based remote device management tool for monitoring and controlling personal devices from restricted environments like school computers.

## Architecture

The system consists of three main components:

- **Agent**: Background daemon that runs on monitored devices, collecting system metrics and executing commands
- **Server**: Central node that handles authentication, device management, command routing, and WebRTC signaling
- **CLI**: Command-line interface for users to monitor and control devices

### Key Technologies

- **WebSocket**: Real-time bidirectional communication between agent and server
- **WebRTC**: Low-latency P2P data channels for interactive terminal sessions
- **PTY (Pseudo-terminal)**: Full terminal emulation support for remote shell access
- **STUN/TURN**: NAT traversal support for direct connections in restrictive network environments

## Features

- **Zero-dependency**: Single-file executables for easy deployment
- **Real-time monitoring**: 30-second bidirectional heartbeat with detailed system snapshots
- **WebRTC P2P Terminal**: Ultra-low latency (<100ms) remote shell via direct P2P connection
- **Full Terminal Support**: PTY-based terminal emulation with window size synchronization
- **Rich Metrics Collection**: CPU load, memory usage, network latency, process count, uptime
- **Command execution**: Execute remote commands on monitored devices
- **Offline support**: View last known status when device is offline
- **Auto-reconnection**: Agent automatically reconnects with exponential backoff (2s, 4s, 8s, max 5min)
- **Secure**: JWT-based authentication, optional TLS support
- **NAT Traversal**: STUN/TURN support for connectivity through firewalls and restrictive networks

## Quick Start

### 1. Start the Server

```bash
./bin/ccy-server -addr :8080 -secret your-secret-key
```

With TLS:

```bash
./bin/ccy-server -tls-addr :8443 -cert server.crt -key server.key -secret your-secret-key
```

With TURN server (for restrictive NAT environments):

```bash
./bin/ccy-server -addr :8080 -secret your-secret-key \
  -turn-uri turn:your-turn-server.com:3478 \
  -turn-user your-username \
  -turn-pass your-password
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

SSH to device (now uses WebRTC P2P):
```bash
./bin/ccy ssh <device-id>
```

This establishes a direct P2P connection with:
- **< 100ms latency** for character input
- Full terminal emulation (vim, top, htop work perfectly)
- Automatic window size synchronization
- Support for Tab completion and Ctrl+C

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

Or use the provided build script:

```bash
./build.sh
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
│   ├── agent/      # Agent implementation (WebRTC PTY, metrics collection)
│   ├── server/     # Server implementation (WebSocket, WebRTC signaling)
│   ├── common/     # Shared data models (WebRTC messages, snapshots)
│   └── cli/        # CLI implementation (PTY raw mode, WebRTC client)
├── pkg/
│   ├── storage/    # In-memory storage with PostgreSQL support
│   ├── auth/       # JWT authentication
│   ├── config/     # Configuration management (TURN server settings)
│   ├── logger/     # Logging utilities
│   ├── metrics/    # System metrics collection
│   └── password/   # Password hashing
├── buildfiles/     # PRD and documentation
├── config/         # Default configuration files
└── bin/            # Compiled binaries
```

### Dependencies

- `github.com/golang-jwt/jwt/v5` - JWT authentication
- `github.com/gorilla/websocket` - WebSocket connections
- `github.com/pion/webrtc/v3` - WebRTC implementation
- `github.com/creack/pty` - PTY (pseudo-terminal) management
- `golang.org/x/sys/unix` - Unix system calls for terminal control
- `github.com/lib/pq` - PostgreSQL driver
- `golang.org/x/crypto` - Password hashing utilities

## API Endpoints

### Authentication

- `POST /api/register` - Register new user
- `POST /api/login` - Login and receive JWT token

### Devices

- `GET /api/devices` - List all devices (requires auth)
- `GET /api/devices/{id}` - Get device status (requires auth)
- `POST /api/devices/{id}` - Send command to device (requires auth)
- `POST /api/devices` - Register new device (requires auth)

### WebSocket

- `GET /api/ws?device_id={id}&device_key={key}` - WebSocket connection for agents
  - Handles bidirectional heartbeat (PING/PONG every 30 seconds)
  - Receives system snapshots and command responses
  - Forwards WebRTC signaling messages

### WebRTC Signaling

- `POST /api/webrtc/offer` - Initiate WebRTC connection (requires auth)
  - Request body: `{ "type": "OFFER", "device_id": "...", "session_id": "...", "sdp": "..." }`
  - Response: `{ "session_id": "..." }`

- `POST /api/webrtc/answer` - Accept WebRTC connection (requires auth)
  - Request body: `{ "type": "ANSWER", "device_id": "...", "session_id": "...", "sdp": "..." }`

- `POST /api/webrtc/ice` - Exchange ICE candidates (requires auth)
  - Request body: `{ "type": "ICE_CANDIDATE", "device_id": "...", "session_id": "...", "candidate": "...", "sdp_mline_index": 0, "sdp_mid": "0" }`

### Metrics and Health

- `GET /metrics` - System metrics in JSON format
- `GET /health` - Health check endpoint

## Security Considerations

1. **Authentication**: Always use TLS in production to protect JWT tokens and credentials
2. **Secret Management**: Change default JWT secret in production and use environment variables
3. **Password Security**: Passwords are hashed using bcrypt before storage
4. **Device Authorization**: Devices are uniquely keyed and authorized per user
5. **Rate Limiting**: Consider implementing rate limiting on authentication endpoints
6. **TURN Credentials**: Secure TURN server credentials and rotate them periodically
7. **Network Isolation**: Restrict WebSocket and WebRTC endpoints to trusted networks when possible
8. **Audit Logging**: Review logs regularly for suspicious activity

## TODO

### Completed (Phase 1 & 2)
- [x] Implement proper password hashing (bcrypt)
- [x] Add TLS certificate generation script
- [x] Add proper error handling and logging
- [x] Implement device registration flow
- [x] Add multi-user support with proper device authorization
- [x] Add configuration file support
- [x] Implement systemd/supervisord integration for agent
- [x] Add database persistence (PostgreSQL)
- [x] Add monitoring and metrics
- [x] Add unit tests and integration tests
- [x] Implement WebSocket bidirectional heartbeat (30s)
- [x] Add WebRTC-based P2P terminal access
- [x] Implement PTY-based terminal emulation
- [x] Add window size synchronization
- [x] Optimize exponential backoff reconnection (2s, 4s, 8s, max 5min)
- [x] Add TURN server configuration support
- [x] Implement real-time metrics collection (CPU, memory, network, processes, uptime)

### Future Enhancements (Phase 3)
- [ ] File transfer over WebRTC data channels
- [ ] Multiple concurrent terminal sessions per device
- [ ] Session recording and playback
- [ ] Web-based management interface
- [ ] Mobile app support
- [ ] Advanced filtering and search for metrics
- [ ] Alert and notification system
- [ ] Device grouping and bulk operations
- [ ] Integration with monitoring tools (Prometheus, Grafana)
- [ ] Load testing and performance optimization
- [ ] Docker and Kubernetes deployment guides
- [ ] Automated CI/CD pipeline

## Performance Characteristics

### Real-Time Communication
- **Heartbeat Interval**: 30 seconds (bidirectional PING/PONG)
- **Connection Timeout**: 90 seconds (automatic offline detection)
- **Reconnection**: Exponential backoff (2s, 4s, 8s, ..., max 5min)

### WebRTC Terminal
- **Latency**: < 100ms for character input in most networks
- **Throughput**: Optimized for low-bandwidth connections
- **NAT Traversal**: Supports STUN and TURN for firewall traversal
- **Fallback**: Gracefully handles P2P connection failures

### Metrics Collection
- **CPU Load**: Current usage percentage
- **Memory Usage**: Used memory percentage
- **Network Latency**: Round-trip time to server
- **Process Count**: Number of running processes
- **Uptime**: System uptime formatted as human-readable string

## Configuration

### Server Configuration
Create a JSON configuration file (default: `~/.ccy/server.json`):

```json
{
  "addr": ":8080",
  "tls_addr": ":8443",
  "cert_file": "server.crt",
  "key_file": "server.key",
  "secret": "your-jwt-secret-key",
  "turn_server": {
    "uri": "turn:your-turn-server.com:3478",
    "username": "your-username",
    "password": "your-password"
  }
}
```

### Agent Configuration
Create a JSON configuration file (default: `~/.ccy/agent.json`):

```json
{
  "server_url": "ws://localhost:8080/api/ws",
  "device_id": "your-device-id",
  "device_key": "your-device-key",
  "heartbeat": 30,
  "turn_server": {
    "uri": "turn:your-turn-server.com:3478",
    "username": "your-username",
    "password": "your-password"
  }
}
```

### CLI Configuration
Create a JSON configuration file (default: `~/.ccy/cli.json`):

```json
{
  "server_url": "http://localhost:8080",
  "turn_server": {
    "uri": "turn:your-turn-server.com:3478",
    "username": "your-username",
    "password": "your-password"
  }
}
```

## License

MIT