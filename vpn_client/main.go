package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	common "github.com/iselt/masque-vpn/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	clientConfig common.ClientConfig
	logger       *zap.Logger

	// Enhanced metrics with additional labels and histograms
	bytesSent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "vpn_client_bytes_sent_total",
		Help: "Total bytes sent by the VPN client",
	}, []string{"direction"})
	
	bytesReceived = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "vpn_client_bytes_received_total",
		Help: "Total bytes received by the VPN client",
	}, []string{"direction"})
	
	connectionStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vpn_client_connection_status",
		Help: "Current connection status (1 = connected, 0 = disconnected)",
	}, []string{"server_addr", "server_name"})
	
	errorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "vpn_client_errors_total",
		Help: "Total number of errors encountered",
	}, []string{"error_type"})
	
	connectionDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "vpn_client_connection_duration_seconds",
		Help: "Duration of VPN connections",
		Buckets: prometheus.DefBuckets,
	})
	
	packetLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "vpn_client_packet_latency_seconds",
		Help: "Packet processing latency",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
	}, []string{"direction"})
	
	activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "vpn_client_active_connections",
		Help: "Number of active QUIC connections",
	})
	
	tunInterfaceStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vpn_client_tun_interface_status",
		Help: "TUN interface status (1 = up, 0 = down)",
	}, []string{"interface_name"})
)

func init() {
	// Register enhanced metrics
	prometheus.MustRegister(bytesSent)
	prometheus.MustRegister(bytesReceived)
	prometheus.MustRegister(connectionStatus)
	prometheus.MustRegister(errorsTotal)
	prometheus.MustRegister(connectionDuration)
	prometheus.MustRegister(packetLatency)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(tunInterfaceStatus)
}

func initLogger(logLevel string) error {
	var config zap.Config
	
	// Determine environment and configure accordingly
	if os.Getenv("ENVIRONMENT") == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	
	// Set log level from configuration
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)
	
	// Configure time encoding
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	
	// Build logger
	logger, err = config.Build()
	if err != nil {
		return err
	}
	
	return nil
}

