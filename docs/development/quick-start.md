# Quick Start Guide

## Build and Run the Server

### 1. Build the executable

```bash
cd vpn_server
go build -o vpn-server .
```

### 2. Generate certificates (if not already done)

```bash
cd cert
sh gen_ca.sh
sh gen_server_keypair.sh
cd ..
```

### 3. Create configuration file

```bash
cp config.server.toml.example config.server.toml
```

Edit `config.server.toml` if needed (defaults should work for testing).

### 4. Run the server

```bash
./vpn-server
```

Or with custom config:

```bash
./vpn-server -c /path/to/config.server.toml
```

### 5. Verify it's running

- VPN server: listening on port 4433 (default)
- Web UI: http://localhost:8080 (admin/admin)
- Metrics: http://localhost:9090/metrics

## Quick Test

Check if metrics are working:

```bash
curl http://localhost:9090/metrics | head -20
```

You should see Prometheus metrics output.

For detailed instructions, see [Getting Started](getting-started.md).

