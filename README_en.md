# masque-vpn

Educational VPN implementation based on the MASQUE (CONNECT-IP) protocol using QUIC transport.

**This project is intended for educational purposes and research in modern network protocols.**

```bash
git clone https://github.com/cloudbridge-research/masque-vpn.git
```

## Features

- **Modern Protocols**: Built on QUIC and MASQUE CONNECT-IP (RFC 9484)
- **Custom MASQUE Implementation**: Own protocol implementation without external dependencies
- **Modular Architecture**: Server split into specialized components
- **REST API**: Full-featured API for management and monitoring
- **Mutual TLS Authentication**: Certificate-based client-server authentication
- **Cross-Platform**: Supports Windows, Linux, and macOS
- **IP Pool Management**: Automatic client IP allocation
- **Prometheus Monitoring**: Detailed metrics for performance analysis
- **Comprehensive Testing**: Unit, integration, and load tests

## Documentation

Comprehensive documentation for students and researchers:

- **[Student Guide (EN)](docs/development/student-guide.md)**: Laboratory assignments and research projects
- **[Student Guide (RU)](docs/development/student-guide.ru.md)**: Лабораторные работы и исследовательские проекты
- **[Architecture](docs/architecture/architecture.ru.md)**: System components and data flows
- **[API Documentation](docs/api/api.ru.md)**: REST API for management and monitoring
- **[Getting Started](docs/development/getting-started.md)**: Build and deployment instructions
- **[Monitoring](docs/monitoring/setup.md)**: Prometheus and Grafana setup

## Architecture

The system consists of four main components:
- **VPN Server**: Handles MASQUE CONNECT-IP requests and traffic routing
- **VPN Client**: Connects to server via QUIC and tunnels IP packets
- **REST API Server**: Client management and monitoring (port 8080)
- **Common Libraries**: Custom MASQUE implementation and utilities

## Quick Start

### 1. Build

```bash
# Build server
cd vpn_server && go build -o vpn-server .

# Build client
cd ../vpn_client && go build -o vpn-client .
```

### 2. Generate Certificates

```bash
cd cert

# Linux/macOS
./generate-test-certs.sh

# Windows
powershell -ExecutionPolicy Bypass -File generate-certs.ps1
```

### 3. Start Server (Local Testing)

```bash
cd vpn_server
./vpn-server -c config.server.local.toml
```

Server will start on:
- MASQUE VPN: `127.0.0.1:4433`
- REST API: `127.0.0.1:8080`

### 4. Test API

```bash
# Health check
curl http://127.0.0.1:8080/health

# Server status
curl http://127.0.0.1:8080/api/v1/status

# Prometheus metrics
curl http://127.0.0.1:8080/metrics
```

### 5. Start Client

```bash
cd vpn_client
# Edit config.client.toml: server_addr = "127.0.0.1:4433"
./vpn-client -c config.client.toml
```

### 6. Test MASQUE Protocol

```bash
# Functional test
go run test_masque_connection.go
```

## Configuration

### Local Testing

For local testing, use the provided configurations:

**Server** (`config.server.local.toml`):
```toml
listen_addr = "127.0.0.1:4433"
assign_cidr = "10.0.0.0/24"
tun_name = ""  # TUN disabled for simplicity
log_level = "debug"

[api_server]
listen_addr = "127.0.0.1:8080"
```

**Client** (`config.client.toml`):
```toml
server_addr = "127.0.0.1:4433"
server_name = "masque-vpn-server"
insecure_skip_verify = true  # For test certificates
```

### Production Deployment

For production use:
- Real certificates (not self-signed)
- TUN devices for traffic routing
- Firewall configuration for ports 4433 and 8080
- Monitoring via Prometheus/Grafana

## REST API

API provides endpoints for management and monitoring:

| Endpoint | Description |
|----------|-------------|
| `GET /health` | API health check |
| `GET /api/v1/status` | Server status |
| `GET /api/v1/clients` | Client list |
| `GET /api/v1/stats` | Server statistics |
| `GET /api/v1/config` | Server configuration |
| `GET /metrics` | Prometheus metrics |

Details in [API documentation](docs/api/api.ru.md).

## Technical Details

### Custom MASQUE Implementation

The project includes a custom MASQUE CONNECT-IP implementation:
- `common/masque_connectip.go` - MASQUE client
- `common/masque_proxy.go` - IP packet proxying functions
- `vpn_server/internal/server/masque_handler.go` - Server handler

### Modular Server Architecture

