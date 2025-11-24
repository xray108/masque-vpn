package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	activeConnections    prometheus.Gauge
	connectionsTotal     *prometheus.CounterVec
	bytesSentTotal      *prometheus.CounterVec
	bytesReceivedTotal  *prometheus.CounterVec
	packetsSentTotal    *prometheus.CounterVec
	packetsReceivedTotal *prometheus.CounterVec
	latency             *prometheus.HistogramVec
	rtt                 *prometheus.GaugeVec
	throughput          *prometheus.GaugeVec
	packetLoss          *prometheus.GaugeVec
	ipPoolUsage         prometheus.Gauge
	ipPoolAvailable     prometheus.Gauge
	ipPoolTotal         prometheus.Gauge
	connectionDuration  *prometheus.HistogramVec
	errorsTotal         *prometheus.CounterVec
}

var (
	metricsInstance *Metrics
	metricsOnce     sync.Once
)

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsInstance = &Metrics{
			activeConnections: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "masque_vpn_active_connections_total",
				Help: "Current number of active VPN connections",
			}),
			connectionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_connections_total",
				Help: "Total number of connections established since startup",
			}, []string{"status"}),
			bytesSentTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_bytes_sent_total",
				Help: "Total bytes sent to clients",
			}, []string{"client_id"}),
			bytesReceivedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_bytes_received_total",
				Help: "Total bytes received from clients",
			}, []string{"client_id"}),
			packetsSentTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_packets_sent_total",
				Help: "Total packets sent to clients",
			}, []string{"client_id"}),
			packetsReceivedTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_packets_received_total",
				Help: "Total packets received from clients",
			}, []string{"client_id"}),
			latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "masque_vpn_latency_ms",
				Help:    "Connection latency in milliseconds",
				Buckets: prometheus.ExponentialBuckets(1, 2, 10),
			}, []string{"client_id"}),
			rtt: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "masque_vpn_rtt_ms",
				Help: "Round-trip time in milliseconds",
			}, []string{"client_id"}),
			throughput: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "masque_vpn_throughput_mbps",
				Help: "Current throughput in Mbps",
			}, []string{"client_id", "direction"}),
			packetLoss: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "masque_vpn_packet_loss_percent",
				Help: "Packet loss percentage",
			}, []string{"client_id"}),
			ipPoolUsage: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "masque_vpn_ip_pool_usage_percent",
				Help: "Percentage of IP pool addresses in use",
			}),
			ipPoolAvailable: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "masque_vpn_ip_pool_available",
				Help: "Number of available IP addresses in pool",
			}),
			ipPoolTotal: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "masque_vpn_ip_pool_total",
				Help: "Total number of IP addresses in pool",
			}),
			connectionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
				Name:    "masque_vpn_connection_duration_seconds",
				Help:    "Duration of connections in seconds",
				Buckets: prometheus.ExponentialBuckets(1, 2, 15),
			}, []string{"client_id"}),
			errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: "masque_vpn_errors_total",
				Help: "Total number of errors",
			}, []string{"type"}),
		}

		prometheus.MustRegister(
			metricsInstance.activeConnections,
			metricsInstance.connectionsTotal,
			metricsInstance.bytesSentTotal,
			metricsInstance.bytesReceivedTotal,
			metricsInstance.packetsSentTotal,
			metricsInstance.packetsReceivedTotal,
			metricsInstance.latency,
			metricsInstance.rtt,
			metricsInstance.throughput,
			metricsInstance.packetLoss,
			metricsInstance.ipPoolUsage,
			metricsInstance.ipPoolAvailable,
			metricsInstance.ipPoolTotal,
			metricsInstance.connectionDuration,
			metricsInstance.errorsTotal,
		)
	})

	return metricsInstance
}

// StartMetricsServer starts the Prometheus metrics HTTP server
func StartMetricsServer(addr string) error {
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(addr, nil)
}

// ClientMetrics tracks metrics for a single client connection
type ClientMetrics struct {
	clientID       string
	startTime      time.Time
	bytesSent      uint64
	bytesReceived  uint64
	packetsSent    uint64
	packetsReceived uint64
	lastUpdate     time.Time
	mu             sync.RWMutex
}

// NewClientMetrics creates a new ClientMetrics instance
func NewClientMetrics(clientID string) *ClientMetrics {
	return &ClientMetrics{
		clientID:   clientID,
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// RecordBytesSent records bytes sent to the client
func (cm *ClientMetrics) RecordBytesSent(bytes uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.bytesSent += bytes
	cm.packetsSent++
	metricsInstance.bytesSentTotal.WithLabelValues(cm.clientID).Add(float64(bytes))
	metricsInstance.packetsSentTotal.WithLabelValues(cm.clientID).Inc()
}

// RecordBytesReceived records bytes received from the client
func (cm *ClientMetrics) RecordBytesReceived(bytes uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.bytesReceived += bytes
	cm.packetsReceived++
	metricsInstance.bytesReceivedTotal.WithLabelValues(cm.clientID).Add(float64(bytes))
	metricsInstance.packetsReceivedTotal.WithLabelValues(cm.clientID).Inc()
}

// RecordLatency records connection latency
func (cm *ClientMetrics) RecordLatency(latencyMs float64) {
	metricsInstance.latency.WithLabelValues(cm.clientID).Observe(latencyMs)
}

// RecordRTT records round-trip time
func (cm *ClientMetrics) RecordRTT(rttMs float64) {
	metricsInstance.rtt.WithLabelValues(cm.clientID).Set(rttMs)
}

// RecordThroughput records throughput
func (cm *ClientMetrics) RecordThroughput(mbps float64, direction string) {
	metricsInstance.throughput.WithLabelValues(cm.clientID, direction).Set(mbps)
}

// RecordPacketLoss records packet loss percentage
func (cm *ClientMetrics) RecordPacketLoss(percent float64) {
	metricsInstance.packetLoss.WithLabelValues(cm.clientID).Set(percent)
}

// Close records connection duration and cleans up
func (cm *ClientMetrics) Close() {
	duration := time.Since(cm.startTime).Seconds()
	metricsInstance.connectionDuration.WithLabelValues(cm.clientID).Observe(duration)
}

// GetStats returns current statistics
func (cm *ClientMetrics) GetStats() (bytesSent, bytesReceived, packetsSent, packetsReceived uint64) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.bytesSent, cm.bytesReceived, cm.packetsSent, cm.packetsReceived
}

