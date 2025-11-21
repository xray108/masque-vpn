# Getting Started

This guide explains how to build and run the masque-vpn server.

## Prerequisites

- Go 1.24.2 or later
- OpenSSL (for certificate generation)
- Linux, macOS, or Windows

## Step 1: Build the Server

```bash
cd vpn_server
go build -o vpn-server .
```

This will create a `vpn-server` executable in the `vpn_server` directory.

## Step 2: Generate Certificates

Before starting the server, you need to generate TLS certificates:

```bash
cd vpn_server/cert

# Generate CA certificate
sh gen_ca.sh

# Generate server certificate
sh gen_server_keypair.sh
```

This will create the following files in the `cert/` directory:
- `ca.crt` - CA certificate
- `ca.key` - CA private key
- `server.crt` - Server certificate
- `server.key` - Server private key

## Step 3: Configure the Server

Copy the example configuration file:

```bash
cd vpn_server
cp config.server.toml.example config.server.toml
```

Edit `config.server.toml` and adjust the following settings:

- `listen_addr` - Server listening address (default: `0.0.0.0:4433`)
- `server_name` - Server name for TLS verification
- `assign_cidr` - IP range for VPN clients (default: `10.99.0.0/24`)
- `advertise_routes` - Routes to advertise to clients

### Metrics Configuration

The metrics section is already configured in the example:

```toml
[metrics]
enabled = true
listen_addr = "0.0.0.0:9090"
```

If you want to disable metrics, set `enabled = false`.

## Step 4: Start the Server

### Using the default configuration file:

```bash
cd vpn_server
./vpn-server
```

### Using a custom configuration file:

```bash
cd vpn_server
./vpn-server -c /path/to/config.server.toml
```

## Step 5: Verify the Server is Running

The server will start several services:

1. **VPN Server** - Listens on the address specified in `listen_addr` (default: `0.0.0.0:4433`)
2. **Web Management UI** - Available at `http://localhost:8080` (default credentials: `admin` / `admin`)
3. **Prometheus Metrics** - Available at `http://localhost:9090/metrics` (if enabled)

### Check Server Logs

You should see output like:

```
Starting VPN Server...
Listen Address: 0.0.0.0:4433
VPN Network: 10.99.0.0/24
Gateway IP: 10.99.0.1
Metrics enabled, will be available at 0.0.0.0:9090/metrics
QUIC Listener started on 0.0.0.0:4433
Starting HTTP/3 server...
```

### Verify Metrics Endpoint

If metrics are enabled, you can check them:

```bash
curl http://localhost:9090/metrics
```

You should see Prometheus metrics in text format.

## Troubleshooting

### Permission Errors (Linux)

If you get permission errors when creating the TUN device, run with sudo:

```bash
sudo ./vpn-server
```

### Certificate Errors

Make sure certificates are generated and paths in `config.server.toml` are correct:

```bash
ls -la vpn_server/cert/
# Should show: ca.crt, ca.key, server.crt, server.key
```

### Port Already in Use

If you get "address already in use" error:

1. Check if another instance is running: `ps aux | grep vpn-server`
2. Change the port in `config.server.toml`
3. On Linux, check if port is in use: `sudo netstat -tulpn | grep 4433`

### Database Errors

The server creates a SQLite database file. Make sure the directory is writable:

```bash
touch vpn_server/masque_admin.db
chmod 644 vpn_server/masque_admin.db
```

## Next Steps

After the server is running:

1. Access the Web UI at `http://localhost:8080`
2. Log in with default credentials (`admin` / `admin`)
3. Generate client certificates through the web interface
4. Configure Prometheus to scrape metrics from `http://localhost:9090/metrics`
5. See [Monitoring Setup](../monitoring/setup.md) for Grafana dashboard configuration

## Running in Production

For production deployment:

1. Use proper TLS certificates (not self-signed)
2. Change default admin password
3. Configure firewall rules
4. Set up proper logging
5. Configure monitoring and alerting
6. Use systemd or similar for service management

See [Deployment Guide](deployment.md) for more details.

