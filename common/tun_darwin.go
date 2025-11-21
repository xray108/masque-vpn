//go:build darwin
// +build darwin

package common

import (
	"fmt"
	"log"
	"net/netip"
	"os/exec"
	"strings"

	"golang.zx2c4.com/wireguard/tun"
)

const (
	// TunPacketOffset is the offset required by WireGuard TUN on macOS
	// The WireGuard TUN implementation on Darwin requires a 4-byte offset
	TunPacketOffset = 4
)

// TUNDevice is the TUN device implementation for macOS
type TUNDevice struct {
	device    tun.Device
	nativeTun *tun.NativeTun
	name      string
	ipAddress netip.Addr
}

// SetIP sets the TUN device IP address to the gateway IP
func (t *TUNDevice) SetIP(ipPrefix netip.Prefix) error {
	ipAddr := ipPrefix.Addr().String()
	mask := ipPrefix.Bits()
	
	if ipPrefix.Addr().Is4() {
		// For macOS TUN devices (point-to-point), we need both local and destination addresses
		// Use the gateway IP as local address and calculate a destination address
		// For a /24 network like 10.0.0.0/24, if gateway is 10.0.0.1, use 10.0.0.2 as destination
		destAddr := ipPrefix.Addr().Next()
		
		// macOS ifconfig format for point-to-point: ifconfig interface inet LOCAL DEST up
		cmd := exec.Command("ifconfig", t.name, "inet", ipAddr, destAddr.String(), "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set IP address via ifconfig: %v, output: %s", err, string(output))
		}
	} else {
		// IPv6 - use prefix length directly
		cmd := exec.Command("ifconfig", t.name, "inet6", fmt.Sprintf("%s/%d", ipAddr, mask), "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set IPv6 address via ifconfig: %v, output: %s", err, string(output))
		}
	}

	t.ipAddress = ipPrefix.Addr()
	log.Printf("Set IP address %s/%d on interface %s", ipAddr, mask, t.name)
	return nil
}

// cidrToNetmask converts CIDR prefix length to netmask (e.g., 24 -> 255.255.255.0)
func cidrToNetmask(cidr int) string {
	if cidr < 0 || cidr > 32 {
		return "255.255.255.0" // default
	}
	
	mask := uint32((0xFFFFFFFF << (32 - cidr)) & 0xFFFFFFFF)
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(mask>>24),
		byte(mask>>16),
		byte(mask>>8),
		byte(mask))
}

// AddRoute adds a route to the TUN device
func (t *TUNDevice) AddRoute(prefix netip.Prefix) error {
	network := prefix.Masked()
	networkStr := network.String()
	gatewayIP := t.ipAddress.String()

	// Use route command on macOS
	var cmd *exec.Cmd
	if prefix.Addr().Is4() {
		// IPv4 route: route add -net <network> <gateway>
		parts := strings.Split(networkStr, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid network format: %s", networkStr)
		}
		cmd = exec.Command("route", "add", "-net", parts[0], gatewayIP)
	} else {
		// IPv6 route: route add -inet6 <network> <gateway>
		cmd = exec.Command("route", "add", "-inet6", networkStr, gatewayIP)
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if route already exists
		if !strings.Contains(string(output), "File exists") {
			return fmt.Errorf("failed to add route: %v, output: %s", err, string(output))
		}
	}

	return nil
}

// CreateTunDevice creates a TUN device with the specified name and gateway IP
func CreateTunDevice(name string, gatewayPrefix netip.Prefix, mtu int) (*TUNDevice, error) {
	if name == "" {
		name = "utun"
	}

	device, err := tun.CreateTUN(name, mtu)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %v", err)
	}

	nativeTun := device.(*tun.NativeTun)
	actualName, err := nativeTun.Name()
	if err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to get device name: %v", err)
	}

	tunDev := &TUNDevice{
		device:    device,
		nativeTun: nativeTun,
		name:      actualName,
		ipAddress: gatewayPrefix.Addr(),
	}

	if err := tunDev.SetIP(gatewayPrefix); err != nil {
		device.Close()
		return nil, fmt.Errorf("failed to set IP: %v", err)
	}

	log.Printf("Created TUN device: %s with IP %s", actualName, gatewayPrefix)

	if err := configureRouting(actualName, gatewayPrefix); err != nil {
		log.Printf("Warning: failed to configure routing: %v", err)
	}

	return tunDev, nil
}

func configureRouting(ifName string, prefix netip.Prefix) error {
	gatewayIP := prefix.Addr()
	networkPrefix := prefix.Masked()
	networkStr := networkPrefix.String()

	var cmd *exec.Cmd
	if prefix.Addr().Is4() {
		// IPv4 route: route add -net <network> <gateway>
		parts := strings.Split(networkStr, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid network format: %s", networkStr)
		}
		cmd = exec.Command("route", "add", "-net", parts[0], gatewayIP.String())
	} else {
		// IPv6 route: route add -inet6 <network> <gateway>
		cmd = exec.Command("route", "add", "-inet6", networkStr, gatewayIP.String())
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		// Ignore error if route already exists
		if !strings.Contains(string(output), "File exists") {
			return fmt.Errorf("failed to add route: %v, output: %s", err, string(output))
		}
	}

	return nil
}
