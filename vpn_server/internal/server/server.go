package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	connectip "github.com/iselt/connect-ip-go"
	common "github.com/iselt/masque-vpn/common"
	common_fec "github.com/iselt/masque-vpn/common/fec"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/yosida95/uritemplate/v3"
)

type Server struct {
	Config      common.ServerConfig
	TunDev      *common.TUNDevice
	IPPool      *common.IPPool
	ClientIPMap map[string]netip.Addr
	IPConnMap   map[netip.Addr]*ClientSession
	IPPoolMu    sync.RWMutex
	Metrics     *Metrics
}

func New(config common.ServerConfig) (*Server, error) {
	// --- 创建 IP 分配器 ---
	networkInfo, err := common.NewNetworkInfo(config.AssignCIDR)
	if err != nil {
		return nil, fmt.Errorf("failed to create IP allocator: %v", err)
	}

	// --- 创建 TUN 设备 ---
	tunDev, err := common.CreateTunDevice(config.TunName, networkInfo.GetGateway(), config.MTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %v", err)
	}

	s := &Server{
		Config:      config,
		TunDev:      tunDev,
		IPPool:      common.NewIPPool(networkInfo.GetPrefix(), networkInfo.GetGateway().Addr()),
		ClientIPMap: make(map[string]netip.Addr),
		IPConnMap:   make(map[netip.Addr]*ClientSession),
	}

	if config.Metrics.Enabled {
		s.Metrics = NewMetrics()
	}

	return s, nil
}

