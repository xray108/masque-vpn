# Monitoring Setup

This guide explains how to set up monitoring for masque-vpn using Prometheus and Grafana.

## Prerequisites

- Prometheus server
- Grafana server
- masque-vpn server with metrics enabled

## Configuration

### Enable Metrics in masque-vpn

Metrics are exposed on the `/metrics` endpoint (default port 9090).

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'masque-vpn'
    static_configs:
      - targets: ['localhost:9090']
```

### Grafana Dashboard

Import the dashboard from `docs/monitoring/grafana-dashboard.json`.

## Metrics Endpoint

The metrics endpoint is available at:
```
http://localhost:9090/metrics
```

For more details, see [Metrics Reference](metrics.md).

