package common

import (
	"fmt"
	"log"
	"net"

	common_fec "github.com/iselt/masque-vpn/common/fec"
)

// ProxyFromTunToMASQUE reads packets from TUN device and sends them through MASQUE connection
func ProxyFromTunToMASQUE(tunDev *TUNDevice, masqueConn *MASQUEConn, errChan chan<- error, fecConfig *common_fec.Config) {
	defer func() {
		if r := recover(); r != nil {
			errChan <- fmt.Errorf("panic in TUN->MASQUE proxy: %v", r)
		}
	}()

	buffer := make([]byte, 2048)
	
	for {
		// Read packet from TUN device
		n, err := tunDev.ReadPacket(buffer, 0)
		if err != nil {
			if isNetworkClosed(err) {
				log.Println("TUN device closed, stopping TUN->MASQUE proxy")
				errChan <- nil
				return
			}
			errChan <- fmt.Errorf("failed to read from TUN device: %w", err)
			return
		}

		if n == 0 {
			continue
		}

		packetData := buffer[:n]

		// Send packet through MASQUE connection
		if err := masqueConn.WritePacket(packetData); err != nil {
			if isNetworkClosed(err) {
				log.Println("MASQUE connection closed, stopping TUN->MASQUE proxy")
				errChan <- nil
				return
			}
			errChan <- fmt.Errorf("failed to write to MASQUE connection: %w", err)
			return
		}
	}
}

// ProxyFromMASQUEToTun reads packets from MASQUE connection and writes them to TUN device
func ProxyFromMASQUEToTun(tunDev *TUNDevice, masqueConn *MASQUEConn, errChan chan<- error, fecConfig *common_fec.Config) {
	defer func() {
		if r := recover(); r != nil {
			errChan <- fmt.Errorf("panic in MASQUE->TUN proxy: %v", r)
		}
	}()

	buffer := make([]byte, 2048)
	
	for {
		// Read packet from MASQUE connection
		n, err := masqueConn.ReadPacket(buffer)
		if err != nil {
			if isNetworkClosed(err) {
				log.Println("MASQUE connection closed, stopping MASQUE->TUN proxy")
				errChan <- nil
				return
			}
			errChan <- fmt.Errorf("failed to read from MASQUE connection: %w", err)
			return
		}

		if n == 0 {
			continue
		}

		packetData := buffer[:n]

		// Write packet to TUN device
		if err := tunDev.WritePacket(packetData, 0); err != nil {
			if isNetworkClosed(err) {
				log.Println("TUN device closed, stopping MASQUE->TUN proxy")
				errChan <- nil
				return
			}
			errChan <- fmt.Errorf("failed to write to TUN device: %w", err)
			return
		}
	}
}

// isNetworkClosed checks if the error indicates a closed network connection
func isNetworkClosed(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for common network closed errors
	if netErr, ok := err.(*net.OpError); ok {
		return netErr.Err.Error() == "use of closed network connection"
	}
	
	// Check for EOF or connection reset
	errStr := err.Error()
	return errStr == "EOF" || 
		   errStr == "connection reset by peer" ||
		   errStr == "use of closed network connection"
}