func main() {
	// Parse command line flags first
	configFile := flag.String("c", "config.client.toml", "Config file path")
	flag.Parse()
	
	// Load configuration
	if _, err := toml.DecodeFile(*configFile, &clientConfig); err != nil {
		panic("Error loading config file " + *configFile + ": " + err.Error())
	}

	// Initialize structured logging with config
	if err := initLogger(clientConfig.LogLevel); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("Starting MASQUE VPN Client",
		zap.String("config_file", *configFile),
		zap.String("log_level", clientConfig.LogLevel),
		zap.String("server_addr", clientConfig.ServerAddr),
		zap.String("server_name", clientConfig.ServerName),
		zap.Int("mtu", clientConfig.MTU),
	)

	if os.Getenv("PERF_PROFILE") != "" {
		f, _ := os.OpenFile("cpu.pprof", os.O_CREATE|os.O_RDWR, 0666)
		defer f.Close()
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
		logger.Info("CPU profiling enabled", zap.String("profile_file", "cpu.pprof"))
	}

	// Start enhanced metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		
		// Add health check endpoint
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		
		logger.Info("Starting metrics server", zap.String("listen_addr", ":9092"))
		if err := http.ListenAndServe(":9092", mux); err != nil {
			logger.Error("Failed to start metrics server", zap.Error(err))
		}
	}()

	// Validate required configuration
	if clientConfig.ServerAddr == "" || clientConfig.ServerName == "" {
		logger.Fatal("Missing required configuration values",
			zap.String("server_addr", clientConfig.ServerAddr),
			zap.String("server_name", clientConfig.ServerName),
		)
	}

	if clientConfig.InsecureSkipVerify {
		logger.Warn("TLS server verification disabled - this is insecure for production use")
	}

	// Create graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	var tunDev *common.TUNDevice
	var masqueConn *common.MASQUEConn
	
	// Track connection start time for duration metric
	connectionStart := time.Now()

	// Establish connection and configure TUN device
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		
		logger.Info("Establishing VPN connection...")
		tunDev, masqueConn, err = establishAndConfigure(ctx)
		if err != nil {
			logger.Error("Failed to establish connection", zap.Error(err))
			errorsTotal.WithLabelValues("connection_failed").Inc()
			stop() // Signal main goroutine to exit if setup fails
			return
		}
		
		logger.Info("Connection established and TUN device configured")
		connectionStatus.WithLabelValues(clientConfig.ServerAddr, clientConfig.ServerName).Set(1)
		activeConnections.Set(1)
		
		if tunDev != nil {
			tunInterfaceStatus.WithLabelValues(tunDev.Name()).Set(1)
		}

		// Start proxy goroutines with enhanced error handling
		errChan := make(chan error, 2)
		var proxyWg sync.WaitGroup

		proxyWg.Add(2)
		go func() {
			defer proxyWg.Done()
			logger.Debug("Starting TUN to VPN proxy")
			common.ProxyFromTunToMASQUE(tunDev, masqueConn, errChan, &clientConfig.FEC)
		}()
		go func() {
			defer proxyWg.Done()
			logger.Debug("Starting VPN to TUN proxy")
			common.ProxyFromMASQUEToTun(tunDev, masqueConn, errChan, &clientConfig.FEC)
		}()

		// Wait for error or shutdown signal
		select {
		case err := <-errChan:
			logger.Error("Proxy error occurred", zap.Error(err))
			errorsTotal.WithLabelValues("proxy_error").Inc()
		case <-ctx.Done():
			logger.Info("Shutdown signal received, stopping proxy")
		}

		// Record connection duration
		connectionDuration.Observe(time.Since(connectionStart).Seconds())
		
		// Update connection status
		connectionStatus.WithLabelValues(clientConfig.ServerAddr, clientConfig.ServerName).Set(0)
		activeConnections.Set(0)
		
		if tunDev != nil {
			tunInterfaceStatus.WithLabelValues(tunDev.Name()).Set(0)
		}

		// Cleanup resources
		logger.Info("Cleaning up resources...")
		if masqueConn != nil {
			if err := masqueConn.Close(); err != nil {
				logger.Warn("Error closing MASQUE connection", zap.Error(err))
			}
		}
		if tunDev != nil {
			if err := tunDev.Close(); err != nil {
				logger.Warn("Error closing TUN device", zap.Error(err))
			}
		}

		// Wait for proxy goroutines to finish
		logger.Debug("Waiting for proxy goroutines to finish")
		proxyWg.Wait()
		logger.Info("All proxy goroutines finished")
	}()

	// Wait for the main goroutine to finish or be signaled
	wg.Wait()
	logger.Info("MASQUE VPN Client shutdown complete")
}

