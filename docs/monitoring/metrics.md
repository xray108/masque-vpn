# Prometheus Metrics Reference

This document describes all Prometheus metrics exposed by masque-vpn.

## Connection Metrics

### masque_vpn_active_connections_total
- **Type**: Gauge
- **Description**: Current number of active VPN connections
- **Labels**: None

### masque_vpn_connections_total
- **Type**: Counter
- **Description**: Total number of connections established since startup
- **Labels**: `status` (established, failed, closed)

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

### masque_vpn_latency_ms
- **Type**: Histogram
- **Description**: Connection latency in milliseconds
- **Labels**: `client_id`, `quantile` (0.5, 0.95, 0.99)

### masque_vpn_rtt_ms
- **Type**: Gauge
- **Description**: Round-trip time in milliseconds
- **Labels**: `client_id`

### masque_vpn_throughput_mbps
- **Type**: Gauge
- **Description**: Current throughput in Mbps
- **Labels**: `client_id`, `direction` (in, out)

### masque_vpn_packet_loss_percent
- **Type**: Gauge
- **Description**: Packet loss percentage
- **Labels**: `client_id`

## Resource Metrics

### masque_vpn_ip_pool_usage_percent
- **Type**: Gauge
- **Description**: Percentage of IP pool addresses in use
- **Labels**: None

### masque_vpn_ip_pool_available
- **Type**: Gauge
- **Description**: Number of available IP addresses in pool
- **Labels**: None

### masque_vpn_ip_pool_total
- **Type**: Gauge
- **Description**: Total number of IP addresses in pool
- **Labels**: None

## Connection Duration

### masque_vpn_connection_duration_seconds
- **Type**: Histogram
- **Description**: Duration of connections in seconds
- **Labels**: `client_id`

## Error Metrics

### masque_vpn_errors_total
- **Type**: Counter
- **Description**: Total number of errors
- **Labels**: `type` (connection, packet, auth, other)

## Example Queries

### Active Connections
```
masque_vpn_active_connections_total
```

### Average Throughput per Client
```
rate(masque_vpn_bytes_sent_total[5m]) * 8 / 1000000
```

### Connection Success Rate
```
rate(masque_vpn_connections_total{status="established"}[5m]) / 
rate(masque_vpn_connections_total[5m])
```

