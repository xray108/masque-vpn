package server

import (
	"sync"

	connectip "github.com/iselt/connect-ip-go"
	common_fec "github.com/iselt/masque-vpn/common/fec"
)

// ClientSession holds per-client state including FEC
type ClientSession struct {
	Conn         *connectip.Conn
	Encoder      *common_fec.XOREncoder
	PacketBuffer [][]byte
	SeqNum       uint32
	FecEnabled   bool
	Mu           sync.Mutex
}
