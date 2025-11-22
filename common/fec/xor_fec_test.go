package fec

import (
	"bytes"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  Config{RedundancyPercent: 10, BlockSize: 10},
			wantErr: false,
		},
		{
			name:    "zero redundancy",
			config:  Config{RedundancyPercent: 0, BlockSize: 10},
			wantErr: false,
		},
		{
			name:    "negative redundancy",
			config:  Config{RedundancyPercent: -1, BlockSize: 10},
			wantErr: true,
		},
		{
			name:    "redundancy too high",
			config:  Config{RedundancyPercent: 101, BlockSize: 10},
			wantErr: true,
		},
		{
			name:    "zero block size",
			config:  Config{RedundancyPercent: 10, BlockSize: 0},
			wantErr: true,
		},
		{
			name:    "block size too large",
			config:  Config{RedundancyPercent: 10, BlockSize: 256},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_CalculateRedundancyPackets(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		dataPackets int
		want        int
	}{
		{
			name:        "10% of 10 packets",
			config:      Config{RedundancyPercent: 10, BlockSize: 10},
			dataPackets: 10,
			want:        1,
		},
		{
			name:        "10% of 100 packets",
			config:      Config{RedundancyPercent: 10, BlockSize: 10},
			dataPackets: 100,
			want:        10,
		},
		{
			name:        "20% of 50 packets",
			config:      Config{RedundancyPercent: 20, BlockSize: 10},
			dataPackets: 50,
			want:        10,
		},
		{
			name:        "0% redundancy",
			config:      Config{RedundancyPercent: 0, BlockSize: 10},
			dataPackets: 100,
			want:        0,
		},
		{
			name:        "small packet count rounds up",
			config:      Config{RedundancyPercent: 10, BlockSize: 10},
			dataPackets: 5,
			want:        1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.CalculateRedundancyPackets(tt.dataPackets)
			if got != tt.want {
				t.Errorf("Config.CalculateRedundancyPackets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXOREncoder_Encode(t *testing.T) {
	config := Config{RedundancyPercent: 10, BlockSize: 10}
	encoder, err := NewXOREncoder(config)
	if err != nil {
		t.Fatalf("NewXOREncoder() error = %v", err)
	}

	// Create test packets
	packets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
		[]byte("packet4"),
		[]byte("packet5"),
		[]byte("packet6"),
		[]byte("packet7"),
		[]byte("packet8"),
		[]byte("packet9"),
		[]byte("packet10"),
	}

	encoded, err := encoder.Encode(packets)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Should have original packets + redundancy packets
	expectedLen := len(packets) + config.CalculateRedundancyPackets(len(packets))
	if len(encoded) != expectedLen {
		t.Errorf("Encode() returned %d packets, want %d", len(encoded), expectedLen)
	}

	// First N packets should be original
	for i := 0; i < len(packets); i++ {
		if !bytes.Equal(encoded[i], packets[i]) {
			t.Errorf("Encode() packet %d modified, got %v, want %v", i, encoded[i], packets[i])
		}
	}
}

func TestXOREncoder_Decode_SingleLoss(t *testing.T) {
	config := Config{RedundancyPercent: 10, BlockSize: 10}
	encoder, err := NewXOREncoder(config)
	if err != nil {
		t.Fatalf("NewXOREncoder() error = %v", err)
	}

	decoder, err := NewXORDecoder(config)
	if err != nil {
		t.Fatalf("NewXORDecoder() error = %v", err)
	}

	// Create test packets
	original := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
		[]byte("packet3"),
		[]byte("packet4"),
		[]byte("packet5"),
		[]byte("packet6"),
		[]byte("packet7"),
		[]byte("packet8"),
		[]byte("packet9"),
		[]byte("packet10"),
	}

	// Encode
	encoded, err := encoder.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Simulate loss of packet 3 (index 2)
	lostIdx := 2
	lostPacket := encoded[lostIdx]
	encoded[lostIdx] = nil

	// Decode
	recovered, err := decoder.Decode(encoded, []int{lostIdx})
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Check if packet was recovered
	if recovered[lostIdx] == nil {
		t.Errorf("Decode() failed to recover packet %d", lostIdx)
	} else if !bytes.Equal(recovered[lostIdx], lostPacket) {
		t.Errorf("Decode() recovered packet %d incorrectly, got %v, want %v", lostIdx, recovered[lostIdx], lostPacket)
	}
}

func TestXOREncoder_EmptyPackets(t *testing.T) {
	config := Config{RedundancyPercent: 10, BlockSize: 10}
	encoder, err := NewXOREncoder(config)
	if err != nil {
		t.Fatalf("NewXOREncoder() error = %v", err)
	}

	// Empty packet list
	encoded, err := encoder.Encode([][]byte{})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if len(encoded) != 0 {
		t.Errorf("Encode() with empty packets returned %d packets, want 0", len(encoded))
	}
}

func TestXOREncoder_ZeroRedundancy(t *testing.T) {
	config := Config{RedundancyPercent: 0, BlockSize: 10}
	encoder, err := NewXOREncoder(config)
	if err != nil {
		t.Fatalf("NewXOREncoder() error = %v", err)
	}

	packets := [][]byte{
		[]byte("packet1"),
		[]byte("packet2"),
	}

	encoded, err := encoder.Encode(packets)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Should return original packets only
	if len(encoded) != len(packets) {
		t.Errorf("Encode() with 0%% redundancy returned %d packets, want %d", len(encoded), len(packets))
	}
}

func BenchmarkXOREncoder_Encode(b *testing.B) {
	config := Config{RedundancyPercent: 10, BlockSize: 10}
	encoder, _ := NewXOREncoder(config)

	// Create 100 packets of 1500 bytes each (typical MTU)
	packets := make([][]byte, 100)
	for i := range packets {
		packets[i] = make([]byte, 1500)
		for j := range packets[i] {
			packets[i][j] = byte(i + j)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = encoder.Encode(packets)
	}
}

func BenchmarkXORDecoder_Decode(b *testing.B) {
	config := Config{RedundancyPercent: 10, BlockSize: 10}
	encoder, _ := NewXOREncoder(config)
	decoder, _ := NewXORDecoder(config)

	// Create and encode packets
	packets := make([][]byte, 100)
	for i := range packets {
		packets[i] = make([]byte, 1500)
		for j := range packets[i] {
			packets[i][j] = byte(i + j)
		}
	}

	encoded, _ := encoder.Encode(packets)
	
	// Simulate loss
	encoded[5] = nil
	lost := []int{5}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = decoder.Decode(encoded, lost)
	}
}
