package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"

	common "github.com/iselt/masque-vpn/common"
)

// handleMASQUERequest обрабатывает MASQUE CONNECT-IP запросы
func (s *Server) handleMASQUERequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// MASQUE CONNECT-IP использует обычный HTTP CONNECT метод с специальными заголовками
	if r.Method != http.MethodConnect {
		// Для других методов возвращаем информацию о сервере
		if r.Method == http.MethodGet && r.URL.Path == "/" {
			s.handleServerInfo(w, r)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Проверяем заголовки MASQUE
	if !s.isMASQUERequest(r) {
		http.Error(w, "Not a MASQUE request", http.StatusBadRequest)
		return
	}

	// Получаем клиентский сертификат для аутентификации
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "Client certificate required", http.StatusUnauthorized)
		return
	}

	clientCert := r.TLS.PeerCertificates[0]
	clientID := clientCert.Subject.CommonName
	if clientID == "" {
		http.Error(w, "Invalid client certificate", http.StatusUnauthorized)
		return
	}

	log.Printf("Client authenticated: %s", clientID)

	// Выделяем IP адрес для клиента
	assignedPrefix, err := s.assignIPToClient(clientID)
	if err != nil {
		log.Printf("Failed to assign IP to client %s: %v", clientID, err)
		http.Error(w, "Failed to assign IP", http.StatusInternalServerError)
		return
	}

	log.Printf("Assigned IP %s to client %s", assignedPrefix, clientID)

	// Отправляем успешный ответ CONNECT
	w.Header().Set("Content-Type", "application/masque")
	w.WriteHeader(http.StatusOK)

	// Для HTTP/3 hijacking нужно использовать другой подход
	// Пока используем упрощенную реализацию без hijacking
	log.Printf("MASQUE CONNECT request accepted for client %s", clientID)

	// Создаем MASQUE соединение без прямого доступа к stream
	// В реальной реализации здесь должен быть HTTP/3 hijacking
	masqueConn := common.NewMASQUEConnForServer(nil)

	// Создаем сессию клиента
	session := &ClientSession{
		Conn:       masqueConn,
		FecEnabled: s.Config.FEC.Enabled,
	}

	// Сохраняем сессию
	s.IPPoolMu.Lock()
	s.IPConnMap[assignedPrefix.Addr()] = session
	s.IPPoolMu.Unlock()

	// Обновляем метрики
	s.Metrics.RecordConnection()
	defer s.Metrics.RecordDisconnection()

	// Запускаем обработку соединения
	s.handleClientConnection(session, clientID, assignedPrefix.Addr(), nil)
}

// isMASQUERequest проверяет, является ли запрос MASQUE CONNECT-IP
func (s *Server) isMASQUERequest(r *http.Request) bool {
	// Проверяем заголовки, специфичные для MASQUE
	capsuleProtocol := r.Header.Get("Capsule-Protocol")
	return capsuleProtocol == "?masque" || strings.Contains(r.Header.Get("Upgrade"), "masque")
}

// handleServerInfo возвращает информацию о сервере
func (s *Server) handleServerInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := fmt.Sprintf(`{
		"service": "masque-vpn-server",
		"version": "1.0.0",
		"protocol": "MASQUE CONNECT-IP",
		"network": "%s"
	}`, s.Config.AssignCIDR)
	w.Write([]byte(response))
}

// assignIPToClient выделяет IP адрес клиенту
func (s *Server) assignIPToClient(clientID string) (netip.Prefix, error) {
	s.IPPoolMu.Lock()
	defer s.IPPoolMu.Unlock()

	// Проверяем, есть ли уже назначенный IP
	if existingIP, exists := s.ClientIPMap[clientID]; exists {
		return netip.PrefixFrom(existingIP, 32), nil
	}

	// Выделяем новый IP
	assignedPrefix, err := s.IPPool.Allocate(clientID)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("failed to allocate IP: %w", err)
	}

	s.ClientIPMap[clientID] = assignedPrefix.Addr()
	return assignedPrefix, nil
}

// handleClientConnection обрабатывает соединение с клиентом
func (s *Server) handleClientConnection(session *ClientSession, clientID string, assignedIP netip.Addr, stream interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in client connection handler for %s: %v", clientID, r)
			s.Metrics.RecordError("panic")
		}
	}()

	log.Printf("Starting connection handler for client %s (IP: %s)", clientID, assignedIP)
	connectionStart := time.Now()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	// Запускаем прокси-горутины
	errChan := make(chan error, 2)
	
	go func() {
		// TUN -> Client proxy
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in TUN->Client proxy: %v", r)
			}
		}()
		
		log.Printf("TUN->Client proxy started for %s", clientID)
		
		// Реализуем прокси от TUN к клиенту
		if err := s.proxyTunToClient(ctx, session, assignedIP); err != nil {
			errChan <- fmt.Errorf("TUN->Client proxy error: %w", err)
			return
		}
		
		errChan <- nil
	}()

	go func() {
		// Client -> TUN proxy
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in Client->TUN proxy: %v", r)
			}
		}()
		
		log.Printf("Client->TUN proxy started for %s", clientID)
		
		// Реализуем прокси от клиента к TUN
		if err := s.proxyClientToTun(ctx, session, assignedIP); err != nil {
			errChan <- fmt.Errorf("Client->TUN proxy error: %w", err)
			return
		}
		
		errChan <- nil
	}()

	// Ждем ошибку или завершения
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("Proxy error for client %s: %v", clientID, err)
			s.Metrics.RecordError("proxy_error")
		}
	case <-ctx.Done():
		log.Printf("Connection timeout for client %s", clientID)
		s.Metrics.RecordError("timeout")
	}

	// Записываем продолжительность соединения
	duration := time.Since(connectionStart).Seconds()
	s.Metrics.RecordConnectionDuration(duration)

	log.Printf("Connection handler finished for client %s (duration: %.2fs)", clientID, duration)
	
	// Очищаем ресурсы
	s.cleanupClientSession(clientID, assignedIP)
}

