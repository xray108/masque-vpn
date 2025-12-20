package common

import (
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrefixToIPNet(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "IPv4 /24",
			prefix:   "192.168.1.0/24",
			expected: "192.168.1.0/24",
		},
		{
			name:     "IPv4 /32",
			prefix:   "10.0.0.1/32",
			expected: "10.0.0.1/32",
		},
		{
			name:     "IPv6 /64",
			prefix:   "2001:db8::/64",
			expected: "2001:db8::/64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := netip.ParsePrefix(tt.prefix)
			require.NoError(t, err)

			ipNet := PrefixToIPNet(prefix)
			assert.Equal(t, tt.expected, ipNet.String())
		})
	}
}

func TestLastIP(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		expected string
	}{
		{
			name:     "IPv4 /24",
			prefix:   "192.168.1.0/24",
			expected: "192.168.1.255",
		},
		{
			name:     "IPv4 /30",
			prefix:   "10.0.0.0/30",
			expected: "10.0.0.3",
		},
		{
			name:     "IPv4 /32",
			prefix:   "10.0.0.1/32",
			expected: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := netip.ParsePrefix(tt.prefix)
			require.NoError(t, err)

			lastIP := LastIP(prefix)
			assert.Equal(t, tt.expected, lastIP.String())
		})
	}
}

func TestNewNetworkInfo(t *testing.T) {
	tests := []struct {
		name        string
		cidr        string
		expectError bool
	}{
		{
			name:        "Valid IPv4 CIDR",
			cidr:        "10.0.0.0/24",
			expectError: false,
		},
		{
			name:        "Valid IPv6 CIDR",
			cidr:        "2001:db8::/64",
			expectError: false,
		},
		{
			name:        "Invalid CIDR",
			cidr:        "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			netInfo, err := NewNetworkInfo(tt.cidr)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, netInfo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, netInfo)
				
				gateway := netInfo.GetGateway()
				prefix := netInfo.GetPrefix()
				
				assert.True(t, gateway.IsValid())
				assert.True(t, prefix.IsValid())
			}
		})
	}
}

func TestGetIPAddresses(t *testing.T) {
	// Create a simple IPv4 packet
	packet := []byte{
		0x45, 0x00, 0x00, 0x1c, // Version, IHL, TOS, Total Length
		0x00, 0x00, 0x40, 0x00, // ID, Flags, Fragment Offset
		0x40, 0x11, 0x00, 0x00, // TTL, Protocol, Header Checksum
		0xc0, 0xa8, 0x01, 0x01, // Source IP: 192.168.1.1
		0xc0, 0xa8, 0x01, 0x02, // Destination IP: 192.168.1.2
	}

	src, dst, err := GetIPAddresses(packet, len(packet))
	require.NoError(t, err)
	
	assert.Equal(t, "192.168.1.1", src.String())
	assert.Equal(t, "192.168.1.2", dst.String())
}

func TestIPPool(t *testing.T) {
	prefix, err := netip.ParsePrefix("10.0.0.0/30")
	require.NoError(t, err)
	
	gateway, err := netip.ParseAddr("10.0.0.1")
	require.NoError(t, err)

	pool := NewIPPool(prefix, gateway)
	
	// Test allocation
	ip1, err := pool.Allocate("client1")
	assert.NoError(t, err)
	assert.True(t, ip1.IsValid())
	
	ip2, err := pool.Allocate("client2")
	assert.NoError(t, err)
	assert.True(t, ip2.IsValid())
	assert.NotEqual(t, ip1.Addr(), ip2.Addr())
	
	// Test pool exhaustion
	_, err = pool.Allocate("client3")
	assert.Error(t, err)
	
	// Test release
	pool.Release(ip1.Addr())
	ip3, err := pool.Allocate("client3")
	assert.NoError(t, err)
	assert.Equal(t, ip1.Addr(), ip3.Addr())
	
	// Test stats
	total, allocated, available := pool.Stats()
	assert.Equal(t, 2, total)
	assert.Equal(t, 2, allocated)
	assert.Equal(t, 0, available)
}

func TestNetipAddrToNetIP(t *testing.T) {
	tests := []struct {
		name string
		addr string
	}{
		{"IPv4", "192.168.1.1"},
		{"IPv6", "2001:db8::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := netip.ParseAddr(tt.addr)
			require.NoError(t, err)

			netIP := NetipAddrToNetIP(addr)
			assert.Equal(t, tt.addr, netIP.String())

			// Test round trip
			convertedAddr, ok := NetIPToNetipAddr(netIP)
			assert.True(t, ok)
			assert.Equal(t, addr, convertedAddr)
		})
	}
}