func (s *Server) Run() {
	defer s.TunDev.Close()

	// Start packet processing loop
	go s.processPackets()

	log.Printf("Starting VPN Server...")
	log.Printf("Listen Address: %s", s.Config.ListenAddr)
	log.Printf("VPN Network: %s", s.Config.AssignCIDR) // Using AssignCIDR as proxy for network info prefix
	log.Printf("Advertised Routes: %v", s.Config.AdvertiseRoutes)

	// --- 准备路由信息 ---
	var routesToAdvertise []connectip.IPRoute
	for _, routeStr := range s.Config.AdvertiseRoutes {
		prefix, err := netip.ParsePrefix(routeStr)
		if err != nil {
			log.Fatalf("Invalid route in advertise_routes '%s': %v", routeStr, err)
		}
		routesToAdvertise = append(routesToAdvertise, connectip.IPRoute{
			StartIP:    prefix.Addr(),
			EndIP:      common.LastIP(prefix),
			IPProtocol: 0, // 0 表示任何协议
		})
	}

	// --- TLS 配置 ---
	var cert tls.Certificate
	var err error
	if s.Config.CertPEM != "" && s.Config.KeyPEM != "" {
		cert, err = tls.X509KeyPair([]byte(s.Config.CertPEM), []byte(s.Config.KeyPEM))
		if err != nil {
			log.Fatalf("Failed to load TLS certificate/key from config PEM: %v", err)
		}
		log.Printf("Loaded TLS certificate/key from config PEM")
	} else {
		cert, err = tls.LoadX509KeyPair(s.Config.CertFile, s.Config.KeyFile)
		if err != nil {
			log.Fatalf("Failed to load TLS certificate/key: %v", err)
		}
		log.Printf("Loaded TLS certificate/key from file: %s, %s", s.Config.CertFile, s.Config.KeyFile)
	}

	// 加载CA证书
	caCertPool := x509.NewCertPool()
	var caCert []byte
	if s.Config.CACertPEM != "" {
		caCert = []byte(s.Config.CACertPEM)
		log.Printf("Loaded CA cert from config PEM")
	} else {
		if s.Config.CACertFile == "" {
			log.Fatal("ca_file is required for mutual TLS authentication")
		}
		caCert, err = os.ReadFile(s.Config.CACertFile)
		if err != nil {
			log.Fatalf("Failed to read CA file %s: %v", s.Config.CACertFile, err)
		}
		log.Printf("Loaded CA cert from file: %s", s.Config.CACertFile)
	}
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("Failed to append CA cert")
	}

	tlsConfig := http3.ConfigureTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	})

	// --- QUIC 配置 ---
	quicConf := &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 30 * time.Second,
	}

	// --- QUIC 监听器 ---
	listenNetAddr, err := net.ResolveUDPAddr("udp", s.Config.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve listen address %s: %v", s.Config.ListenAddr, err)
	}

	udpConn, err := net.ListenUDP("udp", listenNetAddr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP %s: %v", s.Config.ListenAddr, err)
	}
	defer udpConn.Close()

	log.Printf("Creating QUIC listener on %s...", s.Config.ListenAddr)
	ln, err := quic.ListenEarly(udpConn, tlsConfig, quicConf)
	if err != nil {
		log.Fatalf("Failed to create QUIC listener: %v", err)
	}
	defer ln.Close()
	log.Printf("QUIC Listener started on %s", udpConn.LocalAddr())

	// Initialize metrics
	if s.Metrics != nil {
		log.Printf("Metrics enabled, will be available at %s/metrics", s.Config.Metrics.ListenAddr)
		
		// Initialize IP pool metrics
		total, _, available := s.IPPool.Stats()
		s.Metrics.ipPoolTotal.Set(float64(total))
		s.Metrics.ipPoolAvailable.Set(float64(available))
		if total > 0 {
			usagePercent := float64(total-available) / float64(total) * 100
			s.Metrics.ipPoolUsage.Set(usagePercent)
		}
		
		// Start metrics server
		go func() {
			if err := StartMetricsServer(s.Config.Metrics.ListenAddr); err != nil {
				log.Printf("Failed to start metrics server: %v", err)
			}
		}()
		
		// Update IP pool metrics periodically
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			defer ticker.Stop()
			for range ticker.C {
				s.IPPoolMu.RLock()
				total, allocated, available := s.IPPool.Stats()
				s.IPPoolMu.RUnlock()
				s.Metrics.ipPoolTotal.Set(float64(total))
				s.Metrics.ipPoolAvailable.Set(float64(available))
				if total > 0 {
					usagePercent := float64(allocated) / float64(total) * 100
					s.Metrics.ipPoolUsage.Set(usagePercent)
				}
			}
		}()
	}

	// --- CONNECT-IP 代理和 HTTP 处理程序 ---
	p := connectip.Proxy{}
	// 使用配置的服务器名称和端口作为模板
	serverHost, serverPortStr, _ := net.SplitHostPort(s.Config.ListenAddr)
	if serverHost == "0.0.0.0" || serverHost == "[::]" || serverHost == "" {
		serverHost = s.Config.ServerName // 如果监听的是通配符地址，使用配置的名称
	}
	serverPort, _ := strconv.Atoi(serverPortStr)
	template := uritemplate.MustNew(fmt.Sprintf("https://%s:%d/vpn", serverHost, serverPort))

	mux := http.NewServeMux()
	mux.HandleFunc("/vpn", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Incoming VPN request from %s", r.RemoteAddr)
		// 新增：提取 client_id（证书 CommonName）
		var clientID string
		if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
			clientID = r.TLS.PeerCertificates[0].Subject.CommonName
		} else {
			http.Error(w, "未检测到客户端证书", http.StatusUnauthorized)
			return
		}
		// 新增：校验 client_id 是否在数据库中
		clientIDExists := func(id string) bool {
			db, err := sql.Open("sqlite3", s.Config.APIServer.DatabasePath) // 使用配置的数据库路径
			if err != nil {
				log.Printf("Error opening database: %v", err) // Log the error
				return false
			}
			defer db.Close()
			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM clients WHERE client_id = ?", id).Scan(&count)
			if err != nil {
				log.Printf("Error querying database for client_id %s: %v", id, err) // Log the error
				return false
			}
			return count > 0
		}
		if !clientIDExists(clientID) {
			log.Printf("拒绝未知 client_id 连接: %s", clientID)
			http.Error(w, "客户端未授权或已被删除", http.StatusUnauthorized)
			return
		}

		req, err := connectip.ParseRequest(r, template)
		if err != nil {
			log.Printf("Failed to parse connect-ip request: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 建立服务器端 connect-ip 连接
		conn, err := p.Proxy(w, req)
		if err != nil {
			log.Printf("Failed to establish connect-ip proxy connection: %v", err)
			return
		}

		log.Printf("CONNECT-IP session established for %s", clientID)

		// Initialize FEC encoder if enabled
		var encoder *common_fec.XOREncoder
		fecEnabled := s.Config.FEC.Enabled
		if fecEnabled {
			enc, err := common_fec.NewXOREncoder(s.Config.FEC)
			if err != nil {
				log.Printf("Failed to create FEC encoder for client %s: %v", clientID, err)
				fecEnabled = false
			} else {
				encoder = enc
			}
		}

		// Create ClientSession
		session := &ClientSession{
			Conn:       conn,
			Encoder:    encoder,
			FecEnabled: fecEnabled,
			SeqNum:     0,
		}

		// 新增：为客户端分配唯一 IP
		s.IPPoolMu.Lock()
		assignedPrefix, allocErr := s.IPPool.Allocate(clientID)
		if allocErr != nil {
			s.IPPoolMu.Unlock()
			log.Printf("No available IP for client %s: %v", clientID, allocErr)
			conn.Close()
			return
		}
		s.ClientIPMap[clientID] = assignedPrefix.Addr()
		s.IPConnMap[assignedPrefix.Addr()] = session
		s.IPPoolMu.Unlock()
		log.Printf("Allocated IP %s to client %s", assignedPrefix, clientID)

		// 处理客户端连接，传递分配的 IP 和数据库路径
		go s.handleClientConnection(session, clientID, assignedPrefix, routesToAdvertise)
	})

	// 新增：API服务goroutine
	go func() {
		// 传递 serverConfig 给 API Server，并传递监听地址
		StartAPIServer(s)
	}()

	// --- HTTP/3 Server ---
	h3Server := http3.Server{
		Handler:         mux,
		EnableDatagrams: true,
		QUICConfig:      quicConf, // 使用与客户端相同的QUIC配置
	}

	// --- 优雅关闭处理 ---
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting HTTP/3 server...")
		if err := h3Server.ServeListener(ln); err != nil && err != http.ErrServerClosed && err != quic.ErrServerClosed {
			log.Printf("HTTP/3 server error: %v", err)
		}
		log.Println("HTTP/3 server stopped.")
	}()

	// 等待关闭信号
	<-ctx.Done()
	log.Println("Shutdown signal received...")

	// 初始化优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()

	if err := h3Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP/3 server shutdown ungracefully: %v", err)
	} else {
		log.Println("HTTP/3 server shutdown gracefully.")
	}

	// 等待服务器goroutine结束
	wg.Wait()
	log.Println("VPN Server exited.")
}

