package common

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestMASQUEClient_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	
	// Test client creation with nil connection (should not panic)
	client := NewMASQUEClient(nil, logger)
	assert.NotNil(t, client)
	assert.Equal(t, logger, client.logger)
}

func TestMASQUEClient_Close(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMASQUEClient(nil, logger)
	
	// Test closing client with nil connection
	err := client.Close()
	// Should return error because quicConn is nil, but should not panic
	assert.Error(t, err)
	
	// Test double close
	err = client.Close()
	assert.NoError(t, err) // Should be no-op
}

func TestMASQUEClient_ConnectIP_WithoutConnection(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMASQUEClient(nil, logger)
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Should fail because no real QUIC connection
	conn, err := client.ConnectIP(ctx)
	assert.Error(t, err)
	assert.Nil(t, conn)
}

func TestMASQUEClient_ConnectIP_AfterClose(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMASQUEClient(nil, logger)
	
	// Close client first
	client.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	
	// Should fail because client is closed
	conn, err := client.ConnectIP(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is closed")
	assert.Nil(t, conn)
}

func TestMASQUEConn_Methods(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client := NewMASQUEClient(nil, logger)
	
	// Create a MASQUE connection with nil stream (for testing)
	conn := &MASQUEConn{
		Stream: nil,
		client: client,
		Logger: logger,
	}
	
	// Test ReadPacket with closed connection
	conn.closed = true
	n, err := conn.ReadPacket(make([]byte, 100))
	assert.Equal(t, 0, n)
	assert.Error(t, err)
	
	// Test WritePacket with closed connection
	err = conn.WritePacket([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection is closed")
	
	// Test Close
	err = conn.Close()
	assert.NoError(t, err) // Should be no-op since already closed
}