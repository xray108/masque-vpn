# MASQUE VPN Deployment Guide

This guide covers deployment options for the MASQUE VPN system with the implemented high-priority improvements.

## Quick Start with Docker Compose

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- At least 2GB RAM
- Network administrator privileges

### 1. Clone and Setup

```bash
git clone --recurse-submodules https://github.com/your-org/masque-vpn.git
cd masque-vpn
```

### 2. Generate Certificates

```bash
make certs
```

### 3. Configure Services

Copy and edit configuration files:

```bash
# Server configuration
cp vpn_server/config.server.toml.example vpn_server/config.server.toml
# Edit vpn_server/config.server.toml as needed

# Client configuration (for testing)
cp vpn_client/config.client.toml.example vpn_client/config.client.toml
# Edit vpn_client/config.client.toml as needed
```

### 4. Deploy with Docker Compose

```bash
# Build and start all services
make docker-build
make docker-run

# Or use docker-compose directly
docker-compose up -d
```

### 5. Access Services

- **Admin Web UI**: http://localhost:8080
- **Grafana Dashboard**: http://localhost:3001 (admin/admin)
- **Prometheus**: http://localhost:9091
- **Server Metrics**: http://localhost:9090/metrics
- **Client Metrics**: http://localhost:9092/metrics

## Production Deployment

### Environment Variables

Set these environment variables for production:

```bash
# Security
export ENVIRONMENT=production
export TLS_CERT_PATH=/path/to/certs
export DATABASE_PATH=/path/to/data

# Networking
export VPN_LISTEN_ADDR=0.0.0.0:4433
export VPN_ASSIGN_CIDR=10.0.0.0/16
export VPN_SERVER_NAME=vpn.yourdomain.com

# Monitoring
export METRICS_ENABLED=true
export LOG_LEVEL=info

# Database
export DB_PATH=/data/masque_admin.db
```

### Docker Production Setup

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  vpn-server:
    image: your-registry/masque-vpn-server:latest
    restart: always
    ports:
      - "4433:4433/udp"
      - "8080:8080/tcp"
    volumes:
      - /etc/masque-vpn/certs:/app/cert:ro
      - /etc/masque-vpn/config.toml:/app/config.server.toml:ro
      - /var/lib/masque-vpn:/app/data
    environment:
      - ENVIRONMENT=production
    cap_add:
      - NET_ADMIN
    sysctls:
      - net.ipv4.ip_forward=1
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
```

### Kubernetes Deployment

```yaml
# k8s/vpn-server-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: masque-vpn-server
  labels:
    app: masque-vpn-server
spec:
  replicas: 2
  selector:
    matchLabels:
      app: masque-vpn-server
  template:
    metadata:
      labels:
        app: masque-vpn-server
    spec:
      containers:
      - name: vpn-server
        image: your-registry/masque-vpn-server:latest
        ports:
        - containerPort: 4433
          protocol: UDP
        - containerPort: 8080
          protocol: TCP
        - containerPort: 9090
          protocol: TCP
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          capabilities:
            add:
            - NET_ADMIN
        volumeMounts:
        - name: config
          mountPath: /app/config.server.toml
          subPath: config.server.toml
        - name: certs
          mountPath: /app/cert
          readOnly: true
        - name: data
          mountPath: /app/data
      volumes:
      - name: config
        configMap:
          name: masque-vpn-config
      - name: certs
        secret:
          secretName: masque-vpn-certs
      - name: data
        persistentVolumeClaim:
          claimName: masque-vpn-data
---
apiVersion: v1
kind: Service
metadata:
  name: masque-vpn-server
spec:
  selector:
    app: masque-vpn-server
  ports:
  - name: quic
    port: 4433
    protocol: UDP
  - name: http
    port: 8080
    protocol: TCP
  - name: metrics
    port: 9090
    protocol: TCP
  type: LoadBalancer
