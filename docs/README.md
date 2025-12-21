# masque-vpn Documentation

This directory contains comprehensive documentation for the masque-vpn project - an educational implementation of the MASQUE CONNECT-IP protocol for research and learning purposes.

## Structure

- `architecture/` - Architecture and design documents
- `api/` - REST API documentation and reference
- `development/` - Development guides and getting started information
- `monitoring/` - Monitoring, metrics, and observability guides

## Quick Links

### Getting Started
- [Getting Started Guide](development/getting-started.md) - Build and run the server
- [Student Guide (Russian)](development/student-guide.ru.md) - Educational use and lab exercises

### Architecture
- [Architecture Overview (Russian)](architecture/architecture.ru.md) - System design and components

### API Documentation
- [REST API Reference (Russian)](api/api.ru.md) - Complete API documentation

### Monitoring
- [Monitoring Setup](monitoring/setup.md) - Prometheus and Grafana configuration
- [Metrics Reference](monitoring/metrics.md) - Complete metrics documentation

## Project Overview

masque-vpn is an educational implementation of:
- **MASQUE CONNECT-IP Protocol** (RFC 9484)
- **QUIC Transport** (RFC 9000) 
- **HTTP/3** (RFC 9114)
- **Custom VPN Solution** for research and learning

### Key Features

- Custom MASQUE CONNECT-IP implementation
- Modular server architecture
- REST API for management and monitoring
- Comprehensive Prometheus metrics
- Structured logging with zap
- Full test coverage (unit, integration, load)
- Docker containerization support
- Educational documentation and lab exercises

### Educational Focus

This project is specifically designed for:
- University courses on network protocols
- Research projects on VPN technologies
- Performance analysis and comparison studies
- Protocol implementation learning
- Modern Go networking patterns

## Quick Start

1. **Build the project**:
   ```bash
   cd vpn_server && go build -o vpn-server .
   cd ../vpn_client && go build -o vpn-client .
   ```

2. **Generate certificates**:
   ```bash
   cd cert && ./generate-test-certs.sh
   ```

3. **Start the server**:
   ```bash
   cd vpn_server && ./vpn-server -c config.server.local.toml
   ```

4. **Test the connection**:
   ```bash
   go run test_masque_connection.go
   ```

5. **Check API endpoints**:
   ```bash
   curl http://127.0.0.1:8080/health
   curl http://127.0.0.1:8080/api/v1/status
   ```

## Documentation Updates

This documentation has been updated to reflect:
- Go 1.25 compatibility
- Custom MASQUE implementation (no external dependencies)
- Modular server architecture
- REST API server integration
- Comprehensive testing framework
- Updated metrics and monitoring
- Educational focus and lab exercises

## Contributing

For educational institutions and researchers:
- Fork the repository for your specific research needs
- Adapt configurations for your network environment
- Extend metrics for your specific analysis requirements
- Create additional lab exercises for your curriculum

## Support

- Original repository: [iselt/masque-vpn](https://github.com/iselt/masque-vpn)
- Educational materials and courses
- Research collaboration opportunities

