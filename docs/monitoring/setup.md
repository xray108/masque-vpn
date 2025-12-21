# Monitoring Setup

This guide explains how to set up monitoring for masque-vpn using Prometheus and Grafana.

## Prerequisites

- Prometheus server
- Grafana server  
- masque-vpn server with metrics enabled

## Architecture

The monitoring setup consists of:

1. **masque-vpn server** - Exposes metrics on `/metrics` endpoint
2. **Prometheus** - Scrapes and stores metrics
3. **Grafana** - Visualizes metrics with dashboards

## Configuration

### Enable Metrics in masque-vpn

Metrics are enabled by default in the server configuration:

```toml
[metrics]
enabled = true
listen_addr = "127.0.0.1:9090"
```

Note: In the current implementation, metrics are served on the same port as the API server (8080) at the `/metrics` endpoint.

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'masque-vpn'
    static_configs:
      - targets: ['localhost:8080']  # API server port
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s
    
  # If running multiple servers
  - job_name: 'masque-vpn-cluster'
    static_configs:
      - targets: 
        - 'server1:8080'
        - 'server2:8080'
        - 'server3:8080'
    metrics_path: '/metrics'
```

### Docker Compose Setup

For easy deployment, use the provided `docker-compose.yml`:

```yaml
version: '3.8'
services:
  masque-vpn-server:
    build: ./vpn_server
    ports:
      - "4433:4433"  # MASQUE VPN
      - "8080:8080"  # API + Metrics
    volumes:
      - ./cert:/app/cert:ro
      
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      
  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
```

### Grafana Dashboard

#### Import Dashboard

1. Open Grafana at `http://localhost:3000`
2. Login with admin/admin
3. Go to "+" → Import
4. Upload `grafana/masque-vpn-dashboard.json`

#### Key Panels

The dashboard includes:

- **Connection Overview**: Active connections, total connections
- **Traffic Analysis**: Bytes/packets sent and received
- **Performance Metrics**: Packet processing latency, MASQUE request duration
- **Resource Utilization**: IP pool usage, TUN interface status
- **Error Monitoring**: Error rates by type, packet drops
- **MASQUE Protocol**: Request success rates, QUIC stream statistics

#### Custom Queries

Example queries for custom panels:

**Active Connections**:
```promql
masque_vpn_active_connections
```

**Packet Processing Latency (95th percentile)**:
```promql
histogram_quantile(0.95, rate(masque_vpn_packet_processing_duration_seconds_bucket[5m]))
```

**IP Pool Utilization**:
```promql
(masque_vpn_ip_pool_used / masque_vpn_ip_pool_total) * 100
```

**Error Rate**:
```promql
rate(masque_vpn_errors_total[5m])
```

## Metrics Endpoint

The metrics endpoint is available at:
```
http://localhost:8080/metrics
```

### Testing Metrics

Verify metrics are working:

```bash
# Check if metrics endpoint responds
curl http://localhost:8080/metrics

# Check specific metric
curl http://localhost:8080/metrics | grep masque_vpn_active_connections
```

### Sample Metrics Output

```text
# HELP masque_vpn_active_connections Current number of active connections
# TYPE masque_vpn_active_connections gauge
masque_vpn_active_connections 3

# HELP masque_vpn_total_connections Total connections since startup
# TYPE masque_vpn_total_connections counter
masque_vpn_total_connections 15

# HELP masque_vpn_ip_pool_total Total IP addresses in pool
# TYPE masque_vpn_ip_pool_total gauge
masque_vpn_ip_pool_total 254
```

## Alerting

### Prometheus Alerting Rules

Create `alerts.yml`:

```yaml
groups:
- name: masque-vpn.rules
  rules:
  - alert: MASQUEVPNHighErrorRate
    expr: rate(masque_vpn_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate in MASQUE VPN server"
      description: "Error rate is {{ $value }} errors per second"
      
  - alert: MASQUEVPNIPPoolExhaustion
    expr: (masque_vpn_ip_pool_available / masque_vpn_ip_pool_total) < 0.1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "MASQUE VPN IP pool nearly exhausted"
      description: "Only {{ $value }}% of IP addresses available"
      
  - alert: MASQUEVPNHighLatency
    expr: histogram_quantile(0.95, rate(masque_vpn_packet_processing_duration_seconds_bucket[5m])) > 0.1
    for: 3m
    labels:
      severity: warning
    annotations:
      summary: "High packet processing latency"
      description: "95th percentile latency is {{ $value }}s"
```

### Grafana Alerts

Configure alerts in Grafana:

1. Go to Alerting → Alert Rules
2. Create new rule
3. Set query and conditions
4. Configure notification channels (Slack, email, etc.)

## Troubleshooting

### Metrics Not Available

1. Check if metrics are enabled in server config
2. Verify server is running and accessible
3. Check firewall rules for port 8080
4. Review server logs for errors

### Prometheus Connection Issues

1. Verify target configuration in `prometheus.yml`
2. Check Prometheus logs: `docker logs prometheus`
3. Verify network connectivity: `curl http://server:8080/metrics`

### Grafana Dashboard Issues

1. Check data source configuration
2. Verify Prometheus is collecting metrics
3. Check query syntax in panels
4. Review Grafana logs for errors

## Performance Considerations

### Metrics Collection Impact

- Metrics collection has minimal performance impact
- Scrape interval of 10-15 seconds is recommended
- Avoid very frequent scraping (< 5 seconds)

### Storage Requirements

- Prometheus storage grows with number of metrics and retention period
- Default retention is 15 days
- Consider using remote storage for long-term retention

### High Availability

For production deployments:

1. Run multiple Prometheus instances
2. Use Prometheus federation
3. Configure Grafana with multiple data sources
4. Set up alertmanager clustering

## Educational Use

For research and educational purposes:

- Monitor protocol behavior under different conditions
- Analyze performance characteristics
- Study the impact of network conditions on VPN performance
- Compare metrics before and after configuration changes

See [Metrics Reference](metrics.md) for detailed metric descriptions.

