# Getting Started

This guide explains how to build and run the masque-vpn server for educational and research purposes.

## Prerequisites

- Go 1.25 or later
- OpenSSL (for certificate generation)
- Linux, macOS, or Windows
- Administrative privileges (for TUN device creation)

## Project Structure

For developers interested in the codebase:

- `vpn_server/`: Server entry point (`main.go`) and modular server implementation.
- `vpn_server/internal/server/`: Core server logic split into modules:
  - `server.go`: Main server and initialization
  - `api_server.go`: REST API server for management
  - `masque_handler.go`: MASQUE CONNECT-IP request handler
  - `packet_processor.go`: TUN device packet processing
  - `metrics.go`: Prometheus metrics
  - `tls_config.go`: TLS configuration
- `vpn_client/`: Client implementation with MASQUE CONNECT-IP support.
- `common/`: Shared utilities and custom MASQUE implementation:
  - `masque_connectip.go`: Custom MASQUE CONNECT-IP client
  - `masque_proxy.go`: IP packet tunneling functions
  - `errors.go`: Centralized error handling system

## Step 1: Build the Server

```bash
cd vpn_server
go build -o vpn-server .
```

This will create a `vpn-server` executable in the `vpn_server` directory.

## Step 2: Build the Client

```bash
cd vpn_client
go build -o vpn-client .
```

## Step 3: Generate Certificates

Before starting the server, you need to generate TLS certificates:

```bash
cd cert

# For Linux/macOS
./generate-test-certs.sh

# For Windows
powershell -ExecutionPolicy Bypass -File generate-certs.ps1
```

This will create the following files in the `cert/` directory:
- `ca.crt` - CA certificate
- `ca.key` - CA private key
- `server.crt` - Server certificate
- `server.key` - Server private key
- `client.crt` - Client certificate
- `client.key` - Client private key

## Step 4: Configure the Server

Use the provided local configuration:

```bash
cd vpn_server
# Use config.server.local.toml for local testing
```

Key configuration settings in `config.server.local.toml`:

- `listen_addr = "127.0.0.1:4433"` - Server listening address
- `assign_cidr = "10.0.0.0/24"` - IP range for VPN clients
- `tun_name = ""` - TUN device disabled for local testing
- `api_server.listen_addr = "127.0.0.1:8080"` - API server address

### API Server Configuration

The API server is configured in the `[api_server]` section:

```toml
[api_server]
listen_addr = "127.0.0.1:8080"
static_dir = "../admin_webui/dist"
database_path = "masque_admin.db"
```

### Metrics Configuration

Metrics are configured in the `[metrics]` section:

```toml
[metrics]
enabled = true
listen_addr = "127.0.0.1:9090"
```

## Step 5: Start the Server

### Using the local configuration:

```bash
cd vpn_server
./vpn-server -c config.server.local.toml
```

## Step 6: Verify the Server is Running

The server will start several services:

1. **MASQUE VPN Server** - Listens on `127.0.0.1:4433`
2. **REST API Server** - Available at `http://127.0.0.1:8080`
3. **Prometheus Metrics** - Available at `http://127.0.0.1:8080/metrics`

### Check Server Logs

You should see output like:

```
2025-12-21T03:57:31.297+0300    INFO    Starting MASQUE VPN Server
2025/12/21 03:57:31 TUN device disabled (empty tun_name)
2025/12/21 03:57:31 MASQUE VPN Server listening on 127.0.0.1:4433
2025/12/21 03:57:31 API Server will start on 127.0.0.1:8080
```

### Verify API Endpoints

Check health status:
```bash
curl http://127.0.0.1:8080/health
```

Check server status:
```bash
curl http://127.0.0.1:8080/api/v1/status
```

Check metrics:
```bash
curl http://127.0.0.1:8080/metrics
```

## Step 7: Configure and Start the Client

Configure the client for local testing:

```bash
cd vpn_client
# Edit config.client.toml to use server_addr = "127.0.0.1:4433"
```

Start the client:

```bash
./vpn-client -c config.client.toml
```

## Testing the Connection

### Basic Connection Test

Run the included connection test:

```bash
go run test_masque_connection.go
```

This will test basic MASQUE protocol functionality without requiring TUN devices.

### API Testing

Check connected clients:
```bash
curl http://127.0.0.1:8080/api/v1/clients
```

View server statistics:
```bash
curl http://127.0.0.1:8080/api/v1/stats
```

## Troubleshooting

### Permission Errors

If you get permission errors when creating TUN devices:

```bash
# Linux/macOS
sudo ./vpn-server -c config.server.local.toml
sudo ./vpn-client -c config.client.toml

# Windows (run as Administrator)
.\vpn-server.exe -c config.server.local.toml
.\vpn-client.exe -c config.client.toml
```

### Certificate Errors

Make sure certificates are generated and paths are correct:

```bash
ls -la cert/
# Should show: ca.crt, ca.key, server.crt, server.key, client.crt, client.key
```

### Port Already in Use

If you get "address already in use" error:

1. Check if another instance is running: `ps aux | grep vpn-server`
2. Change ports in configuration files
3. On Linux: `sudo netstat -tulpn | grep 4433`

### Connection Issues

- Verify firewall settings allow traffic on configured ports
- Check that client uses correct server address
- Ensure certificates are valid and not expired
- Review server logs for detailed error messages

## Running Tests

### Unit Tests

```bash
# Test common package
cd common && go test -v

# Test server components
cd vpn_server && go test -v ./...

# Test client
cd vpn_client && go test -v
```

### Integration Tests

```bash
# Run integration tests
cd tests/integration && go test -v

# Run load tests
cd tests/load && go test -v
```

### Automated Testing

```bash
# Local testing script
./scripts/test-local.sh

# Docker testing script
./scripts/test-docker.sh
```

## Next Steps

After the server is running:

1. Test basic MASQUE protocol functionality with `test_masque_connection.go`
2. Experiment with different network conditions using `tc` (Linux)
3. Monitor performance using Prometheus metrics
4. Study the custom MASQUE implementation in `common/masque_connectip.go`
5. Explore the modular server architecture in `vpn_server/internal/server/`

## Educational Use

This implementation is designed for educational and research purposes:

- **Protocol Study**: Learn MASQUE CONNECT-IP implementation details
- **Performance Analysis**: Use built-in metrics for performance studies
- **Network Research**: Test behavior under various network conditions
- **Code Analysis**: Study modern Go networking patterns and QUIC usage

See [Student Guide](../development/student-guide.ru.md) for detailed research scenarios and lab exercises.

