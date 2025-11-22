# Grafana Dashboards for MASQUE VPN

This directory contains Grafana dashboard definitions and provisioning configuration for monitoring masque-vpn.

## Dashboards

### overview.json
Main dashboard showing:
- Active connections count
- Traffic rate (bytes sent/received per second)
- IP pool usage percentage
- Round-trip time (RTT) per client
- Connection duration distribution
- Packet rate (packets sent/received per second)

## Setup

### Quick Start with Docker

```bash
# Start Grafana with provisioning
docker run -d \
  -p 3000:3000 \
  -v $(pwd)/grafana/provisioning:/etc/grafana/provisioning \
  -v $(pwd)/grafana/dashboards:/etc/grafana/provisioning/dashboards \
  --name=grafana \
  grafana/grafana
```

Access Grafana at http://localhost:3000 (admin/admin)

### Manual Setup

1. Install Grafana:
```bash
# macOS
brew install grafana
brew services start grafana

# Linux (Ubuntu/Debian)
sudo apt-get install -y grafana
sudo systemctl start grafana-server
```

2. Add Prometheus data source:
   - Open http://localhost:3000
   - Go to Configuration → Data Sources
   - Add Prometheus
   - URL: http://localhost:9090
   - Save & Test

3. Import dashboard:
   - Go to Dashboards → Import
   - Upload `dashboards/overview.json`
   - Select Prometheus data source
   - Import

## Provisioning

The `provisioning/` directory contains configuration for automatic setup:

- `datasources/prometheus.yml` - Prometheus data source configuration
- `dashboards/dashboards.yml` - Dashboard provider configuration

Copy these to Grafana's provisioning directory:
```bash
# macOS (Homebrew)
cp -r grafana/provisioning/* /opt/homebrew/etc/grafana/provisioning/

# Linux
sudo cp -r grafana/provisioning/* /etc/grafana/provisioning/
sudo systemctl restart grafana-server
```

## Customization

All dashboards are editable. Common customizations:

- Adjust time ranges
- Add new panels
- Modify thresholds
- Change refresh intervals
- Add alerts

## For Students

Use these dashboards to:
- Monitor VPN performance in real-time
- Collect data for research projects
- Compare different configurations
- Create visualizations for presentations
- Analyze protocol behavior

See [docs/monitoring/prometheus-setup.md](../docs/monitoring/prometheus-setup.md) for detailed setup instructions.