```

## Monitoring Setup

### Prometheus Configuration

The included Prometheus configuration monitors:

- VPN server metrics (connections, throughput, errors)
- VPN client metrics (connection status, latency, errors)
- System metrics (CPU, memory, network)

### Grafana Dashboards

Pre-configured dashboards include:

1. **VPN Overview**: Connection status, active users, throughput
2. **Performance Metrics**: Latency, packet loss, FEC efficiency
3. **Error Analysis**: Error rates, failure types, recovery metrics
4. **System Health**: Resource usage, service status

### Alerting Rules

Configure alerts for:

```yaml
# monitoring/alerts.yml
groups:
- name: masque-vpn-alerts
  rules:
  - alert: VPNServerDown
    expr: up{job="masque-vpn-server"} == 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "VPN Server is down"
      
  - alert: HighErrorRate
    expr: rate(vpn_client_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      
  - alert: LowThroughput
    expr: rate(vpn_client_bytes_sent_total[5m]) < 1000
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Low throughput detected"
```

## Security Considerations

### Certificate Management

1. **Use proper CA**: Generate certificates with appropriate validity periods
2. **Certificate rotation**: Implement automated certificate renewal
3. **Secure storage**: Store private keys securely (HSM, Vault)

### Network Security

1. **Firewall rules**: Only expose necessary ports
2. **Rate limiting**: Implement connection rate limiting
3. **DDoS protection**: Use appropriate DDoS mitigation

### Container Security

1. **Non-root user**: Containers run as non-root user
2. **Read-only filesystem**: Mount filesystems as read-only where possible
3. **Security scanning**: Regular vulnerability scanning of images

## Performance Tuning

### Server Optimization

```toml
# config.server.toml
[performance]
max_connections = 1000
buffer_size = 65536
worker_threads = 4

[quic]
max_idle_timeout = "60s"
keep_alive_period = "30s"
max_bi_streams = 100
max_uni_streams = 100
```

### Client Optimization

```toml
# config.client.toml
[performance]
mtu = 1413
buffer_size = 32768

[fec]
enabled = true
redundancy = 0.1
block_size = 20
```

### System Tuning

```bash
# Increase UDP buffer sizes
echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf

# Enable IP forwarding
echo 'net.ipv4.ip_forward = 1' >> /etc/sysctl.conf
echo 'net.ipv6.conf.all.forwarding = 1' >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

## Troubleshooting

### Common Issues

1. **Permission denied creating TUN device**
   ```bash
   # Add NET_ADMIN capability
   docker run --cap-add=NET_ADMIN ...
   ```

2. **Certificate verification failed**
   ```bash
   # Check certificate validity
   openssl x509 -in cert/server.crt -text -noout
   ```

3. **High memory usage**
   ```bash
   # Monitor memory usage
   docker stats
   # Adjust buffer sizes in configuration
   ```

### Log Analysis

```bash
# View structured logs
docker-compose logs vpn-server | jq '.'

# Filter error logs
docker-compose logs vpn-server | jq 'select(.level=="error")'

# Monitor metrics
curl http://localhost:9090/metrics | grep vpn_
```

### Health Checks

```bash
# Check server health
curl http://localhost:8080/health

# Check client metrics
curl http://localhost:9092/health

# Verify QUIC connectivity
./tools/quic-test.sh vpn.yourdomain.com:4433
```

## Backup and Recovery

### Data Backup

```bash
# Backup certificates
tar -czf certs-backup.tar.gz vpn_server/cert/

# Backup database
sqlite3 vpn_server/masque_admin.db ".backup backup.db"

# Backup configuration
cp vpn_server/config.server.toml config-backup.toml
```

### Disaster Recovery

1. **Certificate restoration**: Restore certificates from secure backup
2. **Database recovery**: Restore SQLite database from backup
3. **Configuration sync**: Ensure configuration consistency
4. **Service restart**: Restart services in correct order

## Scaling

### Horizontal Scaling

1. **Load balancer**: Use UDP load balancer for QUIC traffic
2. **Shared state**: Use external database for client state
3. **Service discovery**: Implement service discovery for client routing

### Vertical Scaling

1. **Resource limits**: Adjust CPU and memory limits
2. **Connection limits**: Increase maximum connection limits
3. **Buffer sizes**: Optimize buffer sizes for throughput

This deployment guide provides a comprehensive approach to deploying the enhanced MASQUE VPN system with proper monitoring, security, and scalability considerations.