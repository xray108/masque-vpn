# Grafana Dashboards Setup Guide

This guide explains how to set up and use Grafana dashboards for monitoring masque-vpn.

## Prerequisites

- masque-vpn server running with metrics enabled
- Prometheus collecting metrics from the server
- Grafana installed

## Quick Setup

### 1. Install Grafana

**macOS:**
```bash
brew install grafana
brew services start grafana
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install -y grafana
sudo systemctl start grafana-server
sudo systemctl enable grafana-server
```

**Docker:**
```bash
docker run -d \
  -p 3000:3000 \
  -v $(pwd)/grafana/provisioning:/etc/grafana/provisioning \
  -v $(pwd)/grafana/dashboards:/etc/grafana/provisioning/dashboards \
  --name=grafana \
  grafana/grafana
```

### 2. Access Grafana

Open http://localhost:3000

Default credentials:
- Username: `admin`
- Password: `admin`

You will be prompted to change the password on first login.

### 3. Add Prometheus Data Source

1. Click Configuration (gear icon) → Data Sources
2. Click "Add data source"
3. Select "Prometheus"
4. Configure:
   - Name: `Prometheus`
   - URL: `http://localhost:9090`
   - Access: `Server (default)`
5. Click "Save & Test"

Expected: Green checkmark with "Data source is working"

### 4. Import Dashboard

1. Click Dashboards (four squares icon) → Import
2. Click "Upload JSON file"
3. Select `grafana/dashboards/overview.json`
4. Select Prometheus data source
5. Click "Import"

## Dashboard Overview

The overview dashboard includes:

### Active Connections
- Real-time count of connected VPN clients
- Useful for monitoring server load

### Traffic Rate
- Bytes sent/received per second
- Separate lines for each client
- Helps identify bandwidth usage patterns

### IP Pool Usage
- Percentage of allocated IP addresses
- Gauge with color thresholds (green/yellow/red)
- Alerts when pool is nearly exhausted

### Round-Trip Time (RTT)
- Latency measurements per client
- Shows mean and max values
- Useful for diagnosing network issues

### Connection Duration
- Distribution of connection lifetimes
- Shows p50 and p95 percentiles
- Helps understand usage patterns

### Packet Rate
- Packets sent/received per second
- Per-client breakdown
- Useful for protocol analysis

## Using Dashboards for Research

### Collecting Data

1. **Baseline Measurement**
   - Run VPN with default settings
   - Record metrics for 10-15 minutes
   - Export data for analysis

2. **Configuration Changes**
   - Modify server/client settings
   - Run same tests
   - Compare with baseline

3. **Network Conditions**
   - Use tc/netem to simulate latency/loss
   - Observe impact on metrics
   - Document findings

### Exporting Data

**Screenshot:**
- Click panel title → View → Take screenshot
- Use for presentations

**CSV Export:**
- Click panel title → Inspect → Data
- Download CSV
- Import into Excel/Python

**JSON Export:**
- Dashboard settings → JSON Model
- Copy JSON
- Use for programmatic analysis

## Customization

### Adding Panels

1. Click "Add panel" button
2. Select visualization type
3. Configure query:
   ```promql
   rate(masque_vpn_bytes_sent_total[1m])
   ```
4. Adjust display options
5. Save dashboard

### Creating Alerts

1. Edit panel
2. Click Alert tab
3. Configure conditions:
   ```
   WHEN avg() OF query(A, 5m, now) IS ABOVE 100
   ```
4. Set notification channel
5. Save

### Time Ranges

- Default: Last 15 minutes
- Change via time picker (top right)
- Useful ranges for research:
  - Last 1 hour: Real-time monitoring
  - Last 6 hours: Session analysis
  - Last 24 hours: Daily patterns
  - Last 7 days: Weekly trends

## Troubleshooting

### Dashboard Shows "No Data"

**Check:**
1. Prometheus is running: `curl http://localhost:9090`
2. Metrics endpoint works: `curl http://localhost:9090/metrics`
3. VPN server has metrics enabled in config
4. Time range includes recent data

### Metrics Not Updating

**Solutions:**
1. Verify VPN server is running
2. Check Prometheus scrape interval (default: 15s)
3. Refresh dashboard (top right)
4. Check browser console for errors

### Connection Refused

**Check:**
1. Grafana is running: `brew services list` or `systemctl status grafana-server`
2. Port 3000 is not blocked by firewall
3. Grafana logs: `/var/log/grafana/grafana.log`

## Best Practices

### For Students

1. **Document Everything**
   - Take screenshots of dashboards
   - Export data regularly
   - Note configuration changes

2. **Use Consistent Time Ranges**
   - Compare same duration periods
   - Account for warmup time
   - Run multiple iterations

3. **Create Custom Dashboards**
   - Focus on your research questions
   - Combine relevant metrics
   - Share with classmates

### For Research Projects

1. **Baseline First**
   - Always establish baseline performance
   - Document test environment
   - Record all variables

2. **Controlled Changes**
   - Change one variable at a time
   - Run sufficient iterations
   - Use statistical analysis

3. **Visualization**
   - Create clear, labeled graphs
   - Use appropriate scales
   - Include error bars if applicable

## Next Steps

- Set up Prometheus: [prometheus-setup.md](prometheus-setup.md)
- Learn PromQL: [Prometheus Query Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- Create custom dashboards for your research
- Explore advanced Grafana features (annotations, variables, etc.)

## References

- [Grafana Documentation](https://grafana.com/docs/)
- [Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/best-practices/)
- [PromQL Examples](https://prometheus.io/docs/prometheus/latest/querying/examples/)
