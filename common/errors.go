package common

import (
	"errors"
	"fmt"
	"net"
)

// Error types for better error handling and metrics
var (
	// Connection errors
	ErrConnectionFailed    = errors.New("connection failed")
	ErrConnectionTimeout   = errors.New("connection timeout")
	ErrConnectionLost      = errors.New("connection lost")
	ErrAuthenticationFailed = errors.New("authentication failed")
	
	// Configuration errors
	ErrInvalidConfig       = errors.New("invalid configuration")
	ErrMissingConfig       = errors.New("missing required configuration")
	ErrInvalidCertificate  = errors.New("invalid certificate")
	
	// Network errors
	ErrTUNDeviceCreation   = errors.New("TUN device creation failed")
	ErrRouteAddition       = errors.New("route addition failed")
	ErrIPAllocation        = errors.New("IP allocation failed")
	ErrNetworkUnreachable  = errors.New("network unreachable")
	
	// Protocol errors
	ErrMASQUEProtocol      = errors.New("MASQUE protocol error")
	ErrQUICProtocol        = errors.New("QUIC protocol error")
	ErrHTTP3Protocol       = errors.New("HTTP/3 protocol error")
	
	// System errors
	ErrPermissionDenied    = errors.New("permission denied")
	ErrResourceExhausted   = errors.New("resource exhausted")
	ErrSystemCall          = errors.New("system call failed")
)

// VPNError represents a structured error with context
type VPNError struct {
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Cause     error             `json:"cause,omitempty"`
	Context   map[string]string `json:"context,omitempty"`
	Timestamp int64             `json:"timestamp"`
}

func (e *VPNError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *VPNError) Unwrap() error {
	return e.Cause
}

// NewVPNError creates a new structured VPN error
func NewVPNError(errorType, message string, cause error) *VPNError {
	return &VPNError{
		Type:      errorType,
		Message:   message,
		Cause:     cause,
		Context:   make(map[string]string),
		Timestamp: getCurrentTimestamp(),
	}
}

// WithContext adds context information to the error
func (e *VPNError) WithContext(key, value string) *VPNError {
	if e.Context == nil {
		e.Context = make(map[string]string)
	}
	e.Context[key] = value
	return e
}

// Error classification functions
func IsConnectionError(err error) bool {
	var vpnErr *VPNError
	if errors.As(err, &vpnErr) {
		return vpnErr.Type == "connection"
	}
	
	// Check for common network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || !netErr.Temporary()
	}
	
	return errors.Is(err, ErrConnectionFailed) ||
		   errors.Is(err, ErrConnectionTimeout) ||
		   errors.Is(err, ErrConnectionLost)
}

func IsConfigurationError(err error) bool {
	var vpnErr *VPNError
	if errors.As(err, &vpnErr) {
		return vpnErr.Type == "configuration"
	}
	
	return errors.Is(err, ErrInvalidConfig) ||
		   errors.Is(err, ErrMissingConfig) ||
		   errors.Is(err, ErrInvalidCertificate)
}

func IsNetworkError(err error) bool {
	var vpnErr *VPNError
	if errors.As(err, &vpnErr) {
		return vpnErr.Type == "network"
	}
	
	return errors.Is(err, ErrTUNDeviceCreation) ||
		   errors.Is(err, ErrRouteAddition) ||
		   errors.Is(err, ErrIPAllocation) ||
		   errors.Is(err, ErrNetworkUnreachable)
}

func IsProtocolError(err error) bool {
	var vpnErr *VPNError
	if errors.As(err, &vpnErr) {
		return vpnErr.Type == "protocol"
	}
	
	return errors.Is(err, ErrMASQUEProtocol) ||
		   errors.Is(err, ErrQUICProtocol) ||
		   errors.Is(err, ErrHTTP3Protocol)
}

func IsSystemError(err error) bool {
	var vpnErr *VPNError
	if errors.As(err, &vpnErr) {
		return vpnErr.Type == "system"
	}
	
	return errors.Is(err, ErrPermissionDenied) ||
		   errors.Is(err, ErrResourceExhausted) ||
		   errors.Is(err, ErrSystemCall)
}

// Helper function to get current timestamp
func getCurrentTimestamp() int64 {
	// This would typically use time.Now().Unix()
	// Simplified for this example
	return 0
}

// Error recovery strategies
type RecoveryStrategy int

const (
	RecoveryNone RecoveryStrategy = iota
	RecoveryRetry
	RecoveryReconnect
	RecoveryRestart
	RecoveryFallback
)

// GetRecoveryStrategy determines the appropriate recovery strategy for an error
func GetRecoveryStrategy(err error) RecoveryStrategy {
	if IsConnectionError(err) {
		return RecoveryReconnect
	}
	
	if IsNetworkError(err) {
		return RecoveryRetry
	}
	
	if IsProtocolError(err) {
		return RecoveryReconnect
	}
	
	if IsConfigurationError(err) {
		return RecoveryNone // Configuration errors require manual intervention
	}
	
	if IsSystemError(err) {
		return RecoveryRestart
	}
	
	return RecoveryRetry // Default strategy
}