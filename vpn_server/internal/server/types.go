package server

import (
	"sync"

	common "github.com/iselt/masque-vpn/common"
	common_fec "github.com/iselt/masque-vpn/common/fec"
)

// ClientSession holds per-client state including FEC
type ClientSession struct {
	Conn         *common.MASQUEConn
	Encoder      *common_fec.XOREncoder
	PacketBuffer [][]byte
	SeqNum       uint32
	FecEnabled   bool
	Mu           sync.Mutex
}
