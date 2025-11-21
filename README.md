# masque-vpn

[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/iselt/masque-vpn)

A VPN implementation based on the MASQUE (CONNECT-IP) protocol using QUIC transport.

**⚠️ This project is in early development and is not ready for production use. It is intended for educational purposes and to demonstrate the MASQUE protocol.**

**This project includes a submodule `connect-ip-go`. Please clone the repository with `--recurse-submodules`:**

```bash
git clone --recurse-submodules https://github.com/twogc/masque-vpn.git
```

Or if you're using the original repository:

```bash
git clone --recurse-submodules https://github.com/iselt/masque-vpn.git
```

## Features

- **Modern Protocols**: Built on QUIC and MASQUE CONNECT-IP
- **Mutual TLS Authentication**: Certificate-based client-server authentication
- **Web Management UI**: Browser-based client management and configuration
- **Cross-Platform**: Supports Windows, Linux, and macOS
- **IP Pool Management**: Automatic client IP allocation and routing
- **Real-time Monitoring**: Live client connection status

## Architecture

The system consists of:
- **VPN Server**: Handles client connections and traffic routing
- **VPN Client**: Connects to server and routes local traffic
- **Web UI**: Management interface for certificates and clients
- **Certificate System**: PKI-based authentication using mutual TLS

## Quick Start

### 1. Build

```bash
cd vpn_client && go build
cd ../vpn_server && go build
cd ../admin_webui && npm install && npm run build
```

### 2. Certificate Setup

```bash
cd vpn_server/cert
# Generate CA certificate
sh gen_ca.sh
# Generate server certificate
sh gen_server_keypair.sh
```

### 3. Server Configuration

Copy and edit the server configuration:
```bash
cp vpn_server/config.server.toml.example vpn_server/config.server.toml
```

### 4. Start Server

```bash
cd vpn_server
./vpn-server
```

### 5. Web Management

- Access: `http://<server-ip>:8080/`
- Default credentials: `admin` / `admin`
- Generate client configurations through the web interface

### 6. Start Client

```bash
cd vpn_client
./vpn-client
```

## Configuration

### Server Configuration

Key configuration options in `config.server.toml`:

| Option | Description | Example |
|--------|-------------|---------|
| `listen_addr` | Server listening address | `"0.0.0.0:4433"` |
| `assign_cidr` | IP range for clients | `"10.0.0.0/24"` |
| `advertise_routes` | Routes to advertise | `["0.0.0.0/0"]` |
| `cert_file` | Server certificate path | `"cert/server.crt"` |
| `key_file` | Server private key path | `"cert/server.key"` |

### Client Configuration

Generated automatically via Web UI or manually configured:

| Option | Description |
|--------|-------------|
| `server_addr` | VPN server address |
| `server_name` | Server name for TLS |
| `ca_pem` | CA certificate (embedded) |
| `cert_pem` | Client certificate (embedded) |
| `key_pem` | Client private key (embedded) |

## Web Management Interface

The web interface provides:

- **Client Management**: Generate, download, and delete client configurations
- **Live Monitoring**: View connected clients and their IP assignments
- **Certificate Management**: Automated certificate generation and distribution
- **Configuration**: Server settings management

## Technical Details

### Dependencies