// establishAndConfigure establishes connection to server, sets up TUN device and routing
func establishAndConfigure(ctx context.Context) (*common.TUNDevice, *common.MASQUEConn, error) {
	logger.Info("Configuring TLS settings")
	
	// TLS configuration
	tlsConfig := &tls.Config{
		ServerName:         clientConfig.ServerName,
		InsecureSkipVerify: clientConfig.InsecureSkipVerify,
		NextProtos:         []string{http3.NextProtoH3}, // Required for http3
	}
	
	// Load CA certificate - prioritize PEM string from config
	if clientConfig.CAPEM != "" {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(clientConfig.CAPEM)) {
			return nil, nil, fmt.Errorf("failed to append CA cert from config ca_pem")
		}
		tlsConfig.RootCAs = caCertPool
		tlsConfig.InsecureSkipVerify = false
		logger.Info("Using custom CA from config PEM")
	} else if clientConfig.CAFile != "" {
		caCert, err := os.ReadFile(clientConfig.CAFile)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read CA file %s: %w", clientConfig.CAFile, err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, nil, fmt.Errorf("failed to append CA cert from %s", clientConfig.CAFile)
		}
		tlsConfig.RootCAs = caCertPool
		tlsConfig.InsecureSkipVerify = false
		logger.Info("Using custom CA from file", zap.String("ca_file", clientConfig.CAFile))
	}
	
	// Load client certificate and key - prioritize PEM strings from config
	if clientConfig.CertPEM != "" && clientConfig.KeyPEM != "" {
		cert, err := tls.X509KeyPair([]byte(clientConfig.CertPEM), []byte(clientConfig.KeyPEM))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load client certificate/key from config PEM: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		logger.Info("Loaded client certificate from config PEM")
	} else if clientConfig.TLSCert != "" && clientConfig.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(clientConfig.TLSCert, clientConfig.TLSKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load client certificate/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		logger.Info("Loaded client certificate from files", 
			zap.String("cert_file", clientConfig.TLSCert),
			zap.String("key_file", clientConfig.TLSKey))
	} else {
		return nil, nil, fmt.Errorf("tls_cert and tls_key or cert_pem and key_pem must be set in config for mutual TLS authentication")
	}
	
	// Configure TLS key logging if specified
	if clientConfig.KeyLogFile != "" {
		keyLogWriter, err := os.OpenFile(clientConfig.KeyLogFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			logger.Warn("Failed to create key log file", 
				zap.String("key_log_file", clientConfig.KeyLogFile),
				zap.Error(err))
		} else {
			tlsConfig.KeyLogWriter = keyLogWriter
			defer keyLogWriter.Close()
			logger.Info("TLS key logging enabled", zap.String("key_log_file", clientConfig.KeyLogFile))
		}
	}

	// QUIC connection configuration
	quicConf := &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
	}

	logger.Info("Establishing QUIC connection", zap.String("server_addr", clientConfig.ServerAddr))
	
	// Create UDP socket for dialing
	udpConn, err := net.ListenUDP("udp", nil) // Let OS choose source IP/port
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen on UDP: %w", err)
	}

	serverUdpAddr, err := net.ResolveUDPAddr("udp", clientConfig.ServerAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve server address %s: %w", clientConfig.ServerAddr, err)
	}

	// Dial with timeout
	dialCtx, dialCancel := context.WithTimeout(ctx, 15*time.Second)
	defer dialCancel()

	quicConn, err := quic.Dial(dialCtx, udpConn, serverUdpAddr, tlsConfig, quicConf)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to dial QUIC connection to %s: %w", clientConfig.ServerAddr, err)
	}
	logger.Info("QUIC connection established", 
		zap.String("remote_addr", quicConn.RemoteAddr().String()),
		zap.String("local_addr", quicConn.LocalAddr().String()))

	// Create MASQUE client
	masqueClient := common.NewMASQUEClient(quicConn, logger)

	// Establish MASQUE CONNECT-IP session
	logger.Info("Establishing MASQUE CONNECT-IP session")
	connectCtx, connectCancel := context.WithTimeout(ctx, 10*time.Second)
	defer connectCancel()

	masqueConn, err := masqueClient.ConnectIP(connectCtx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to establish MASQUE CONNECT-IP session: %w", err)
	}
	logger.Info("MASQUE CONNECT-IP session established successfully")

	// Get assigned IP from server (this would be implemented in masque-go)
	// For now, we'll use a default assignment
	assignedIP := "10.0.0.2/32"
	_, assignedPrefix, err := net.ParseCIDR(assignedIP)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse assigned IP: %w", err)
	}

	logger.Info("Configuring TUN device", 
		zap.String("assigned_ip", assignedIP),
		zap.String("tun_name", clientConfig.TunName),
		zap.Int("mtu", clientConfig.MTU))

	dev, err := common.CreateTunDevice(clientConfig.TunName, *assignedPrefix, clientConfig.MTU)
	if err != nil {
		masqueConn.Close()
		return nil, nil, fmt.Errorf("failed to create and configure TUN device: %w", err)
	}
	logger.Info("TUN device configured successfully", 
		zap.String("device_name", dev.Name()),
		zap.String("assigned_ip", assignedIP))

	// Add default route through VPN
	defaultRoute := "0.0.0.0/0"
	_, defaultNet, _ := net.ParseCIDR(defaultRoute)
	if err := dev.AddRoute(*defaultNet); err != nil {
		logger.Warn("Failed to add default route", zap.Error(err))
	} else {
		logger.Info("Added default route through VPN")
	}

	return dev, masqueConn, nil
}