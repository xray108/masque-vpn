package common

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	connectip "github.com/iselt/connect-ip-go"
	"github.com/iselt/masque-vpn/common/fec"
	"github.com/quic-go/quic-go"
	"golang.zx2c4.com/wireguard/tun"
)

// 定义缓冲区大小
const (
	BufferSize      = 2048 // 标准MTU大小
	VirtioNetHdrLen = 10   // virtio-net 头部长度
	PacketHeaderLen = 4    // 4 bytes for sequence number
)

// 为TUN->VPN方向创建缓冲区池
var tunToVPNBufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, BufferSize)
	},
}

// 为VPN->TUN方向创建缓冲区池（包含virtio-net头部）
var vpnToTunBufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, BufferSize+VirtioNetHdrLen)
		// 预先清零头部区域
		for i := 0; i < VirtioNetHdrLen; i++ {
			buf[i] = 0
		}
		return buf
	},
}

// isNetworkClosed 判断错误是否表示网络连接已关闭
func isNetworkClosed(err error) bool {
	var netErr *net.OpError
	var qErr *quic.ApplicationError

	return errors.Is(err, io.EOF) ||
		errors.Is(err, net.ErrClosed) ||
		errors.As(err, &netErr) ||
		errors.As(err, &qErr)
}

// ProxyFromTunToVPN 从TUN设备读取数据包并发送到VPN连接
// 优化版本：直接使用批量读取
func ProxyFromTunToVPN(dev *TUNDevice, ipconn *connectip.Conn, errChan chan<- error, fecConfig *fec.Config) {
	// 确定批量大小
	batchSize := dev.BatchSize()
	if batchSize <= 0 {
		batchSize = 32 // 默认批量大小
	}

	// FEC setup
	var encoder *fec.XOREncoder
	var fecEnabled bool
	var seqNum uint32
	var packetBuffer [][]byte

	if fecConfig != nil && fecConfig.Enabled {
		fecEnabled = true
		config := fec.Config{
			RedundancyPercent: fecConfig.RedundancyPercent,
			BlockSize:         fecConfig.BlockSize,
		}
		var err error
		encoder, err = fec.NewXOREncoder(config)
		if err != nil {
			errChan <- fmt.Errorf("failed to create FEC encoder: %w", err)
			return
		}
		// Calculate total packets per block (data + redundancy)
		packetBuffer = make([][]byte, 0, config.BlockSize)
		log.Printf("FEC enabled: %d%% redundancy, block size %d", config.RedundancyPercent, config.BlockSize)
	}

	// 预先分配批量读取的缓冲区
	packetBufs := make([][]byte, batchSize)
	sizes := make([]int, batchSize)
	for i := range packetBufs {
		packetBufs[i] = make([]byte, BufferSize)
	}

	for {
		// 直接从TUN设备批量读取数据包
		n, err := dev.Read(packetBufs, sizes, 0)

		if err != nil {
			if errors.Is(err, os.ErrClosed) || errors.Is(err, net.ErrClosed) {
				log.Println("TUN device closed, stopping Tun->VPN proxy.")
				errChan <- nil
				return
			}

			if errors.Is(err, tun.ErrTooManySegments) {
				log.Println("Warning: Too many segments in TUN device read, continuing...")
				continue
			}

			errChan <- fmt.Errorf("failed to read batch from TUN device %s: %w", dev.Name(), err)
			return
		}

		if n == 0 {
			continue // 这批次没有数据包
		}

		// 处理批次中的每个数据包
		for i := 0; i < n; i++ {
			packetData := packetBufs[i][:sizes[i]]

			if fecEnabled {
				// Copy packet data because packetBufs is reused
				pktCopy := make([]byte, len(packetData))
				copy(pktCopy, packetData)
				packetBuffer = append(packetBuffer, pktCopy)

				// If buffer full, encode and send
				if len(packetBuffer) >= encoder.Config().BlockSize {
					if err := EncodeAndSendBlock(encoder, ipconn, packetBuffer, &seqNum); err != nil {
						if isNetworkClosed(err) {
							log.Println("Connection closed during FEC send, stopping Tun->VPN proxy.")
							errChan <- nil
						} else {
							errChan <- fmt.Errorf("failed to send FEC block: %w", err)
						}
						return
					}
					packetBuffer = packetBuffer[:0] // Clear buffer
				}
			} else {
				// No FEC: Just send raw packet
				_, writeErr := ipconn.WritePacket(packetData)
				if writeErr != nil {
					if isNetworkClosed(writeErr) {
						log.Println("Connection closed, stopping Tun->VPN proxy.")
						errChan <- nil
					} else {
						errChan <- fmt.Errorf("failed to write to connect-ip connection: %w", writeErr)
					}
					return
				}
			}
		}

		// Flush remaining packets in buffer if any (and if using FEC)
		// Note: Ideally we should use a timer to flush partial blocks to avoid latency
		// For now, we flush if we have pending packets after processing the batch
		// This might add latency if batch is small.
		// TODO: Implement timer-based flushing
		if fecEnabled && len(packetBuffer) > 0 {
			if err := EncodeAndSendBlock(encoder, ipconn, packetBuffer, &seqNum); err != nil {
				if isNetworkClosed(err) {
					log.Println("Connection closed during FEC flush, stopping Tun->VPN proxy.")
					errChan <- nil
				} else {
					errChan <- fmt.Errorf("failed to send FEC block: %w", err)
				}
				return
			}
			packetBuffer = packetBuffer[:0]
		}
	}
}