- **QUIC**: [quic-go](https://github.com/quic-go/quic-go) - QUIC protocol implementation
- **MASQUE**: [connect-ip-go](https://github.com/quic-go/connect-ip-go) - MASQUE CONNECT-IP protocol
- **Database**: SQLite for client and configuration storage
- **TUN**: Cross-platform TUN device management

### Security

- **Mutual TLS**: Both client and server authenticate using certificates
- **Certificate Authority**: Self-signed CA for certificate management
- **Unique Client IDs**: Each client has a unique identifier
- **IP Isolation**: Clients receive individual IP assignments

## Development

### Project Structure

```
masque-vpn/
├── common/           # Shared code and utilities
├── vpn_client/       # Client implementation
├── vpn_server/       # Server implementation
│   └── cert/         # Certificate generation scripts
├── admin_webui/      # Web UI assets
└── README.md
```

### Building from Source

Requirements:
- Go 1.24.2 or later
- OpenSSL (for certificate generation)

## Troubleshooting

### Common Issues

1. **Certificate Errors**: Ensure CA and certificates are properly generated
2. **Permission Issues**: TUN device creation requires administrator privileges
3. **Firewall**: Ensure server port (default 4433) is accessible
4. **MTU Issues**: Adjust MTU settings if experiencing connectivity problems


### macOS-Specific Issues

#### Technical Details

**Point-to-Point Interfaces**

TUN devices on macOS create a virtual point-to-point tunnel between two endpoints:
- **Server (local)**: gateway IP (e.g., 10.0.0.1)
- **Client (destination)**: next IP address (e.g., 10.0.0.2)

The `Next()` method from the `netip` package returns the next IP address:
- `10.0.0.1.Next()` → `10.0.0.2`
- `192.168.1.1.Next()` → `192.168.1.2`

**WireGuard TUN Offset**

WireGuard TUN on macOS uses an offset for packet header placement:
- **4 bytes** - minimum offset required for macOS
- **10 bytes** - used in `proxy.go` (VirtioNetHdrLen)

The offset indicates where packet data begins in the buffer. The library writes the header before this position:

```
Buffer: [header (4 bytes)][packet data (n bytes)]
                         ^
                         offset=4
```

**Platform Compatibility**

These fixes are macOS-specific and don't affect other platforms:
- **Linux**: uses `tun_linux.go`
- **Windows**: uses `tun_windows.go`
## Contributing

This project is for educational purposes. Contributions are welcome for:
- Protocol improvements
- Cross-platform compatibility
- Documentation enhancements
- Bug fixes

## References

- [MASQUE Protocol Specification](https://datatracker.ietf.org/doc/draft-ietf-masque-connect-ip/)
- [QUIC Protocol](https://datatracker.ietf.org/doc/rfc9000/)
- [quic-go Library](https://github.com/quic-go/quic-go)
- [connect-ip-go Library](https://github.com/quic-go/connect-ip-go)

### Standards and RFCs

- **MASQUE CONNECT-IP**: [RFC 9484](https://datatracker.ietf.org/doc/html/rfc9484) - Proxying IP in HTTP
- **MASQUE CONNECT-UDP**: [RFC 9298](https://datatracker.ietf.org/doc/html/rfc9298) - Proxying UDP in HTTP
- **QUIC Transport**: [RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000) - QUIC: A UDP-Based Multiplexed and Secure Transport
- **HTTP Datagrams**: [RFC 9297](https://datatracker.ietf.org/doc/html/rfc9297) - HTTP Datagrams and the Capsule Protocol

This project is built upon the following open-source libraries:

* [quic-go](https://github.com/quic-go/quic-go) - A QUIC implementation in Go
* [connect-ip-go](https://github.com/quic-go/connect-ip-go) - A Go implementation of the MASQUE CONNECT-IP protocol

## For MPEI Students and Researchers

This project is part of the educational materials for the National Research University "MPEI" (Moscow Power Engineering Institute) course on modern network protocols for autonomous systems (UAS - Unmanned Aerial Systems).

### Connection to MPEI Curriculum

This implementation demonstrates **MASQUE CONNECT-IP (RFC 9484)** - a modern protocol for tunneling IP packets over HTTP/3 (QUIC), which is covered in the National Research University "MPEI" course "Personnel for Autonomous Systems".

**Course Resources:**
- OpenEdu Course: [openedu.mpei.ru/course/BAS_2](https://openedu.mpei.ru/course/BAS_2)
- Federal Project: "Personnel for Autonomous Systems"

### Why MASQUE for Autonomous Systems?

MASQUE CONNECT-IP solves critical connectivity challenges for autonomous systems:

1. **Corporate Network Bypass**: Many corporate networks block UDP traffic. MASQUE tunnels IP packets over QUIC (UDP:443), making it look like HTTPS traffic and bypassing firewalls.

2. **Mobile Network Optimization**: QUIC's built-in features (0-RTT, connection migration, multiplexing) provide better performance than traditional VPN protocols in mobile scenarios.

3. **Standards-Based**: Unlike proprietary VPN solutions, MASQUE is an IETF standard (RFC 9484), ensuring interoperability and future compatibility.

### Research Topics for Students

This project can serve as a foundation for diploma work, research projects, or laboratory assignments:

#### 1. Performance Analysis
- Compare MASQUE VPN performance vs traditional VPN protocols (OpenVPN, WireGuard)
- Measure latency, throughput, and jitter under different network conditions
- Analyze QUIC connection migration impact on VPN stability

#### 2. Security Evaluation
- Evaluate mutual TLS authentication implementation
- Analyze certificate management and PKI security
- Study anti-replay mechanisms for 0-RTT connections

#### 3. Network Optimization
- Implement and test BBRv3 congestion control for MASQUE tunnels
- Optimize IP pool management and routing algorithms
- Study the impact of packet loss on VPN performance

#### 4. Integration with Autonomous Systems
- Integrate MASQUE VPN with ground control stations
- Implement handover prediction for mobile scenarios
- Design failover mechanisms for critical connections

#### 5. Protocol Extensions
- Implement additional MASQUE features (CONNECT-UDP for specific use cases)
- Add support for IPv6 tunneling
- Integrate with AI-based routing systems

### Laboratory Assignments

**Basic Level:**
1. Set up MASQUE VPN server and client
2. Configure certificate-based authentication
3. Test connectivity through corporate firewalls
4. Monitor connection metrics (latency, throughput)

**Advanced Level:**
1. Implement custom routing policies
2. Add performance monitoring and metrics collection
3. Integrate with network emulation tools (tc, netem)
4. Compare performance with traditional VPN solutions

**Research Level:**
1. Implement and test protocol optimizations
2. Design and evaluate new features
3. Publish results in academic conferences
4. Contribute improvements back to the project

### Getting Started for Research

1. **Fork this repository**: Create your own fork for experiments
2. **Read the code**: Understand the architecture in `vpn_server/` and `vpn_client/`
3. **Set up test environment**: Use network emulation to simulate various conditions
4. **Collect metrics**: Implement logging and monitoring for your research
5. **Document findings**: Write reports comparing different configurations

### Contact and Collaboration

For questions about using this project in National Research University "MPEI" research:
- Check the original repository: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- CloudBridge Research: [2gc.ru](https://2gc.ru)

### Related Projects

- [CloudBridge QUIC Test Suite](https://github.com/twogc/cloudbridge-relay-installer) - QUIC protocol testing and benchmarking
- [CloudBridge Research](https://cloudbridge-research.ru) - Research documentation and publications

## 中文文档

请参考 [README_zh.md](README_zh.md) 获取中文使用说明。
## Документация на других языках

Русская документация: [README_ru.md](README_ru.md)
