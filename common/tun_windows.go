//go:build windows

package common

import (
	"fmt"
	"log"
	"net"
	"os/exec"
)

const (
	// TunPacketOffset is the offset required by WireGuard TUN on Windows
	TunPacketOffset = 0
)

// getDefaultPlatformTunName returns the default TUN name for Windows
func getDefaultPlatformTunName() string {
	return "wintun"
}

// setPlatformIP sets the IP address on Windows
func setPlatformIP(tunDev *TUNDevice, ipNet net.IPNet) error {
	ip := ipNet.IP.String()
	mask, _ := ipNet.Mask.Size()
	
	// Convert CIDR to netmask
	netmask := cidrToNetmask(mask)
	
	// Use netsh command on Windows
	cmd := exec.Command("netsh", "interface", "ip", "set", "address", 
		"name="+tunDev.Name(), "static", ip, netmask)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set IP address: %v, output: %s", err, string(output))
	}

	log.Printf("Set IP address %s/%d on interface %s", ip, mask, tunDev.Name())
	return nil
}

// addPlatformRoute adds a route on Windows
func addPlatformRoute(tunDev *TUNDevice, ipNet net.IPNet) error {
	network := ipNet.String()

	// Use route command on Windows
	cmd := exec.Command("route", "add", network, "0.0.0.0", "if", tunDev.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add route: %v, output: %s", err, string(output))
	}

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