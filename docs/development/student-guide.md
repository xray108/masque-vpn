# Student Guide: Using masque-vpn for Research

This guide helps university students use masque-vpn for research projects, diploma work, and laboratory assignments.

## Overview

masque-vpn is an educational implementation of the MASQUE CONNECT-IP protocol (RFC 9484) for tunneling IP packets over HTTP/3 (QUIC). This makes it ideal for studying modern network protocols and comparing VPN performance in academic research.

## Prerequisites

- Basic understanding of VPN concepts
- Familiarity with QUIC protocol (see course materials)
- Go programming language basics (Go 1.25+)
- Linux/macOS/Windows command line experience
- Administrative privileges for TUN device creation

## Getting Started

### 1. Build and Deploy

```bash
# Clone repository
git clone https://github.com/iselt/masque-vpn.git
cd masque-vpn

# Build server and client
cd vpn_server && go build -o vpn-server .
cd ../vpn_client && go build -o vpn-client .
```

### 2. Generate Certificates

```bash
cd cert

# For Linux/macOS
./generate-test-certs.sh

# For Windows
powershell -ExecutionPolicy Bypass -File generate-certs.ps1
```

### 3. Configure Server

Use the provided local configuration for testing:

```bash
cd vpn_server
# Use config.server.local.toml for local testing
```

Key configuration options:
- `listen_addr = "127.0.0.1:4433"` - Server listening address
- `assign_cidr = "10.0.0.0/24"` - IP range for clients
- `tun_name = ""` - TUN device disabled for local testing
- `api_server.listen_addr = "127.0.0.1:8080"` - API server address

### 4. Start Server

```bash
# Local testing (no TUN device required)
./vpn-server -c config.server.local.toml

# With TUN device (requires admin privileges)
sudo ./vpn-server -c config.server.toml
```

### 5. Configure and Start Client

```bash
cd vpn_client
# Edit config.client.toml to use server_addr = "127.0.0.1:4433"
./vpn-client -c config.client.toml
```

### 6. Test Basic Functionality

```bash
# Test MASQUE protocol functionality
go run test_masque_connection.go

# Check API endpoints
curl http://127.0.0.1:8080/health
curl http://127.0.0.1:8080/api/v1/status
```

## Research Use Cases

### Protocol Analysis

Study the MASQUE CONNECT-IP implementation:

1. **Custom Implementation**: Analyze the custom MASQUE implementation in `common/masque_connectip.go`
2. **Packet Flow**: Study packet processing in `vpn_server/internal/server/masque_handler.go`
3. **QUIC Integration**: Examine QUIC stream usage for IP packet tunneling
4. **Error Handling**: Review centralized error handling in `common/errors.go`

### Performance Analysis

Compare MASQUE VPN with traditional VPN protocols:

1. **Latency measurement**: Use ping to measure RTT through VPN
2. **Throughput testing**: Use iperf3 to measure bandwidth
3. **Connection establishment**: Measure QUIC handshake time
4. **Resource usage**: Monitor CPU and memory consumption

### Network Simulation

Test VPN behavior under various network conditions using tc/netem (Linux):

```bash
# Add 100ms latency
sudo tc qdisc add dev eth0 root netem delay 100ms

# Add 10% packet loss
sudo tc qdisc add dev eth0 root netem loss 10%

# Limit bandwidth to 1 Mbps
sudo tc qdisc add dev eth0 root tbf rate 1mbit burst 32kbit latency 400ms
```

### Metrics Collection

Use built-in Prometheus metrics for analysis:

```bash
# View available metrics
curl http://127.0.0.1:8080/metrics

# Monitor active connections
curl http://127.0.0.1:8080/api/v1/status

# View client statistics
curl http://127.0.0.1:8080/api/v1/clients
```

See [monitoring/setup.md](../monitoring/setup.md) for Grafana dashboard setup.

## Laboratory Assignments

### Lab 1: Basic Deployment and Testing

**Objective**: Deploy and test masque-vpn locally

**Tasks**:
1. Build server and client
2. Generate test certificates
3. Start server in local mode (no TUN)
4. Connect client and test basic functionality
5. Analyze API endpoints

**Duration**: 2-3 hours

### Lab 2: Performance Analysis

**Objective**: Measure and analyze VPN performance

**Tasks**:
1. Set up monitoring with Prometheus metrics
2. Measure connection establishment time
3. Test packet processing latency
4. Compare with direct connection baseline
5. Create performance report

**Duration**: 3-4 hours

### Lab 3: Protocol Deep Dive

**Objective**: Understand MASQUE protocol implementation

**Tasks**:
1. Study custom MASQUE implementation code
2. Capture QUIC traffic with Wireshark
3. Analyze MASQUE CONNECT-IP requests
4. Trace packet flow through the system
5. Document protocol behavior

**Duration**: 4-5 hours