// processPackets handles reading from TUN and forwarding to clients
func (s *Server) processPackets() {
	buf := make([]byte, 2048+common.TunPacketOffset)
	for {
		n, err := s.TunDev.ReadPacket(buf, common.TunPacketOffset)
		if err != nil && err != os.ErrClosed {
			log.Printf("TUN read error: %v", err)
			continue
		}
		if n == 0 {
			continue
		}
		// Extract packet data (skip offset bytes)
		packet := make([]byte, n)
		copy(packet, buf[common.TunPacketOffset:common.TunPacketOffset+n])

		// 提取目标IP
		dstIP, err := common.GetDestinationIP(packet, n)
		if err != nil {
			log.Printf("Failed to parse destination IP: %v", err)
			continue
		}

		s.IPPoolMu.RLock()
		session, ok := s.IPConnMap[dstIP]
		s.IPPoolMu.RUnlock()
		if ok {
			session.Mu.Lock()
			if session.FecEnabled {
				// Add to buffer
				session.PacketBuffer = append(session.PacketBuffer, packet)
				
				// If buffer full, encode and send
				if len(session.PacketBuffer) >= session.Encoder.Config().BlockSize {
					if err := common.EncodeAndSendBlock(session.Encoder, session.Conn, session.PacketBuffer, &session.SeqNum); err != nil {
						log.Printf("Failed to send FEC block to client %s: %v", dstIP, err)
					}
					session.PacketBuffer = session.PacketBuffer[:0]
				}
			} else {
				_, err := session.Conn.WritePacket(packet)
				if err != nil {
					log.Printf("Failed to forward packet to client %s: %v", dstIP, err)
				}
			}
			session.Mu.Unlock()
		}
	}
}

