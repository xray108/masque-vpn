package common

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"go.uber.org/zap"
)

// MASQUEClient represents a MASQUE CONNECT-IP client
type MASQUEClient struct {
	quicConn   *quic.Conn
	httpClient *http.Client
	logger     *zap.Logger
	mu         sync.RWMutex
	closed     bool
}

// MASQUEConn represents a MASQUE CONNECT-IP connection for IP packet tunneling
type MASQUEConn struct {
	Stream     *quic.Stream
	client     *MASQUEClient
	Logger     *zap.Logger
	mu         sync.RWMutex
	closed     bool
	// Для тестирования добавляем каналы
	readChan   chan []byte
	writeChan  chan []byte
}

// NewMASQUEClient creates a new MASQUE client
func NewMASQUEClient(quicConn *quic.Conn, logger *zap.Logger) *MASQUEClient {
	// Create HTTP/3 client for MASQUE requests
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &MASQUEClient{
		quicConn:   quicConn,
		httpClient: httpClient,
		logger:     logger,
	}
}

// ConnectIP establishes a CONNECT-IP session for IP packet tunneling
func (c *MASQUEClient) ConnectIP(ctx context.Context) (*MASQUEConn, error) {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return nil, fmt.Errorf("client is closed")
	}
	c.mu.RUnlock()

	if c.quicConn == nil {
		return nil, fmt.Errorf("QUIC connection is nil")
	}

	// Create HTTP CONNECT request for MASQUE
	serverAddr := c.quicConn.RemoteAddr().String()
	
	// Create request with proper MASQUE headers
	req, err := http.NewRequestWithContext(ctx, http.MethodConnect, serverAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create CONNECT request: %w", err)
	}

	// Add MASQUE-specific headers
	req.Header.Set("Capsule-Protocol", "?masque")
	req.Header.Set("Upgrade", "masque")
	req.Header.Set("Connection", "Upgrade")
	
	c.logger.Info("Sending MASQUE CONNECT request",
		zap.String("server_addr", serverAddr),
		zap.String("method", req.Method))

	// For now, we'll use direct QUIC stream approach
	// In a full implementation, we would use the HTTP client
	stream, err := c.quicConn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open QUIC stream: %w", err)
	}

	c.logger.Info("Opened QUIC stream for MASQUE CONNECT-IP session",
		zap.Uint64("stream_id", uint64(stream.StreamID())))

	// Send a simplified CONNECT request over the stream
	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\n", serverAddr) +
		"Capsule-Protocol: ?masque\r\n" +
		"Upgrade: masque\r\n" +
		"Connection: Upgrade\r\n" +
		"\r\n"

	if _, err := stream.Write([]byte(connectReq)); err != nil {
		stream.Close()
		return nil, fmt.Errorf("failed to send CONNECT request: %w", err)
	}

	// Read response
	respBuf := make([]byte, 1024)
	n, err := stream.Read(respBuf)
	if err != nil {
		stream.Close()
		return nil, fmt.Errorf("failed to read CONNECT response: %w", err)
	}

	response := string(respBuf[:n])
	c.logger.Debug("MASQUE CONNECT response", zap.String("response", response))

	// Check for successful response (HTTP 200)
	if !strings.Contains(response, "200") && !strings.Contains(response, "OK") {
		stream.Close()
		return nil, fmt.Errorf("MASQUE CONNECT request failed: %s", response)
	}

	c.logger.Info("MASQUE CONNECT-IP session established successfully")
	
	return &MASQUEConn{
		Stream:    stream,
		client:    c,
		Logger:    c.logger,
		readChan:  make(chan []byte, 100),
		writeChan: make(chan []byte, 100),
	}, nil
}

// NewMASQUEConnForServer creates a new MASQUE connection for server side
func NewMASQUEConnForServer(logger *zap.Logger) *MASQUEConn {
	// Для тестирования создаем связанные каналы
	readChan := make(chan []byte, 100)
	writeChan := make(chan []byte, 100)
	
	conn := &MASQUEConn{
		Stream:    nil,
		client:    nil,
		Logger:    logger,
		readChan:  readChan,
		writeChan: writeChan,
	}
	
	// Запускаем горутину для связывания каналов (для тестирования)
	go func() {
		for packet := range writeChan {
			select {
			case readChan <- packet:
			default:
				// Канал заполнен, пропускаем пакет
			}
		}
	}()
	
	return conn
}

// Close closes the MASQUE client
func (c *MASQUEClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	if c.quicConn == nil {
		return fmt.Errorf("QUIC connection is nil")
	}

	return c.quicConn.CloseWithError(0, "client shutdown")
}

// ReadPacket reads an IP packet from the MASQUE connection
func (m *MASQUEConn) ReadPacket(buf []byte) (int, error) {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return 0, io.EOF
	}
	m.mu.RUnlock()

	// Если есть stream, используем его
	if m.Stream != nil {
		// Set read deadline to prevent blocking indefinitely
		if err := m.Stream.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return 0, fmt.Errorf("failed to set read deadline: %w", err)
		}

		n, err := m.Stream.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return 0, fmt.Errorf("read timeout: %w", err)
			}
			return 0, fmt.Errorf("failed to read from MASQUE stream: %w", err)
		}
		return n, nil
	}

	// Иначе используем канал для тестирования
	select {
	case packet := <-m.readChan:
		n := copy(buf, packet)
		return n, nil
	case <-time.After(30 * time.Second):
		return 0, fmt.Errorf("read timeout")
	}
}

// WritePacket writes an IP packet to the MASQUE connection
func (m *MASQUEConn) WritePacket(packet []byte) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("connection is closed")
	}
	m.mu.RUnlock()

	// Если есть stream, используем его
	if m.Stream != nil {
		// Set write deadline to prevent blocking indefinitely
		if err := m.Stream.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}

		_, err := m.Stream.Write(packet)
		if err != nil {
			return fmt.Errorf("failed to write to MASQUE stream: %w", err)
		}
		return nil
	}

	// Иначе используем канал для тестирования
	packetCopy := make([]byte, len(packet))
	copy(packetCopy, packet)
	
	select {
	case m.writeChan <- packetCopy:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("write timeout")
	}
}

// Close closes the MASQUE connection
func (m *MASQUEConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	// Закрываем каналы
	if m.readChan != nil {
		close(m.readChan)
	}
	if m.writeChan != nil {
		close(m.writeChan)
	}

	// Закрываем stream если есть
	if m.Stream != nil {
		return m.Stream.Close()
	}
	return nil
}

// LocalAddr returns the local address (not applicable for MASQUE)
func (m *MASQUEConn) LocalAddr() net.Addr {
	return m.client.quicConn.LocalAddr()
}

// RemoteAddr returns the remote address
func (m *MASQUEConn) RemoteAddr() net.Addr {
	return m.client.quicConn.RemoteAddr()
}