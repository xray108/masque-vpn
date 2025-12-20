package common

import (
	"net"
	"net/netip"
)

// NetipAddrToNetIP converts netip.Addr to net.IP
func NetipAddrToNetIP(addr netip.Addr) net.IP {
	return net.IP(addr.AsSlice())
}

// NetIPToNetipAddr converts net.IP to netip.Addr
func NetIPToNetipAddr(ip net.IP) (netip.Addr, bool) {
	return netip.AddrFromSlice(ip)
}