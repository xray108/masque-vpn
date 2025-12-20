package server

import (
	"log"
	"net"
	"net/netip"
)

// processPackets обрабатывает пакеты из TUN устройства
func (s *Server) processPackets() {
	buffer := make([]byte, 2048)
	
	log.Printf("Starting packet processor for TUN device %s", s.TunDev.Name())
	
	for {
		// Читаем пакет из TUN устройства
		n, err := s.TunDev.ReadPacket(buffer, 0)
		if err != nil {
			if isNetworkClosed(err) {
				log.Printf("TUN device closed, stopping packet processor")
				return
			}
			log.Printf("Error reading from TUN device: %v", err)
			continue
		}

		if n == 0 {
			continue
		}

		packetData := buffer[:n]
		
		// Парсим IP пакет для определения назначения
		destIP, err := s.parseDestinationIP(packetData)
		if err != nil {
			log.Printf("Failed to parse packet destination: %v", err)
			continue
		}

		// Находим клиентскую сессию для этого IP
		session := s.findClientSession(destIP)
		if session == nil {
			log.Printf("No client session found for destination IP %s", destIP)
			continue
		}

		// Отправляем пакет клиенту
		if err := s.forwardPacketToClient(session, packetData); err != nil {
			log.Printf("Failed to forward packet to client: %v", err)
			continue
		}

		// Обновляем метрики
		s.Metrics.PacketsForwarded.Inc()
		s.Metrics.BytesForwarded.Add(float64(n))
	}
}

// parseDestinationIP извлекает IP назначения из пакета
func (s *Server) parseDestinationIP(packet []byte) (netip.Addr, error) {
	if len(packet) < 20 {
		return netip.Addr{}, net.ErrWriteToConnected
	}

	// Проверяем версию IP
	version := packet[0] >> 4
	
	switch version {
	case 4:
		// IPv4 - адрес назначения в байтах 16-19
		if len(packet) < 20 {
			return netip.Addr{}, net.ErrWriteToConnected
		}
		destBytes := packet[16:20]
		addr := netip.AddrFrom4([4]byte{destBytes[0], destBytes[1], destBytes[2], destBytes[3]})
		return addr, nil
		
	case 6:
		// IPv6 - адрес назначения в байтах 24-39
		if len(packet) < 40 {
			return netip.Addr{}, net.ErrWriteToConnected
		}
		destBytes := packet[24:40]
		var addr16 [16]byte
		copy(addr16[:], destBytes)
		addr := netip.AddrFrom16(addr16)
		return addr, nil
		
	default:
		return netip.Addr{}, net.ErrWriteToConnected
	}
}

// findClientSession находит клиентскую сессию по IP адресу
func (s *Server) findClientSession(ip netip.Addr) *ClientSession {
	s.IPPoolMu.RLock()
	defer s.IPPoolMu.RUnlock()
	
	session, exists := s.IPConnMap[ip]
	if !exists {
		return nil
	}
	
	return session
}

// forwardPacketToClient отправляет пакет клиенту
func (s *Server) forwardPacketToClient(session *ClientSession, packet []byte) error {
	if session.Conn == nil {
		return net.ErrClosed
	}

	// TODO: Добавить поддержку FEC если включена
	if session.FecEnabled {
		return s.forwardPacketWithFEC(session, packet)
	}

	// Простая отправка без FEC
	return session.Conn.WritePacket(packet)
}

// forwardPacketWithFEC отправляет пакет с FEC кодированием
func (s *Server) forwardPacketWithFEC(session *ClientSession, packet []byte) error {
	session.Mu.Lock()
	defer session.Mu.Unlock()

	// Добавляем пакет в буфер
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)
	session.PacketBuffer = append(session.PacketBuffer, packetCopy)

	// Если буфер заполнен, кодируем и отправляем блок
	if session.Encoder != nil && len(session.PacketBuffer) >= session.Encoder.Config().BlockSize {
		return s.encodeAndSendBlock(session)
	}

	return nil
}

// encodeAndSendBlock кодирует и отправляет блок пакетов с FEC
func (s *Server) encodeAndSendBlock(session *ClientSession) error {
	if session.Encoder == nil {
		return net.ErrClosed
	}

	// Кодируем блок
	encoded, err := session.Encoder.Encode(session.PacketBuffer)
	if err != nil {
		return err
	}

	// Отправляем все пакеты блока
	for _, pkt := range encoded {
		if err := session.Conn.WritePacket(pkt); err != nil {
			return err
		}
		session.SeqNum++
	}

	// Очищаем буфер
	session.PacketBuffer = session.PacketBuffer[:0]
	
	return nil
}

// isNetworkClosed проверяет, указывает ли ошибка на закрытое соединение
func isNetworkClosed(err error) bool {
	if err == nil {
		return false
	}
	
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Err.Error() == "use of closed network connection"
	}
	
	errStr := err.Error()
	return errStr == "EOF" || 
		   errStr == "connection reset by peer" ||
		   errStr == "use of closed network connection"
}