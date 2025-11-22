package fec

import (
	"fmt"
)

// XOREncoder implements simple XOR-based FEC encoding
// For each block of N data packets, it creates redundancy packets by XORing all data packets
type XOREncoder struct {
	config Config
}

// NewXOREncoder creates a new XOR-based FEC encoder
func NewXOREncoder(config Config) (*XOREncoder, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid FEC config: %w", err)
	}
	return &XOREncoder{config: config}, nil
}

// Encode encodes data packets with XOR-based FEC
// Returns original packets + redundancy packets
func (e *XOREncoder) Encode(packets [][]byte) ([][]byte, error) {
	if len(packets) == 0 {
		return packets, nil
	}

	// Calculate number of redundancy packets needed
	numRedundancy := e.config.CalculateRedundancyPackets(len(packets))
	if numRedundancy == 0 {
		return packets, nil
	}

	// Process packets in blocks
	result := make([][]byte, 0, len(packets)+numRedundancy)
	result = append(result, packets...)

	// Calculate number of complete blocks
	numBlocks := (len(packets) + e.config.BlockSize - 1) / e.config.BlockSize

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		blockStart := blockIdx * e.config.BlockSize
		blockEnd := blockStart + e.config.BlockSize
		if blockEnd > len(packets) {
			blockEnd = len(packets)
		}

		blockPackets := packets[blockStart:blockEnd]
		
		// Create redundancy packet by XORing all packets in the block
		redundancyPacket := e.xorPackets(blockPackets)
		if redundancyPacket != nil {
			result = append(result, redundancyPacket)
		}
	}

	return result, nil
}

// xorPackets XORs all packets together to create a redundancy packet
// The redundancy packet includes length information for each packet in the block
func (e *XOREncoder) xorPackets(packets [][]byte) []byte {
	if len(packets) == 0 {
		return nil
	}

	// Find maximum packet size
	maxSize := 0
	for _, pkt := range packets {
		if len(pkt) > maxSize {
			maxSize = len(pkt)
		}
	}

	if maxSize == 0 {
		return nil
	}

	// Create redundancy packet with header for packet lengths
	// Header: [num_packets(1 byte)][len1(2 bytes)][len2(2 bytes)]...
	headerSize := 1 + len(packets)*2
	redundancy := make([]byte, headerSize+maxSize)
	
	// Write header
	redundancy[0] = byte(len(packets))
	for i, pkt := range packets {
		offset := 1 + i*2
		redundancy[offset] = byte(len(pkt) >> 8)     // High byte
		redundancy[offset+1] = byte(len(pkt) & 0xFF) // Low byte
	}

	// XOR all packets (after header)
	for _, pkt := range packets {
		for i := 0; i < len(pkt); i++ {
			redundancy[headerSize+i] ^= pkt[i]
		}
	}

	return redundancy
}

// XORDecoder implements XOR-based FEC decoding
type XORDecoder struct {
	config Config
}

// NewXORDecoder creates a new XOR-based FEC decoder
func NewXORDecoder(config Config) (*XORDecoder, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid FEC config: %w", err)
	}
	return &XORDecoder{config: config}, nil
}

// Decode attempts to recover lost packets using XOR-based FEC
// packets: all received packets (may include nil for lost packets)
// lost: indices of lost packets
// Returns recovered packets (same indices as lost)
func (d *XORDecoder) Decode(packets [][]byte, lost []int) ([][]byte, error) {
	if len(lost) == 0 {
		return [][]byte{}, nil
	}

	// For XOR FEC, we can only recover if there's exactly one lost packet per block
	// and we have the redundancy packet for that block

	recovered := make([][]byte, len(packets))
	
	// Group lost packets by block
	lostByBlock := make(map[int][]int)
	for _, lostIdx := range lost {
		blockIdx := lostIdx / d.config.BlockSize
		lostByBlock[blockIdx] = append(lostByBlock[blockIdx], lostIdx)
	}

	// Try to recover each block
	for blockIdx, lostIndices := range lostByBlock {
		// XOR can only recover if exactly 1 packet is lost in the block
		if len(lostIndices) != 1 {
			continue
		}

		lostIdx := lostIndices[0]
		blockStart := blockIdx * d.config.BlockSize
		blockEnd := blockStart + d.config.BlockSize
		if blockEnd > len(packets) {
			blockEnd = len(packets)
		}

		// Find redundancy packet for this block
		// Redundancy packets are stored after data packets
		numDataPackets := len(packets) - (len(packets) / (d.config.BlockSize + 1))
		redundancyIdx := numDataPackets + blockIdx
		
		if redundancyIdx >= len(packets) || packets[redundancyIdx] == nil {
			continue
		}

		redundancyPacket := packets[redundancyIdx]
		
		// Parse header to get packet lengths
		if len(redundancyPacket) < 1 {
			continue
		}
		
		numPacketsInBlock := int(redundancyPacket[0])
		headerSize := 1 + numPacketsInBlock*2
		
		if len(redundancyPacket) < headerSize {
			continue
		}
		
		// Get original packet length for the lost packet
		lostPacketIndexInBlock := lostIdx - blockStart
		if lostPacketIndexInBlock >= numPacketsInBlock {
			continue
		}
		
		offset := 1 + lostPacketIndexInBlock*2
		originalLen := int(redundancyPacket[offset])<<8 | int(redundancyPacket[offset+1])

		// Collect all available packets in the block (excluding lost one)
		var availablePackets [][]byte
		for i := blockStart; i < blockEnd; i++ {
			if i != lostIdx && i < len(packets) && packets[i] != nil {
				availablePackets = append(availablePackets, packets[i])
			}
		}

		// Recover lost packet by XORing available packets with redundancy data
		recoveredPacket := d.xorPacketsWithRedundancy(availablePackets, redundancyPacket, headerSize, originalLen)
		if recoveredPacket != nil {
			recovered[lostIdx] = recoveredPacket
		}
	}

	return recovered, nil
}

// xorPacketsWithRedundancy XORs available packets with redundancy packet to recover lost packet
func (d *XORDecoder) xorPacketsWithRedundancy(availablePackets [][]byte, redundancyPacket []byte, headerSize int, originalLen int) []byte {
	if len(redundancyPacket) < headerSize {
		return nil
	}

	// Find maximum size needed
	maxSize := originalLen
	for _, pkt := range availablePackets {
		if len(pkt) > maxSize {
			maxSize = len(pkt)
		}
	}

	// Start with redundancy data (skip header)
	result := make([]byte, maxSize)
	redundancyData := redundancyPacket[headerSize:]
	copy(result, redundancyData[:min(len(redundancyData), maxSize)])

	// XOR with all available packets
	for _, pkt := range availablePackets {
		for i := 0; i < len(pkt); i++ {
			result[i] ^= pkt[i]
		}
	}

	// Return only the original length
	return result[:originalLen]
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
