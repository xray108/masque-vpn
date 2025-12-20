//go:build darwin

package common

import (
	"fmt"
	"log"
	"net"
	"os/exec"
	"strings"

	"golang.zx2c4.com/wireguard/tun"
)

const (
	// TunPacketOffset is the offset required by WireGuard TUN on macOS
	TunPacketOffset = 4
)

// getDefaultPlatformTunName returns the default TUN name for macOS
func getDefaultPlatformTunName() string {
	return "utun"
}

// setPlatformIP sets the IP address on macOS
func setPlatformIP(tunDev *TUNDevice, ipNet net.IPNet) error {
	ip := ipNet.IP.String()
	mask, _ := ipNet.Mask.Size()
	
	if ipNet.IP.To4() != nil {
		// IPv4 - macOS requires point-to-point configuration
		destIP := make(net.IP, len(ipNet.IP))
		copy(destIP, ipNet.IP)
		destIP[len(destIP)-1]++ // Next IP as destination
		
		cmd := exec.Command("ifconfig", tunDev.Name(), "inet", ip, destIP.String(), "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set IPv4 address: %v, output: %s", err, string(output))
		}
	} else {
		// IPv6
		cmd := exec.Command("ifconfig", tunDev.Name(), "inet6", fmt.Sprintf("%s/%d", ip, mask), "up")
		if output, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set IPv6 address: %v, output: %s", err, string(output))
		}
	}

	log.Printf("Set IP address %s/%d on interface %s", ip, mask, tunDev.Name())
	return nil
}

// addPlatformRoute adds a route on macOS
func addPlatformRoute(tunDev *TUNDevice, ipNet net.IPNet) error {
	network := ipNet.String()
	
	var cmd *exec.Cmd
	if ipNet.IP.To4() != nil {
		// IPv4 route
		parts := strings.Split(network, "/")
		if len(parts) != 2 {
			return fmt.Errorf("invalid network format: %s", network)
		}
		cmd = exec.Command("route", "add", "-net", parts[0], "-interface", tunDev.Name())
	} else {
		// IPv6 route
		cmd = exec.Command("route", "add", "-inet6", network, "-interface", tunDev.Name())
	}

	if output, err := cmd.CombinedOutput(); err != nil {
		if !strings.Contains(string(output), "File exists") {
			return fmt.Errorf("failed to add route: %v, output: %s", err, string(output))
		}
	}

	return nil
}