### Lab 4: Network Conditions Testing

**Objective**: Test VPN behavior under adverse conditions

**Tasks**:
1. Set up network emulation (tc/netem)
2. Test with various latency levels (10ms, 100ms, 500ms)
3. Test with packet loss (1%, 5%, 10%)
4. Test with bandwidth limitations
5. Analyze performance degradation patterns

**Duration**: 3-4 hours

## Research Project Topics

### Topic 1: MASQUE Protocol Optimization

**Title**: "Performance Optimization of MASQUE CONNECT-IP Implementation"

**Scope**:
- Analyze current implementation bottlenecks
- Implement performance improvements
- Compare with other VPN protocols
- Measure and document improvements

**Duration**: 4-6 months

### Topic 2: Network Resilience Study

**Title**: "MASQUE VPN Behavior Under Network Stress Conditions"

**Scope**:
- Systematic testing under various network conditions
- Analysis of connection recovery mechanisms
- Comparison with traditional VPN protocols
- Development of resilience metrics

**Duration**: 3-4 months

### Topic 3: Security Analysis

**Title**: "Security Analysis of MASQUE CONNECT-IP Protocol Implementation"

**Scope**:
- Code security audit
- Protocol security analysis
- Penetration testing
- Security recommendations

**Duration**: 4-5 months

### Topic 4: Monitoring and Observability

**Title**: "Comprehensive Monitoring System for MASQUE VPN"

**Scope**:
- Extend Prometheus metrics
- Develop custom dashboards
- Implement alerting systems
- Create operational runbooks

**Duration**: 3-4 months

## Best Practices

### Testing Methodology

- Always test in isolated environment first
- Use consistent test configurations
- Document all environmental factors
- Run multiple iterations for statistical significance
- Keep detailed logs of all experiments

### Data Collection

- Collect baseline measurements before changes
- Use automated data collection where possible
- Store raw data for later analysis
- Document data collection methodology
- Validate data quality

### Documentation

- Document all experiments and configurations
- Include code snippets and examples
- Create clear visualizations
- Maintain version control for configurations
- Write reproducible procedures

## Code Analysis Guide

### Key Components to Study

1. **MASQUE Client** (`common/masque_connectip.go`):
   - Connection establishment
   - Packet encapsulation/decapsulation
   - Error handling

2. **Server Handler** (`vpn_server/internal/server/masque_handler.go`):
   - Request processing
   - Client session management
   - Packet routing

3. **API Server** (`vpn_server/internal/server/api_server.go`):
   - REST endpoint implementation
   - Client management
   - Statistics collection

4. **Testing Framework** (`tests/`):
   - Unit test patterns
   - Integration test setup
   - Load testing methodology

## Troubleshooting

### Common Issues

**Build errors**:
- Ensure Go 1.25+ is installed
- Check module dependencies with `go mod tidy`
- Verify GOPATH and GOROOT settings

**Certificate errors**:
- Regenerate certificates if expired
- Check certificate paths in configuration
- Verify CA certificate is properly installed

**Connection failures**:
- Check firewall settings for ports 4433 and 8080
- Verify server is listening on correct addresses
- Test with `curl` commands for API endpoints

**Performance issues**:
- Monitor system resources (CPU, memory)
- Check network interface statistics
- Review server logs for errors
- Verify MTU settings

### Debug Mode

Enable debug logging for detailed information:

```bash
# Server debug mode
./vpn-server -c config.server.local.toml  # Already has log_level = "debug"

# Check API server logs
curl http://127.0.0.1:8080/api/v1/logs
```

## Additional Resources

### Documentation
- [Architecture Overview](../architecture/architecture.ru.md)
- [API Reference](../api/api.ru.md)
- [Monitoring Setup](../monitoring/setup.md)
- [Getting Started Guide](getting-started.md)

### Standards and RFCs
- [MASQUE RFC 9484](https://datatracker.ietf.org/doc/html/rfc9484)
- [QUIC RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000)
- [HTTP/3 RFC 9114](https://datatracker.ietf.org/doc/html/rfc9114)

### Tools and Libraries
- [quic-go library](https://github.com/quic-go/quic-go)
- [Prometheus monitoring](https://prometheus.io/)
- [Wireshark for protocol analysis](https://www.wireshark.org/)

## Educational Value

This project provides hands-on experience with:

- **Modern Network Protocols**: QUIC, HTTP/3, MASQUE
- **Go Programming**: Concurrent programming, networking, testing
- **System Design**: Modular architecture, API design, monitoring
- **Research Methodology**: Performance analysis, data collection, documentation
- **DevOps Practices**: Containerization, monitoring, automation

## Support and Collaboration

For academic use and research collaboration:
- Original repository: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- Issues and discussions on GitHub
- Educational institutions welcome to fork and adapt
- Research papers and publications encouraged
