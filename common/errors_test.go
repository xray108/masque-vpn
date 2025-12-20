package common

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVPNError(t *testing.T) {
	cause := errors.New("underlying error")
	vpnErr := NewVPNError("connection", "failed to connect", cause)
	
	assert.Equal(t, "connection", vpnErr.Type)
	assert.Equal(t, "failed to connect", vpnErr.Message)
	assert.Equal(t, cause, vpnErr.Cause)
	assert.NotNil(t, vpnErr.Context)
	
	// Test Error() method
	errStr := vpnErr.Error()
	assert.Contains(t, errStr, "connection")
	assert.Contains(t, errStr, "failed to connect")
	assert.Contains(t, errStr, "underlying error")
	
	// Test Unwrap() method
	assert.Equal(t, cause, vpnErr.Unwrap())
}

func TestVPNErrorWithContext(t *testing.T) {
	vpnErr := NewVPNError("network", "route failed", nil)
	vpnErr.WithContext("interface", "tun0")
	vpnErr.WithContext("ip", "10.0.0.1")
	
	assert.Equal(t, "tun0", vpnErr.Context["interface"])
	assert.Equal(t, "10.0.0.1", vpnErr.Context["ip"])
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "VPN connection error",
			err:      NewVPNError("connection", "failed", nil),
			expected: true,
		},
		{
			name:     "Network timeout error",
			err:      &net.OpError{Err: &timeoutError{}},
			expected: true,
		},
		{
			name:     "Standard connection error",
			err:      ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "Non-connection error",
			err:      NewVPNError("configuration", "invalid", nil),
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConnectionError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsConfigurationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "VPN configuration error",
			err:      NewVPNError("configuration", "invalid config", nil),
			expected: true,
		},
		{
			name:     "Standard config error",
			err:      ErrInvalidConfig,
			expected: true,
		},
		{
			name:     "Missing config error",
			err:      ErrMissingConfig,
			expected: true,
		},
		{
			name:     "Non-configuration error",
			err:      NewVPNError("network", "failed", nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsConfigurationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRecoveryStrategy(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected RecoveryStrategy
	}{
		{
			name:     "Connection error",
			err:      ErrConnectionFailed,
			expected: RecoveryReconnect,
		},
		{
			name:     "Network error",
			err:      ErrTUNDeviceCreation,
			expected: RecoveryRetry,
		},
		{
			name:     "Protocol error",
			err:      ErrMASQUEProtocol,
			expected: RecoveryReconnect,
		},
		{
			name:     "Configuration error",
			err:      ErrInvalidConfig,
			expected: RecoveryNone,
		},
		{
			name:     "System error",
			err:      ErrPermissionDenied,
			expected: RecoveryRestart,
		},
		{
			name:     "Unknown error",
			err:      errors.New("unknown"),
			expected: RecoveryRetry,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := GetRecoveryStrategy(tt.err)
			assert.Equal(t, tt.expected, strategy)
		})
	}
}

// Mock timeout error for testing
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }