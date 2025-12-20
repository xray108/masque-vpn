package common

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
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
	stream   quic.Stream
	client   *MASQUEClient
	logger   *zap.Logger
	mu       sync.RWMutex
	closed   bool
}

// NewMASQUEClient creates a new MASQUE client
func NewMASQUEClient(quicConn *quic.Conn, logger *zap.Logger) *MASQUEClient {
	// Create HTTP/3 client using the existing QUIC connection
	roundTripper := &http3.RoundTripper{
		// We'll use the existing connection, so no need to dial
	}
	
	httpClient := &http.Client{
		Transport: roundTripper,
		Timeout:   30 * time.Second,
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

	// For MASQUE CONNECT-IP, we need to use HTTP datagrams over QUIC
	// This is a simplified implementation - in a full implementation,
	// we would use proper HTTP/3 CONNECT-IP method
	
	// Open a bidirectional stream for the CONNECT-IP session
	stream, err := c.quicConn.OpenStreamSync(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open QUIC stream: %w", err)
	}

	c.logger.Info("Opened QUIC stream for MASQUE CONNECT-IP session",
		zap.Uint64("stream_id", uint64(stream.StreamID())))

	// In a real implementation, we would send proper HTTP/3 CONNECT-IP request
	// For now, we'll use the stream directly for IP packet tunneling
	
	return &MASQUEConn{
		stream: stream,
		client: c,
		logger: c.logger,
	}, nil
}

// Close closes the MASQUE client
func (c *MASQUEClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

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

	// Set read deadline to prevent blocking indefinitely
	if err := m.stream.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return 0, fmt.Errorf("failed to set read deadline: %w", err)
	}

	n, err := m.stream.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return 0, fmt.Errorf("read timeout: %w", err)
		}
		return 0, fmt.Errorf("failed to read from MASQUE stream: %w", err)
	}

	return n, nil
}

// WritePacket writes an IP packet to the MASQUE connection
func (m *MASQUEConn) WritePacket(packet []byte) error {
	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return fmt.Errorf("connection is closed")
	}
	m.mu.RUnlock()

	// Set write deadline to prevent blocking indefinitely
	if err := m.stream.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return fmt.Errorf("failed to set write deadline: %w", err)
	}

	_, err := m.stream.Write(packet)
	if err != nil {
		return fmt.Errorf("failed to write to MASQUE stream: %w", err)
	}

	return nil
}

// Close closes the MASQUE connection
func (m *MASQUEConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return nil
	}
	m.closed = true

	return m.stream.Close()
}

// LocalAddr returns the local address (not applicable for MASQUE)
func (m *MASQUEConn) LocalAddr() net.Addr {
	return m.client.quicConn.LocalAddr()
}

// RemoteAddr returns the remote address
func (m *MASQUEConn) RemoteAddr() net.Addr {
	return m.client.quicConn.RemoteAddr()
}