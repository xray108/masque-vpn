package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"sync"
	"time"

	common "github.com/iselt/masque-vpn/common"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
)

// Server представляет MASQUE VPN сервер
type Server struct {
	Config      common.ServerConfig
	TunDev      *common.TUNDevice
	IPPool      *common.IPPool
	ClientIPMap map[string]netip.Addr
	IPConnMap   map[netip.Addr]*ClientSession
	IPPoolMu    sync.RWMutex
	Metrics     *Metrics
}

// New создает новый экземпляр сервера
func New(config common.ServerConfig) (*Server, error) {
	// Создаем IP пул
	networkInfo, err := common.NewNetworkInfo(config.AssignCIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to create network info: %w", err)
	}

	ipPool := common.NewIPPool(networkInfo.GetPrefix(), networkInfo.GetGateway().Addr())

	// Создаем TUN устройство
	tunDev, err := common.CreateTunDevice(config.TunName, net.IPNet{
		IP:   networkInfo.GetGateway().Addr().AsSlice(),
		Mask: net.CIDRMask(networkInfo.GetGateway().Bits(), networkInfo.GetGateway().Addr().BitLen()),
	}, config.MTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Инициализируем метрики
	metrics := NewMetrics()
	metrics.TunInterfaceStatus.Set(1) // TUN устройство активно

	server := &Server{
		Config:      config,
		TunDev:      tunDev,
		IPPool:      ipPool,
		ClientIPMap: make(map[string]netip.Addr),
		IPConnMap:   make(map[netip.Addr]*ClientSession),
		Metrics:     metrics,
	}

	// Запускаем обработчик пакетов
	go server.processPackets()

	log.Printf("MASQUE VPN Server initialized")
	log.Printf("Listen Address: %s", server.Config.ListenAddr)
	log.Printf("VPN Network: %s", server.Config.AssignCIDR)
	log.Printf("TUN Device: %s", tunDev.Name())
	log.Printf("Advertised Routes: %v", server.Config.AdvertiseRoutes)

	return server, nil
}

// Run запускает сервер
func (s *Server) Run(ctx context.Context) error {
	// Настраиваем TLS
	tlsConfig, err := s.setupTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to setup TLS: %w", err)
	}

	// Настраиваем QUIC
	quicConf := &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
	}

	// Создаем HTTP/3 сервер
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleMASQUERequest)
	
	// Добавляем эндпоинт для метрик
	mux.Handle("/metrics", s.createMetricsHandler())
	
	// Добавляем эндпоинт для проверки здоровья
	mux.HandleFunc("/health", s.handleHealthCheck)

	server := &http3.Server{
		Addr:       s.Config.ListenAddr,
		Handler:    mux,
		TLSConfig:  tlsConfig,
		QUICConfig: quicConf,
	}

	log.Printf("MASQUE VPN Server listening on %s", s.Config.ListenAddr)
	
	// Запускаем сервер в отдельной горутине
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ListenAndServe()
	}()

	// Ждем завершения или ошибки
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Printf("Shutting down server...")
		return server.Close()
	}
}

// handleHealthCheck обрабатывает запросы проверки здоровья
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"masque-vpn-server"}`))
}

// createMetricsHandler создает обработчик для метрик Prometheus
func (s *Server) createMetricsHandler() http.Handler {
	// Импортируем prometheus handler
	// return promhttp.Handler()
	// Пока что простая заглушка
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metrics endpoint"))
	})
}

// Close закрывает сервер и освобождает ресурсы
func (s *Server) Close() error {
	log.Printf("Closing MASQUE VPN Server...")
	
	// Закрываем все клиентские соединения
	s.IPPoolMu.Lock()
	for clientID, ip := range s.ClientIPMap {
		if session, exists := s.IPConnMap[ip]; exists {
			if session.Conn != nil {
				session.Conn.Close()
			}
		}
		log.Printf("Closed connection for client %s", clientID)
	}
	s.IPPoolMu.Unlock()

	// Закрываем TUN устройство
	if s.TunDev != nil {
		if err := s.TunDev.Close(); err != nil {
			log.Printf("Error closing TUN device: %v", err)
		}
		s.Metrics.TunInterfaceStatus.Set(0)
	}

	log.Printf("MASQUE VPN Server closed")
	return nil
}