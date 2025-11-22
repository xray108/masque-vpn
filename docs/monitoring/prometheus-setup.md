# Prometheus Setup Guide

This guide explains how to set up Prometheus monitoring for masque-vpn.

## Prerequisites

- masque-vpn server with metrics enabled
- Prometheus installed
- Basic understanding of PromQL

## Installation

### macOS

```bash
brew install prometheus
```

### Linux (Ubuntu/Debian)

```bash
sudo apt-get update
sudo apt-get install prometheus
```

### Linux (Manual)

```bash
wget https://github.com/prometheus/prometheus/releases/download/v2.48.0/prometheus-2.48.0.linux-amd64.tar.gz
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

## Configuration

### 1. Enable Metrics in masque-vpn

Edit `vpn_server/config.server.toml`:

```toml
metrics_enabled = true
metrics_addr = ":9090"
```

### 2. Create Prometheus Configuration

Create `prometheus.yml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'masque-vpn'
    static_configs:
      - targets: ['localhost:9090']
        labels:
          instance: 'vpn-server-1'
```

For remote server, replace `localhost` with server IP address.

### 3. Start Prometheus

```bash
prometheus --config.file=prometheus.yml
```

### 4. Verify Setup

Open http://localhost:9090 in browser.

Check targets: Status → Targets

Expected: `masque-vpn` target shows as "UP"

## Basic Queries

### Connection Metrics

```promql
# Current active connections
masque_vpn_active_connections_total

# Total connections over time
rate(masque_vpn_total_connections_total[5m])

# Connection duration histogram
histogram_quantile(0.95, masque_vpn_connection_duration_seconds_bucket)
```

### Traffic Metrics

```promql
# Bytes sent per second (all clients)
rate(masque_vpn_bytes_sent_total[1m])

# Bytes received per second (specific client)
rate(masque_vpn_bytes_received_total{client_id="client-123"}[1m])

# Total traffic (sent + received)
sum(rate(masque_vpn_bytes_sent_total[5m])) + sum(rate(masque_vpn_bytes_received_total[5m]))
```

### Performance Metrics

```promql
# Average RTT per client
avg(masque_vpn_rtt_milliseconds) by (client_id)

# 95th percentile RTT
histogram_quantile(0.95, masque_vpn_rtt_milliseconds_bucket)

# Current throughput
masque_vpn_throughput_mbps

# Packet loss percentage
masque_vpn_packet_loss_percent
```

### Resource Metrics

```promql
# IP pool utilization percentage
(masque_vpn_ip_pool_allocated / masque_vpn_ip_pool_total) * 100

# Available IPs
masque_vpn_ip_pool_available
```

## Advanced Configuration

### Long-term Storage

By default, Prometheus stores data for 15 days. For research projects, you may want longer retention:

```bash
prometheus --config.file=prometheus.yml --storage.tsdb.retention.time=90d
```

### Recording Rules

Create `rules.yml` for pre-computed queries:

```yaml
groups:
  - name: masque_vpn
    interval: 30s
    rules:
      - record: masque_vpn:throughput_mbps:rate5m
        expr: rate(masque_vpn_bytes_sent_total[5m]) * 8 / 1000000
      
      - record: masque_vpn:rtt_p95:5m
        expr: histogram_quantile(0.95, rate(masque_vpn_rtt_milliseconds_bucket[5m]))
```

Update `prometheus.yml`:

```yaml
rule_files:
  - "rules.yml"
```

### Alerting

Create `alerts.yml`:

```yaml
groups:
  - name: masque_vpn_alerts
    rules:
      - alert: HighPacketLoss
        expr: masque_vpn_packet_loss_percent > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High packet loss detected"
          description: "Client {{ $labels.client_id }} has {{ $value }}% packet loss"
      
      - alert: IPPoolExhausted
        expr: masque_vpn_ip_pool_available < 5
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "IP pool nearly exhausted"
          description: "Only {{ $value }} IPs available"
```

## Exporting Data

### Export to CSV

For analysis in Excel/Python:

```bash
# Export specific metric
curl 'http://localhost:9090/api/v1/query?query=masque_vpn_active_connections_total' | jq -r '.data.result[0].value[1]'

# Export time series
curl 'http://localhost:9090/api/v1/query_range?query=masque_vpn_bytes_sent_total&start=2024-01-01T00:00:00Z&end=2024-01-02T00:00:00Z&step=60s'
```

### Export to JSON

```bash
curl 'http://localhost:9090/api/v1/query?query=masque_vpn_active_connections_total' > metrics.json
```

## Troubleshooting

### Target Down

**Symptom**: masque-vpn target shows as "DOWN"

**Solutions**:
- Verify masque-vpn server is running
- Check metrics are enabled in config
- Verify firewall allows port 9090
- Check Prometheus can reach server IP

### No Data

**Symptom**: Queries return no results

**Solutions**:
- Verify metrics endpoint: `curl http://localhost:9090/metrics`
- Check time range in query
- Ensure VPN has active connections
- Verify metric names are correct

### High Memory Usage

**Symptom**: Prometheus uses excessive memory

**Solutions**:
- Reduce retention time
- Increase scrape interval
- Use recording rules for complex queries
- Consider remote storage

## Next Steps

- Set up Grafana dashboards: [grafana-dashboards.md](grafana-dashboards.md)
- Learn more about metrics: [metrics.md](metrics.md)
- Create custom queries for your research

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [PromQL Basics](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Best Practices](https://prometheus.io/docs/practices/)
