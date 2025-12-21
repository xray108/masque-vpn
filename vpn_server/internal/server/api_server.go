package server

import (
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// APIServer представляет HTTP API сервер для управления VPN
type APIServer struct {
	server *Server
	router *gin.Engine
	// Временное хранение в памяти вместо SQLite
	connectionLogs []ConnectionLog
	logsMutex      sync.RWMutex
}

// ConnectionLog представляет лог соединения
type ConnectionLog struct {
	ID        int       `json:"id"`
	ClientID  string    `json:"client_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details"`
}

// ClientInfo информация о клиенте для API
type ClientInfo struct {
	ID          string    `json:"id"`
	AssignedIP  string    `json:"assigned_ip"`
	ConnectedAt time.Time `json:"connected_at"`
	BytesSent   int64     `json:"bytes_sent"`
	BytesRecv   int64     `json:"bytes_received"`
	Status      string    `json:"status"`
}

// ServerStats статистика сервера
type ServerStats struct {
	ActiveConnections int           `json:"active_connections"`
	TotalConnections  int64         `json:"total_connections"`
	NetworkCIDR       string        `json:"network_cidr"`
	TunDevice         string        `json:"tun_device"`
	Uptime            time.Duration `json:"uptime"`
	PacketsForwarded  int64         `json:"packets_forwarded"`
	BytesForwarded    int64         `json:"bytes_forwarded"`
}

// NewAPIServer создает новый API сервер
func NewAPIServer(server *Server) (*APIServer, error) {
	// Настраиваем Gin в production режиме
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	apiServer := &APIServer{
		server:         server,
		router:         router,
		connectionLogs: make([]ConnectionLog, 0),
	}

	// Настраиваем маршруты
	apiServer.setupRoutes()

	return apiServer, nil
}

// setupRoutes настраивает API маршруты
func (api *APIServer) setupRoutes() {
	// Статические файлы (admin UI)
	if api.server.Config.APIServer.StaticDir != "" {
		api.router.Static("/static", api.server.Config.APIServer.StaticDir)
		api.router.StaticFile("/", filepath.Join(api.server.Config.APIServer.StaticDir, "index.html"))
	}

	// API маршруты
	v1 := api.router.Group("/api/v1")
	{
		// Информация о сервере
		v1.GET("/status", api.getServerStatus)
		v1.GET("/stats", api.getServerStats)

		// Управление клиентами
		v1.GET("/clients", api.getClients)
		v1.GET("/clients/:id", api.getClient)
		v1.DELETE("/clients/:id", api.disconnectClient)

		// Логи соединений
		v1.GET("/logs", api.getConnectionLogs)

		// Конфигурация
		v1.GET("/config", api.getConfig)
	}

	// Health check
	api.router.GET("/health", api.healthCheck)

	// Метрики Prometheus
	api.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// Start запускает API сервер
func (api *APIServer) Start() error {
	log.Printf("Starting API server on %s", api.server.Config.APIServer.ListenAddr)
	return api.router.Run(api.server.Config.APIServer.ListenAddr)
}

// Close закрывает API сервер
func (api *APIServer) Close() error {
	// Очищаем логи
	api.logsMutex.Lock()
	api.connectionLogs = nil
	api.logsMutex.Unlock()
	return nil
}

// healthCheck обработчик проверки здоровья
func (api *APIServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "masque-vpn-server",
		"time":    time.Now().UTC(),
	})
}

// getServerStatus возвращает статус сервера
func (api *APIServer) getServerStatus(c *gin.Context) {
	api.server.IPPoolMu.RLock()
	activeConnections := len(api.server.IPConnMap)
	api.server.IPPoolMu.RUnlock()

	tunDevice := "disabled"
	if api.server.TunDev != nil {
		tunDevice = api.server.TunDev.Name()
	}

	status := gin.H{
		"status":             "running",
		"active_connections": activeConnections,
		"network_cidr":       api.server.Config.AssignCIDR,
		"tun_device":         tunDevice,
		"listen_addr":        api.server.Config.ListenAddr,
		"server_name":        api.server.Config.ServerName,
	}

	c.JSON(http.StatusOK, status)
}

// getServerStats возвращает статистику сервера
func (api *APIServer) getServerStats(c *gin.Context) {
	api.server.IPPoolMu.RLock()
	activeConnections := len(api.server.IPConnMap)
	api.server.IPPoolMu.RUnlock()

	tunDevice := "disabled"
	if api.server.TunDev != nil {
		tunDevice = api.server.TunDev.Name()
	}

	stats := ServerStats{
		ActiveConnections: activeConnections,
		NetworkCIDR:       api.server.Config.AssignCIDR,
		TunDevice:         tunDevice,
		// TODO: Добавить реальные метрики из Prometheus
		TotalConnections: 0,
		Uptime:           0,
		PacketsForwarded: 0,
		BytesForwarded:   0,
	}

	c.JSON(http.StatusOK, stats)
}