// EncodeAndSendBlock encodes buffered packets and sends them with sequence numbers
func EncodeAndSendBlock(encoder *fec.XOREncoder, ipconn *connectip.Conn, packets [][]byte, seqNum *uint32) error {
	encoded, err := encoder.Encode(packets)
	if err != nil {
		return err
	}

	for _, pkt := range encoded {
		// Add sequence number header
		// [SeqNum (4 bytes)][Data]
		dataWithHeader := make([]byte, 4+len(pkt))
		binary.BigEndian.PutUint32(dataWithHeader[0:4], *seqNum)
		copy(dataWithHeader[4:], pkt)

		_, writeErr := ipconn.WritePacket(dataWithHeader)
		if writeErr != nil {
			return writeErr
		}
		*seqNum++
	}
	return nil
}

// ProxyFromVPNToTun 从VPN连接读取数据包并写入TUN设备
func ProxyFromVPNToTun(dev *TUNDevice, ipconn *connectip.Conn, errChan chan<- error, fecConfig *fec.Config) {
	// FEC setup
	var decoder *fec.XORDecoder
	var fecEnabled bool
	var currentBlockID uint32 = 0xFFFFFFFF // Invalid initial block ID
	var blockBuffer [][]byte
	var receivedIndices []int
	var totalBlockSize int

	if fecConfig != nil && fecConfig.Enabled {
		fecEnabled = true
		config := fec.Config{
			RedundancyPercent: fecConfig.RedundancyPercent,
			BlockSize:         fecConfig.BlockSize,
		}
		var err error
		decoder, err = fec.NewXORDecoder(config)
		if err != nil {
			errChan <- fmt.Errorf("failed to create FEC decoder: %w", err)
			return
		}
		redundancy := config.CalculateRedundancyPackets(config.BlockSize)
		totalBlockSize = config.BlockSize + redundancy
		blockBuffer = make([][]byte, totalBlockSize)
		receivedIndices = make([]int, 0, totalBlockSize)
	}

	for {
		// 从池中获取预先准备好virtio头的缓冲区
		buf := vpnToTunBufferPool.Get().([]byte)

		// 直接读取到缓冲区的offset位置，避免后续的复制操作
		n, err := ipconn.ReadPacket(buf[VirtioNetHdrLen:])

		if err != nil {
			buf = buf[:cap(buf)]
			vpnToTunBufferPool.Put(buf) // 归还缓冲区

			if isNetworkClosed(err) {
				log.Println("Connection closed, stopping VPN->Tun proxy.")
				errChan <- nil
			} else {
				errChan <- fmt.Errorf("failed to read from connect-ip connection: %w", err)
			}
			return
		}

		if n == 0 {
			buf = buf[:cap(buf)]
			vpnToTunBufferPool.Put(buf) // 归还缓冲区
			continue
		}

		packetData := buf[VirtioNetHdrLen : VirtioNetHdrLen+n]

		if fecEnabled {
			if n < 4 {
				// Too short for header, drop
				buf = buf[:cap(buf)]
				vpnToTunBufferPool.Put(buf)
				continue
			}

			// Extract sequence number
			seq := binary.BigEndian.Uint32(packetData[0:4])
			payload := packetData[4:]

			blockID := seq / uint32(totalBlockSize)
			indexInBlock := int(seq % uint32(totalBlockSize))

			// If new block, flush old one
			if blockID != currentBlockID {
				if currentBlockID != 0xFFFFFFFF {
					// Try to recover missing packets in previous block
					recoverAndWriteBlock(decoder, dev, blockBuffer, receivedIndices, currentBlockID)
				}
				
				// Reset buffer for new block
				// If we skipped multiple blocks, we lost them entirely
				currentBlockID = blockID
				for i := range blockBuffer {
					blockBuffer[i] = nil
				}
				receivedIndices = receivedIndices[:0]
			}

			// Store packet in buffer
			if indexInBlock < len(blockBuffer) {
				// Copy payload because buf is reused
				pktCopy := make([]byte, len(payload))
				copy(pktCopy, payload)
				blockBuffer[indexInBlock] = pktCopy
				receivedIndices = append(receivedIndices, indexInBlock)
			}

			// If we have enough packets (original data packets), we can write them immediately?
			// No, we might receive redundancy packets first.
			// But for low latency, we should write original data packets as soon as they arrive.
			// And only use redundancy to recover LOST packets later.
			// But we need to know if it's a data packet or redundancy packet.
			// In our scheme: 0..BlockSize-1 are Data, BlockSize..Total-1 are Redundancy.
			
			if indexInBlock < decoder.Config().BlockSize {
				// It's a data packet, write immediately to TUN
				// We need to copy to a buffer with Virtio header for dev.WritePacket
				// But we already have it in 'buf' (with header offset), just need to shift/adjust?
				// 'buf' contains [Header(4)][Payload].
				// We need [Virtio][Payload].
				// We can reuse 'buf' if we shift payload?
				// Or just allocate new buffer for WritePacket (easier)
				
				// Write packet to TUN
				if _, err := dev.WritePacket(payload, 0); err != nil { // 0 offset because we pass pure payload? No, WritePacket expects buffer with offset space?
					// common.TUNDevice.WritePacket implementation:
					// It expects 'data' to start with the packet, but if offset > 0, it expects 'offset' bytes before.
					// Wait, common/tun_darwin.go WritePacket:
					// func (t *TUNDevice) WritePacket(data []byte, offset int) (int, error)
					// It writes data[offset:] to TUN.
					// So we should pass data containing [offset bytes][packet].
					
					// Here 'payload' is just the packet data.
					// We need to prepend offset bytes.
					// Let's use a temp buffer from pool?
					// Or just make a new buffer.
					
					// Optimization: We can modify 'buf' in place to remove SeqNum header and add Virtio header?
					// buf: [Virtio(10)][Seq(4)][Payload...]
					// We want: [Virtio(10)][Payload...]
					// We can shift payload back by 4 bytes.
					copy(buf[VirtioNetHdrLen:], payload)
					// Now buf[VirtioNetHdrLen:] contains Payload.
					// Total length is VirtioNetHdrLen + len(payload).
					
					if _, err := dev.WritePacket(buf[:VirtioNetHdrLen+len(payload)], VirtioNetHdrLen); err != nil {
						if !errors.Is(err, os.ErrClosed) && !errors.Is(err, net.ErrClosed) {
							log.Printf("Warning: failed to write packet to TUN: %v", err)
						}
					}
				}
			}
			
			// We still keep it in blockBuffer for potential recovery of OTHER packets
			// But we don't need to write it again in recoverAndWriteBlock.
			
			buf = buf[:cap(buf)]
			vpnToTunBufferPool.Put(buf)

		} else {
			// No FEC: Just write packet to TUN
			if _, err := dev.WritePacket(buf[:n+VirtioNetHdrLen], VirtioNetHdrLen); err != nil {
				buf = buf[:cap(buf)]
				vpnToTunBufferPool.Put(buf) // 归还缓冲区

				if errors.Is(err, os.ErrClosed) || errors.Is(err, net.ErrClosed) {
					log.Println("TUN device closed, stopping VPN->Tun proxy.")
					errChan <- nil
				} else {
					errChan <- fmt.Errorf("failed to write packet to TUN device %s: %w", dev.Name(), err)
				}
				return
			}
			buf = buf[:cap(buf)]
			vpnToTunBufferPool.Put(buf)
		}
	}
}

