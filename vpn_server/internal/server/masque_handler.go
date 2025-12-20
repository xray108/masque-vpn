package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"time"

	common "github.com/iselt/masque-vpn/common"
)

// handleMASQUERequest обрабатывает MASQUE CONNECT-IP запросы
func (s *Server) handleMASQUERequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received MASQUE request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

	// Проверяем, что это CONNECT-IP запрос
	if r.Method != "CONNECT-IP" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/connect-ip")
	w.WriteHeader(http.StatusOK)

	// TODO: Получить HTTP/3 stream для туннелирования
	// Пока что создаем заглушку для MASQUE соединения
	masqueConn := &common.MASQUEConn{
		// TODO: Инициализировать с реальным stream
	}

	// Создаем сессию клиента
	session := &ClientSession{
		Conn:       masqueConn,
		FecEnabled: s.Config.FEC.Enabled,
	}

	// Сохраняем сессию
	s.IPPoolMu.Lock()
	s.IPConnMap[assignedPrefix.Addr()] = session
	s.IPPoolMu.Unlock()

	// Запускаем обработку соединения
	go s.handleClientConnection(session, clientID, assignedPrefix.Addr())

	// Блокируем до закрытия соединения
	<-r.Context().Done()
	
	// Очищаем ресурсы
	s.cleanupClientSession(clientID, assignedPrefix.Addr())
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
func (s *Server) handleClientConnection(session *ClientSession, clientID string, assignedIP netip.Addr) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in client connection handler for %s: %v", clientID, r)
		}
	}()

	log.Printf("Starting connection handler for client %s (IP: %s)", clientID, assignedIP)

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	// Запускаем прокси-горутины
	errChan := make(chan error, 2)
	
	go func() {
		// TUN -> Client
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in TUN->Client proxy: %v", r)
			}
		}()
		
		// TODO: Реализовать прокси от TUN к клиенту
		log.Printf("TUN->Client proxy started for %s", clientID)
		<-ctx.Done()
		errChan <- nil
	}()

	go func() {
		// Client -> TUN
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("panic in Client->TUN proxy: %v", r)
			}
		}()
		
		// TODO: Реализовать прокси от клиента к TUN
		log.Printf("Client->TUN proxy started for %s", clientID)
		<-ctx.Done()
		errChan <- nil
	}()

	// Ждем ошибку или завершения
	select {
	case err := <-errChan:
		if err != nil {
			log.Printf("Proxy error for client %s: %v", clientID, err)
		}
	case <-ctx.Done():
		log.Printf("Connection timeout for client %s", clientID)
	}

	log.Printf("Connection handler finished for client %s", clientID)
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