// handleClientConnection 处理客户端VPN连接
func (s *Server) handleClientConnection(session *ClientSession, clientID string, assignedPrefix netip.Prefix, routes []connectip.IPRoute) {
	defer session.Conn.Close()

	log.Printf("Handling connection for client %s", clientID)
	
	// Initialize client metrics
	var clientMetrics *ClientMetrics
	if s.Metrics != nil {
		clientMetrics = NewClientMetrics(clientID)
		s.Metrics.activeConnections.Inc()
		s.Metrics.connectionsTotal.WithLabelValues("established").Inc()
		defer func() {
			s.Metrics.activeConnections.Dec()
			if clientMetrics != nil {
				clientMetrics.Close()
			}
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// --- 为客户端分配唯一 IP 前缀 ---
	if err := session.Conn.AssignAddresses(ctx, []netip.Prefix{assignedPrefix}); err != nil {
		log.Printf("Error assigning address %s to client %s: %v", assignedPrefix, clientID, err)
		if s.Metrics != nil {
			s.Metrics.connectionsTotal.WithLabelValues("failed").Inc()
			s.Metrics.errorsTotal.WithLabelValues("connection").Inc()
		}
		// 释放 IP
		s.IPPoolMu.Lock()
		s.IPPool.Release(assignedPrefix.Addr())
		delete(s.ClientIPMap, clientID)
		delete(s.IPConnMap, assignedPrefix.Addr())
		s.IPPoolMu.Unlock()
		return
	}
	log.Printf("Assigned IP %s to client %s", assignedPrefix, clientID)

	// --- 向客户端广播路由 ---
	if err := session.Conn.AdvertiseRoute(ctx, routes); err != nil {
		log.Printf("Error advertising routes to client %s: %v", clientID, err)
		if s.Metrics != nil {
			s.Metrics.connectionsTotal.WithLabelValues("failed").Inc()
			s.Metrics.errorsTotal.WithLabelValues("connection").Inc()
		}
		s.IPPoolMu.Lock()
		s.IPPool.Release(assignedPrefix.Addr())
		delete(s.ClientIPMap, clientID)
		delete(s.IPConnMap, assignedPrefix.Addr())
		s.IPPoolMu.Unlock()
		return
	}
	log.Printf("Advertised %d routes to client %s", len(routes), clientID)

	// --- 用户组与访问控制策略 ---
	// 修改：调用 api_server.go 中的 getGroupsAndPoliciesForClient，并传递 dbPath
	groupIDs, policies := getGroupsAndPoliciesForClient(s.Config.APIServer.DatabasePath, clientID)
	session.Conn.SetAccessControl(clientID, groupIDs, policies)

	// --- 只保留VPN->TUN方向 ---
	errChan := make(chan error, 1)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Pass &serverConfig.FEC to enable decoding
		common.ProxyFromVPNToTun(s.TunDev, session.Conn, errChan, &s.Config.FEC)
	}()

	err := <-errChan
	log.Printf("Proxying stopped for client %s: %v", clientID, err)
	session.Conn.Close()

	wg.Wait()
	log.Printf("Finished handling client %s", clientID)

	// 连接结束时释放 IP
	s.IPPoolMu.Lock()
	// 修复：只有当 clientID 还在 clientIPMap 时才释放，避免被 API 删除后重复释放
	if ip, ok := s.ClientIPMap[clientID]; ok && ip == assignedPrefix.Addr() {
		s.IPPool.Release(assignedPrefix.Addr())
		delete(s.ClientIPMap, clientID)
		delete(s.IPConnMap, assignedPrefix.Addr())
	}
	s.IPPoolMu.Unlock()
}

// getGroupsAndPoliciesForClient retrieves ACLs from the database
func getGroupsAndPoliciesForClient(dbPath string, clientID string) ([]string, []connectip.AccessPolicy) {
	groupIDs := []string{}
	policies := []connectip.AccessPolicy{}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("[ACL] 打开数据库失败: %v", err)
		return groupIDs, policies
	}
	defer db.Close()
	// 查询groupIDs
	rows, err := db.Query("SELECT group_id FROM group_members WHERE client_id = ?", clientID)
	if err == nil {
		for rows.Next() {
			var gid string
			if err := rows.Scan(&gid); err == nil {
				groupIDs = append(groupIDs, gid)
			}
		}
		rows.Close()
	}
	if len(groupIDs) == 0 {
		return groupIDs, policies
	}
	// 查询所有相关策略
	query, args := "SELECT action, ip_prefix, priority FROM access_policies WHERE group_id IN ("+strings.Repeat("?,", len(groupIDs)-1)+"?) ORDER BY priority ASC", make([]interface{}, len(groupIDs))
	for i, gid := range groupIDs {
		args[i] = gid
	}
	rows, err = db.Query(query, args...)
	if err == nil {
		for rows.Next() {
			var action, ipPrefix string
			var priority int
			if err := rows.Scan(&action, &ipPrefix, &priority); err == nil {
				if prefix, err := netip.ParsePrefix(ipPrefix); err == nil {
					policies = append(policies, connectip.AccessPolicy{
						Action:   action,
						IPPrefix: prefix,
						Priority: priority,
					})
				}
			}
		}
		rows.Close()
	}
	return groupIDs, policies
}
