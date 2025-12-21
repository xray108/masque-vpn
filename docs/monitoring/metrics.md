# Prometheus Metrics Reference

This document describes all Prometheus metrics exposed by masque-vpn server.

## Metrics Endpoint

Metrics are available at: `http://server:8080/metrics` (same port as API server)

## Connection Metrics

### masque_vpn_active_connections
- **Type**: Gauge
- **Description**: Current number of active VPN connections
- **Labels**: None

### masque_vpn_total_connections
- **Type**: Counter  
- **Description**: Total number of connections established since startup
- **Labels**: None

### masque_vpn_connection_duration_seconds
- **Type**: Histogram
- **Description**: Duration of connections in seconds
- **Labels**: `client_id`

## Traffic Metrics

### masque_vpn_bytes_sent_total
- **Type**: Counter
- **Description**: Total bytes sent to clients
- **Labels**: `client_id`

### masque_vpn_bytes_received_total
- **Type**: Counter
- **Description**: Total bytes received from clients
- **Labels**: `client_id`

### masque_vpn_packets_sent_total
- **Type**: Counter
- **Description**: Total packets sent to clients
- **Labels**: `client_id`

### masque_vpn_packets_received_total
- **Type**: Counter
- **Description**: Total packets received from clients
- **Labels**: `client_id`

## Performance Metrics

### masque_vpn_packet_processing_duration_seconds
- **Type**: Histogram
- **Description**: Time spent processing packets
- **Labels**: `direction` (inbound, outbound)
- **Buckets**: 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0

### masque_vpn_quic_stream_duration_seconds
- **Type**: Histogram
- **Description**: Duration of QUIC streams
- **Labels**: `client_id`
- **Buckets**: 1, 5, 10, 30, 60, 300, 600, 1800, 3600

### masque_vpn_masque_request_duration_seconds
- **Type**: Histogram
- **Description**: Time to process MASQUE CONNECT-IP requests
- **Labels**: `method` (CONNECT)
- **Buckets**: 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0

## Resource Metrics

### masque_vpn_ip_pool_total
- **Type**: Gauge
- **Description**: Total number of IP addresses in pool
- **Labels**: None

### masque_vpn_ip_pool_used
- **Type**: Gauge
- **Description**: Number of IP addresses currently assigned
- **Labels**: None

### masque_vpn_ip_pool_available
- **Type**: Gauge
- **Description**: Number of available IP addresses in pool
- **Labels**: None

## System Metrics

### masque_vpn_tun_interface_status
- **Type**: Gauge
- **Description**: TUN interface status (1=active, 0=inactive)
- **Labels**: None

### masque_vpn_server_uptime_seconds
- **Type**: Counter
- **Description**: Server uptime in seconds
- **Labels**: None

## Error Metrics

### masque_vpn_errors_total
- **Type**: Counter
- **Description**: Total number of errors by type
- **Labels**: `error_type` (connection, packet, auth, tun, masque)

### masque_vpn_packet_drops_total
- **Type**: Counter
- **Description**: Total number of dropped packets
- **Labels**: `reason` (invalid, oversized, processing_error)

## MASQUE Protocol Metrics

### masque_vpn_masque_requests_total
- **Type**: Counter
- **Description**: Total MASQUE CONNECT-IP requests
- **Labels**: `method` (CONNECT), `status` (success, failed)

### masque_vpn_quic_streams_total
- **Type**: Counter
- **Description**: Total QUIC streams created
- **Labels**: `client_id`, `stream_type` (bidirectional, unidirectional)

## Example Queries

### Active Connections
```promql
masque_vpn_active_connections
```

### Connection Success Rate
```promql
rate(masque_vpn_total_connections[5m])
```

### Average Packet Processing Time
```promql
rate(masque_vpn_packet_processing_duration_seconds_sum[5m]) / 
rate(masque_vpn_packet_processing_duration_seconds_count[5m])
```

### IP Pool Utilization
```promql
(masque_vpn_ip_pool_used / masque_vpn_ip_pool_total) * 100
```

### Error Rate by Type
```promql
rate(masque_vpn_errors_total[5m])
```

### MASQUE Request Latency (95th percentile)
```promql
histogram_quantile(0.95, rate(masque_vpn_masque_request_duration_seconds_bucket[5m]))
```

## Grafana Dashboard

Import the dashboard from `grafana/masque-vpn-dashboard.json` for pre-configured visualizations.

Key panels include:
- Active connections over time
- Packet processing latency
- IP pool utilization
- Error rates
- MASQUE protocol statistics

## Alerting Rules

Example Prometheus alerting rules:

```yaml
groups:
- name: masque-vpn
  rules:
  - alert: HighErrorRate
    expr: rate(masque_vpn_errors_total[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate in MASQUE VPN"
      
  - alert: IPPoolExhaustion
    expr: (masque_vpn_ip_pool_available / masque_vpn_ip_pool_total) < 0.1
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "IP pool nearly exhausted"
```

