# Student Guide: Using masque-vpn for Research

This guide helps MPEI students use masque-vpn for research projects, diploma work, and laboratory assignments.

## Overview

masque-vpn implements the MASQUE CONNECT-IP protocol (RFC 9484) for tunneling IP packets over HTTP/3 (QUIC). This makes it ideal for studying modern network protocols and comparing VPN performance.

## Prerequisites

- Basic understanding of VPN concepts
- Familiarity with QUIC protocol (see course materials)
- Go programming language basics
- Linux/macOS command line experience

## Getting Started

### 1. Build and Deploy

```bash
# Clone repository
git clone --recurse-submodules https://github.com/twogc/masque-vpn.git
cd masque-vpn

# Build server and client
cd vpn_server && go build -o vpn-server
cd ../vpn_client && go build -o vpn-client
```

### 2. Configure Server

```bash
cd vpn_server
cp config.server.toml.example config.server.toml
# Edit config.server.toml as needed
```

Key configuration options:
- `listen_addr`: Server listening address (default: "0.0.0.0:4433")
- `assign_cidr`: IP range for clients (default: "10.0.0.0/24")
- `advertise_routes`: Routes to advertise (default: ["0.0.0.0/0"])

### 3. Generate Certificates

```bash
cd cert
sh gen_ca.sh
sh gen_server_keypair.sh
```

### 4. Start Server

```bash
cd ..
sudo ./vpn-server
```

### 5. Generate Client Configuration

Access web UI at http://localhost:8080 (admin/admin) and generate client configuration.

### 6. Start Client

```bash
cd ../vpn_client
# Copy client config from web UI to config.client.toml
sudo ./vpn-client
```

## Research Use Cases

### Performance Analysis

Compare MASQUE VPN with traditional VPN protocols:

1. **Latency measurement**: Use ping to measure RTT through VPN
2. **Throughput testing**: Use iperf3 to measure bandwidth
3. **Packet loss analysis**: Monitor packet loss under different network conditions

### Network Simulation

Test VPN behavior under various network conditions using tc/netem:

```bash
# Add 100ms latency
sudo tc qdisc add dev eth0 root netem delay 100ms

# Add 10% packet loss
sudo tc qdisc add dev eth0 root netem loss 10%

# Add bandwidth limitation
sudo tc qdisc add dev eth0 root tbf rate 1mbit burst 32kbit latency 400ms
```

### Metrics Collection

Once Phase 1 (Prometheus metrics) is implemented, you can:

1. Collect real-time performance data
2. Create graphs for presentations
3. Compare different configurations
4. Analyze protocol behavior

See [monitoring/prometheus-setup.md](../monitoring/prometheus-setup.md) for details.

## Laboratory Assignments

### Lab 1: Basic Deployment

**Objective**: Deploy and test masque-vpn

**Tasks**:
1. Build server and client
2. Configure certificates
3. Establish VPN connection
4. Test connectivity

**Duration**: 2-3 hours

### Lab 2: Performance Testing

**Objective**: Measure VPN performance

**Tasks**:
1. Measure baseline latency and throughput
2. Compare with direct connection
3. Test under different network conditions
4. Analyze results

**Duration**: 3-4 hours

### Lab 3: Protocol Analysis

**Objective**: Understand MASQUE protocol behavior

**Tasks**:
1. Capture QUIC traffic with Wireshark
2. Analyze packet structure
3. Identify MASQUE-specific features
4. Compare with HTTP/3 traffic

**Duration**: 4-5 hours

## Diploma Work Topics

### Topic 1: Performance Optimization

**Title**: "Optimizing MASQUE VPN Performance with BBRv3 and FEC"

**Scope**:
- Implement BBRv3 congestion control
- Add Forward Error Correction
- Conduct comparative analysis
- Publish results

**Duration**: 4-6 months

### Topic 2: Monitoring and Analytics

**Title**: "Monitoring and Analytics for MASQUE Protocol VPN"

**Scope**:
- Implement Prometheus metrics
- Create Grafana dashboards
- Develop analysis tools
- Write documentation

**Duration**: 3-4 months

### Topic 3: Security Analysis

**Title**: "Post-Quantum Cryptography in VPN Solutions"

**Scope**:
- Study PQC algorithms
- Implement hybrid cryptography
- Analyze performance impact
- Compare with classical crypto

**Duration**: 6-8 months

## Best Practices

### Testing

- Always test in isolated environment first
- Use network emulation for reproducible results
- Document all configuration changes
- Keep detailed logs

### Data Collection

- Collect baseline measurements before changes
- Run multiple test iterations for statistical significance
- Record all environmental factors
- Use consistent test methodology

### Documentation

- Document all experiments
- Include configuration files
- Save raw data for later analysis
- Create clear graphs and charts

## Troubleshooting

### Common Issues

**Certificate errors**:
- Verify CA and certificates are properly generated
- Check certificate paths in configuration
- Ensure certificates are not expired

**Permission issues**:
- TUN device creation requires root/administrator privileges
- Use `sudo` when starting server and client

**Connection failures**:
- Check firewall settings
- Verify server is listening on correct port
- Ensure client has correct server address

**Performance issues**:
- Check MTU settings
- Verify network conditions
- Monitor system resources

## Additional Resources

- [Architecture Overview](../architecture/overview.md)
- [Monitoring Setup](../monitoring/setup.md)
- [Development Guide](development/guide.md)
- [MASQUE RFC 9484](https://datatracker.ietf.org/doc/html/rfc9484)
- [QUIC RFC 9000](https://datatracker.ietf.org/doc/html/rfc9000)

## Support

For questions and collaboration:
- Check original repository: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- CloudBridge Research: [cloudbridge-research.ru](https://cloudbridge-research.ru)
- MPEI course materials: [openedu.mpei.ru/course/BAS_2](https://openedu.mpei.ru/course/BAS_2)