// recoverAndWriteBlock tries to recover lost packets in a block and write them to TUN
func recoverAndWriteBlock(decoder *fec.XORDecoder, dev *TUNDevice, blockBuffer [][]byte, receivedIndices []int, blockID uint32) {
	// Identify lost packets
	lostIndices := make([]int, 0)
	receivedMap := make(map[int]bool)
	for _, idx := range receivedIndices {
		receivedMap[idx] = true
	}
	
	for i := 0; i < len(blockBuffer); i++ {
		if !receivedMap[i] {
			lostIndices = append(lostIndices, i)
		}
	}
	
	if len(lostIndices) == 0 {
		return // No loss
	}
	
	// Decode
	recovered, err := decoder.Decode(blockBuffer, lostIndices)
	if err != nil {
		log.Printf("FEC decode error: %v", err)
		return
	}
	
	// Write RECOVERED DATA packets to TUN
	// Only write if index < BlockSize (data packets)
	// Redundancy packets are not written to TUN
	for i, pkt := range recovered {
		if pkt != nil && i < decoder.Config().BlockSize {
			// This was a lost data packet, now recovered
			// log.Printf("FEC recovered packet %d in block %d", i, blockID)
			
			// We need to write it to TUN.
			// Allocate buffer with offset
			buf := make([]byte, VirtioNetHdrLen+len(pkt))
			copy(buf[VirtioNetHdrLen:], pkt)
			
			if _, err := dev.WritePacket(buf, VirtioNetHdrLen); err != nil {
				log.Printf("Warning: failed to write recovered packet to TUN: %v", err)
			}
		}
	}
}
