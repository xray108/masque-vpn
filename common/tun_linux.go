//go:build linux

package common

import (
	"fmt"
	"log"
	"net"
	"os/exec"
)

const (
	// TunPacketOffset is the offset required by TUN on Linux
	TunPacketOffset = 0
)

// getDefaultPlatformTunName returns the default TUN name for Linux
func getDefaultPlatformTunName() string {
	return "tun0"
}

// setPlatformIP sets the IP address on Linux
func setPlatformIP(tunDev *TUNDevice, ipNet net.IPNet) error {
	ip := ipNet.IP.String()
	mask, _ := ipNet.Mask.Size()
	
	// Use ip command on Linux
	cmd := exec.Command("ip", "addr", "add", fmt.Sprintf("%s/%d", ip, mask), "dev", tunDev.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to set IP address: %v, output: %s", err, string(output))
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", "dev", tunDev.Name(), "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring interface up: %v, output: %s", err, string(output))
	}

	log.Printf("Set IP address %s/%d on interface %s", ip, mask, tunDev.Name())
	return nil
}

// addPlatformRoute adds a route on Linux
func addPlatformRoute(tunDev *TUNDevice, ipNet net.IPNet) error {
	network := ipNet.String()

	// Use ip route command on Linux
	cmd := exec.Command("ip", "route", "add", network, "dev", tunDev.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to add route: %v, output: %s", err, string(output))
	}

	return nil
}