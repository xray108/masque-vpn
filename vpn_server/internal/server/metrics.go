package server

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics содержит все метрики сервера
type Metrics struct {
	// Счетчики соединений
	ActiveConnections prometheus.Gauge
	TotalConnections  prometheus.Counter
	
	// Счетчики пакетов
	PacketsForwarded prometheus.Counter
	PacketsDropped   prometheus.Counter
	BytesForwarded   prometheus.Counter
	
	// Метрики производительности
	PacketProcessingDuration prometheus.Histogram
	ConnectionDuration       prometheus.Histogram
	
	// Метрики ошибок
	ErrorsTotal *prometheus.CounterVec
	
	// Метрики FEC
	FECPacketsEncoded prometheus.Counter
	FECPacketsDecoded prometheus.Counter
	FECRecoveredPackets prometheus.Counter
	
	// Метрики TUN устройства
	TunInterfaceStatus prometheus.Gauge
	TunPacketsRead     prometheus.Counter
	TunPacketsWritten  prometheus.Counter
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics() *Metrics {
	metrics := &Metrics{
		ActiveConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_server_active_connections",
			Help: "Number of active VPN connections",
		}),
		
		TotalConnections: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_total_connections",
			Help: "Total number of VPN connections established",
		}),
		
		PacketsForwarded: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_packets_forwarded_total",
			Help: "Total number of packets forwarded",
		}),
		
		PacketsDropped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_packets_dropped_total",
			Help: "Total number of packets dropped",
		}),
		
		BytesForwarded: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_bytes_forwarded_total",
			Help: "Total bytes forwarded",
		}),
		
		PacketProcessingDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "vpn_server_packet_processing_duration_seconds",
			Help: "Time spent processing packets",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1},
		}),
		
		ConnectionDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "vpn_server_connection_duration_seconds",
			Help: "Duration of VPN connections",
			Buckets: prometheus.DefBuckets,
		}),
		
		ErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "vpn_server_errors_total",
			Help: "Total number of errors by type",
		}, []string{"error_type"}),
		
		FECPacketsEncoded: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_fec_packets_encoded_total",
			Help: "Total number of FEC encoded packets",
		}),
		
		FECPacketsDecoded: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_fec_packets_decoded_total",
			Help: "Total number of FEC decoded packets",
		}),
		
		FECRecoveredPackets: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_fec_recovered_packets_total",
			Help: "Total number of packets recovered using FEC",
		}),
		
		TunInterfaceStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_server_tun_interface_status",
			Help: "TUN interface status (1 = up, 0 = down)",
		}),
		
		TunPacketsRead: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_tun_packets_read_total",
			Help: "Total packets read from TUN interface",
		}),
		
		TunPacketsWritten: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_server_tun_packets_written_total",
			Help: "Total packets written to TUN interface",
		}),
	}
	
	// Регистрируем все метрики
	prometheus.MustRegister(
		metrics.ActiveConnections,
		metrics.TotalConnections,
		metrics.PacketsForwarded,
		metrics.PacketsDropped,
		metrics.BytesForwarded,
		metrics.PacketProcessingDuration,
		metrics.ConnectionDuration,
		metrics.ErrorsTotal,
		metrics.FECPacketsEncoded,
		metrics.FECPacketsDecoded,
		metrics.FECRecoveredPackets,
		metrics.TunInterfaceStatus,
		metrics.TunPacketsRead,
		metrics.TunPacketsWritten,
	)
	
	return metrics
}

// RecordConnection записывает метрики нового соединения
func (m *Metrics) RecordConnection() {
	m.TotalConnections.Inc()
	m.ActiveConnections.Inc()
}

// RecordDisconnection записывает метрики закрытого соединения
func (m *Metrics) RecordDisconnection() {
	m.ActiveConnections.Dec()
}

// RecordError записывает метрику ошибки
func (m *Metrics) RecordError(errorType string) {
	m.ErrorsTotal.WithLabelValues(errorType).Inc()
}

// RecordPacketProcessing записывает время обработки пакета
func (m *Metrics) RecordPacketProcessing(duration float64) {
	m.PacketProcessingDuration.Observe(duration)
}

// RecordConnectionDuration записывает продолжительность соединения
func (m *Metrics) RecordConnectionDuration(duration float64) {
	m.ConnectionDuration.Observe(duration)
}