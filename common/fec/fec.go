package fec

import "fmt"

// Encoder encodes data packets with forward error correction
type Encoder interface {
	// Encode takes a slice of data packets and returns original + redundancy packets
	Encode(packets [][]byte) ([][]byte, error)
}

// Decoder decodes packets and recovers lost data using FEC
type Decoder interface {
	// Decode attempts to recover lost packets from received packets
	// packets: all received packets (data + redundancy)
	// lost: indices of lost packets
	// Returns recovered packets
	Decode(packets [][]byte, lost []int) ([][]byte, error)
}

// Config holds FEC configuration parameters
type Config struct {
	// RedundancyPercent is the percentage of redundancy packets to add
	// For example, 10 means 10% redundancy (1 redundancy packet per 10 data packets)
	RedundancyPercent int

	// BlockSize is the number of data packets per FEC block
	// Redundancy packets are calculated per block
	BlockSize int
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.RedundancyPercent < 0 || c.RedundancyPercent > 100 {
		return fmt.Errorf("redundancy_percent must be between 0 and 100, got %d", c.RedundancyPercent)
	}
	if c.BlockSize <= 0 {
		return fmt.Errorf("block_size must be positive, got %d", c.BlockSize)
	}
	if c.BlockSize > 255 {
		return fmt.Errorf("block_size must be <= 255, got %d", c.BlockSize)
	}
	return nil
}

// CalculateRedundancyPackets calculates number of redundancy packets for given data packets
func (c *Config) CalculateRedundancyPackets(dataPackets int) int {
	if c.RedundancyPercent == 0 {
		return 0
	}
	// Calculate redundancy: dataPackets * (redundancyPercent / 100)
	redundancy := (dataPackets * c.RedundancyPercent) / 100
	if redundancy == 0 && c.RedundancyPercent > 0 {
		redundancy = 1 // At least 1 redundancy packet if enabled
	}
	return redundancy
}