// getClients возвращает список подключенных клиентов
func (api *APIServer) getClients(c *gin.Context) {
	api.server.IPPoolMu.RLock()
	defer api.server.IPPoolMu.RUnlock()

	clients := make([]ClientInfo, 0, len(api.server.ClientIPMap))

	for clientID, assignedIP := range api.server.ClientIPMap {
		status := "disconnected"
		if _, exists := api.server.IPConnMap[assignedIP]; exists {
			status = "connected"
		}

		client := ClientInfo{
			ID:         clientID,
			AssignedIP: assignedIP.String(),
			Status:     status,
			// TODO: Добавить реальные метрики
			ConnectedAt: time.Now(),
			BytesSent:   0,
			BytesRecv:   0,
		}
		clients = append(clients, client)
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clients,
		"total":   len(clients),
	})
}

// getClient возвращает информацию о конкретном клиенте
func (api *APIServer) getClient(c *gin.Context) {
	clientID := c.Param("id")

	api.server.IPPoolMu.RLock()
	assignedIP, exists := api.server.ClientIPMap[clientID]
	if !exists {
		api.server.IPPoolMu.RUnlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	status := "disconnected"
	if _, connected := api.server.IPConnMap[assignedIP]; connected {
		status = "connected"
	}
	api.server.IPPoolMu.RUnlock()

	client := ClientInfo{
		ID:         clientID,
		AssignedIP: assignedIP.String(),
		Status:     status,
		// TODO: Добавить реальные метрики
		ConnectedAt: time.Now(),
		BytesSent:   0,
		BytesRecv:   0,
	}

	c.JSON(http.StatusOK, client)
}

// disconnectClient отключает клиента
func (api *APIServer) disconnectClient(c *gin.Context) {
	clientID := c.Param("id")

	api.server.IPPoolMu.Lock()
	assignedIP, exists := api.server.ClientIPMap[clientID]
	if !exists {
		api.server.IPPoolMu.Unlock()
		c.JSON(http.StatusNotFound, gin.H{"error": "Client not found"})
		return
	}

	// Закрываем соединение если оно активно
	if session, connected := api.server.IPConnMap[assignedIP]; connected {
		if session.Conn != nil {
			session.Conn.Close()
		}
		delete(api.server.IPConnMap, assignedIP)
	}

	// Удаляем клиента из карты
	delete(api.server.ClientIPMap, clientID)
	api.server.IPPoolMu.Unlock()

	// Освобождаем IP
	api.server.IPPool.Release(assignedIP)

	log.Printf("API: Disconnected client %s (IP: %s)", clientID, assignedIP)

	c.JSON(http.StatusOK, gin.H{
		"message":   "Client disconnected",
		"client_id": clientID,
	})
}

// getConnectionLogs возвращает логи соединений
func (api *APIServer) getConnectionLogs(c *gin.Context) {
	api.logsMutex.RLock()
	defer api.logsMutex.RUnlock()

	// Возвращаем последние 100 логов
	logs := api.connectionLogs
	if len(logs) > 100 {
		logs = logs[len(logs)-100:]
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  logs,
		"total": len(logs),
	})
}

// AddConnectionLog добавляет лог соединения
func (api *APIServer) AddConnectionLog(clientID, eventType, details string) {
	api.logsMutex.Lock()
	defer api.logsMutex.Unlock()

	log := ConnectionLog{
		ID:        len(api.connectionLogs) + 1,
		ClientID:  clientID,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Details:   details,
	}

	api.connectionLogs = append(api.connectionLogs, log)

	// Ограничиваем размер логов
	if len(api.connectionLogs) > 1000 {
		api.connectionLogs = api.connectionLogs[100:]
	}
}

// getConfig возвращает конфигурацию сервера (без секретов)
func (api *APIServer) getConfig(c *gin.Context) {
	config := gin.H{
		"listen_addr":       api.server.Config.ListenAddr,
		"assign_cidr":       api.server.Config.AssignCIDR,
		"advertise_routes":  api.server.Config.AdvertiseRoutes,
		"server_name":       api.server.Config.ServerName,
		"mtu":               api.server.Config.MTU,
		"log_level":         api.server.Config.LogLevel,
		"enable_ipv6":       api.server.Config.EnableIPv6,
		"fec_enabled":       api.server.Config.FEC.Enabled,
		"metrics_enabled":   api.server.Config.Metrics.Enabled,
	}

	c.JSON(http.StatusOK, config)
}