Server is split into specialized modules:
- `server.go` - Main server and initialization
- `api_server.go` - REST API server
- `masque_handler.go` - MASQUE request handler
- `packet_processor.go` - TUN device packet processing
- `metrics.go` - Prometheus metrics
- `tls_config.go` - TLS configuration

### Dependencies

- **Go 1.25** - Modern language version
- **QUIC**: [quic-go v0.57.1](https://github.com/quic-go/quic-go) - QUIC protocol
- **HTTP Framework**: [Gin](https://github.com/gin-gonic/gin) - for REST API
- **Metrics**: [Prometheus client](https://github.com/prometheus/client_golang) - metrics
- **Logging**: [Zap](https://github.com/uber-go/zap) - structured logging

### Security

- **Mutual TLS**: Client and server authenticate with certificates
- **Self-signed CA**: For testing and educational purposes
- **Client isolation**: Each client receives a unique IP address
- **Request validation**: MASQUE header and parameter validation

## Testing

### Run Tests

```bash
# Unit tests
cd common && go test -v
cd vpn_server && go test -v ./...
cd vpn_client && go test -v

# Integration tests
cd tests/integration && go test -v

# Load tests
cd tests/load && go test -v
```

### Automated Testing

```bash
# Local testing
./scripts/test-local.sh

# Docker testing
./scripts/test-docker.sh
```

## Development

### Project Structure

```
masque-vpn/
├── common/                    # Common libraries
│   ├── masque_connectip.go   # MASQUE client
│   ├── masque_proxy.go       # Packet proxying
│   └── errors.go             # Error system
├── vpn_server/               # Server
│   ├── internal/server/      # Server modules
│   └── main.go              # Entry point
├── vpn_client/              # Client
├── tests/                   # Tests
├── docs/                    # Documentation
└── cert/                    # Certificate generation
```

### Development Requirements

- Go 1.25 or later
- OpenSSL (for certificate generation)
- Docker (for containerization)
- Make (for automation)

## Troubleshooting

### Common Issues

**Build errors**:
- Ensure Go 1.25+ is installed
- Run `go mod tidy` to update dependencies

**Certificate errors**:
- Regenerate certificates if expired
- Check certificate paths in configuration

**Connection failures**:
- Check firewall settings for ports 4433 and 8080
- Verify server is listening on correct addresses
- Test API endpoints with curl

**Performance issues**:
- Monitor system resources (CPU, memory)
- Check server logs for errors
- Use Prometheus metrics for analysis

### Debug Mode

```bash
# Run with debug logging
./vpn-server -c config.server.local.toml  # already has log_level = "debug"

# Check API logs
curl http://127.0.0.1:8080/api/v1/logs
```

## Educational Use

### For Students and Researchers

This project is designed for studying:
- **Modern Network Protocols**: QUIC, HTTP/3, MASQUE
- **Go Programming**: Concurrency, networking, testing
- **System Design**: Modular architecture, API design, monitoring
- **Research Methodology**: Performance analysis, data collection, documentation

### Laboratory Assignments

1. **Basic Deployment** - build and run the system
2. **Performance Analysis** - measure VPN metrics
3. **Protocol Deep Dive** - study MASQUE implementation
4. **Network Conditions Testing** - emulate network problems

### Research Projects

- MASQUE performance optimization
- Load behavior analysis
- Security research
- Monitoring and observability systems

Details in [student guide](docs/development/student-guide.md).

## Contributing

This project is created for educational purposes. Contributions welcome for:
- Protocol implementation improvements
- Cross-platform compatibility
- Documentation enhancements
- Bug fixes
- New test scenarios

## Standards and RFCs

- **MASQUE CONNECT-IP**: [RFC 9484](https://datatracker.ietf.org/doc/html/rfc9484) - Proxying IP in HTTP
- **QUIC Transport**: [RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000) - QUIC transport
- **HTTP/3**: [RFC 9114](https://datatracker.ietf.org/doc/html/rfc9114) - HTTP/3 protocol

## License and Copyright

This project is distributed under the MIT License.

**Copyright (c) 2025 CloudBridge Research / 2GC Network Protocol Suite**  
**Copyright (c) 2025 Original Authors (iselt/masque-vpn)**

This project is based on the original work [iselt/masque-vpn](https://github.com/iselt/masque-vpn) and is developed by CloudBridge Research for educational and research purposes.

## Support

For questions and collaboration:
- Project repository: [cloudbridge-research/masque-vpn](https://github.com/cloudbridge-research/masque-vpn)
- Original repository: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- GitHub discussions and issues
- Educational institutions welcome to fork and adapt

## Other Languages

- [Русская документация](README.md) - Russian documentation
- [中文文档](README_zh.md) - Chinese documentation