// proxyTunToClient проксирует пакеты от TUN устройства к клиенту
func (s *Server) proxyTunToClient(ctx context.Context, session *ClientSession, clientIP netip.Addr) error {
	log.Printf("Starting TUN->Client proxy for IP %s", clientIP)
	
	// Проверяем наличие TUN устройства
	if s.TunDev == nil {
		log.Printf("TUN device not available, TUN->Client proxy disabled for IP %s", clientIP)
		<-ctx.Done()
		return nil
	}
	
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("TUN->Client proxy stopped for IP %s (context cancelled)", clientIP)
			return nil
		default:
		}
		
		// Читаем пакет из TUN устройства с таймаутом
		n, err := s.TunDev.ReadPacket(buffer, 100) // 100ms timeout
		if err != nil {
			if isNetworkClosed(err) {
				log.Printf("TUN device closed, stopping TUN->Client proxy for IP %s", clientIP)
				return nil
			}
			// Игнорируем таймауты и продолжаем
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("failed to read from TUN device: %w", err)
		}

		if n == 0 {
			continue
		}

		packetData := buffer[:n]
		
		// Парсим IP пакет для проверки назначения
		destIP, err := s.parseDestinationIP(packetData)
		if err != nil {
			continue // Пропускаем некорректные пакеты
		}
		
		// Проверяем, что пакет предназначен для этого клиента
		if destIP != clientIP {
			continue // Пакет не для этого клиента
		}

		// Отправляем пакет клиенту через MASQUE соединение
		if err := session.Conn.WritePacket(packetData); err != nil {
			if isNetworkClosed(err) {
				log.Printf("MASQUE connection closed, stopping TUN->Client proxy for IP %s", clientIP)
				return nil
			}
			return fmt.Errorf("failed to write packet to MASQUE connection: %w", err)
		}
		
		// Обновляем метрики
		s.Metrics.PacketsForwarded.Inc()
		s.Metrics.BytesForwarded.Add(float64(n))
	}
}

// proxyClientToTun проксирует пакеты от клиента к TUN устройству
func (s *Server) proxyClientToTun(ctx context.Context, session *ClientSession, clientIP netip.Addr) error {
	log.Printf("Starting Client->TUN proxy for IP %s", clientIP)
	
	// Проверяем наличие TUN устройства
	if s.TunDev == nil {
		log.Printf("TUN device not available, Client->TUN proxy disabled for IP %s", clientIP)
		<-ctx.Done()
		return nil
	}
	
	buffer := make([]byte, 2048)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Client->TUN proxy stopped for IP %s (context cancelled)", clientIP)
			return nil
		default:
		}
		
		// Читаем пакет от клиента через MASQUE соединение
		n, err := session.Conn.ReadPacket(buffer)
		if err != nil {
			if isNetworkClosed(err) {
				log.Printf("MASQUE connection closed, stopping Client->TUN proxy for IP %s", clientIP)
				return nil
			}
			// Игнорируем таймауты и продолжаем
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("failed to read from MASQUE connection: %w", err)
		}

		if n == 0 {
			continue
		}

		packetData := buffer[:n]
		
		// Парсим IP пакет для проверки источника
		srcIP, err := s.parseSourceIP(packetData)
		if err != nil {
			continue // Пропускаем некорректные пакеты
		}
		
		// Проверяем, что пакет от правильного клиента
		if srcIP != clientIP {
			log.Printf("Packet from wrong source IP %s, expected %s", srcIP, clientIP)
			continue
		}

		// Отправляем пакет в TUN устройство
		if err := s.TunDev.WritePacket(packetData, 0); err != nil {
			if isNetworkClosed(err) {
				log.Printf("TUN device closed, stopping Client->TUN proxy for IP %s", clientIP)
				return nil
			}
			return fmt.Errorf("failed to write packet to TUN device: %w", err)
		}
		
		// Обновляем метрики
		s.Metrics.TunPacketsWritten.Inc()
		s.Metrics.BytesForwarded.Add(float64(n))
	}
}

// cleanupClientSession очищает ресурсы клиентской сессии
func (s *Server) cleanupClientSession(clientID string, assignedIP netip.Addr) {
	s.IPPoolMu.Lock()
	defer s.IPPoolMu.Unlock()

	// Удаляем из карт
	delete(s.ClientIPMap, clientID)
	if session, exists := s.IPConnMap[assignedIP]; exists {
		if session.Conn != nil {
			session.Conn.Close()
		}
		delete(s.IPConnMap, assignedIP)
	}

	// Освобождаем IP
	s.IPPool.Release(assignedIP)

	log.Printf("Cleaned up session for client %s (IP: %s)", clientID, assignedIP)
}