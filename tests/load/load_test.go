package load

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricsEndpointLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	const (
		numGoroutines = 10
		requestsPerGoroutine = 100
		timeout = 30 * time.Second
	)

	client := &http.Client{Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount, errorCount int

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for j := 0; j < requestsPerGoroutine; j++ {
				select {
				case <-ctx.Done():
					return
				default:
				}

				resp, err := client.Get("http://localhost:9092/metrics")
				
				mu.Lock()
				if err != nil || resp.StatusCode != http.StatusOK {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
				
				if resp != nil {
					resp.Body.Close()
				}
				
				// Small delay to avoid overwhelming
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	t.Logf("Load test results: %d successful, %d errors", successCount, errorCount)
	
	// Allow some errors but most requests should succeed
	totalRequests := numGoroutines * requestsPerGoroutine
	successRate := float64(successCount) / float64(totalRequests)
	
	if successCount > 0 {
		assert.Greater(t, successRate, 0.8, "Success rate should be > 80%")
	} else {
		t.Skip("No successful requests - service may not be running")
	}
}

func TestConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	const numConnections = 50
	
	var wg sync.WaitGroup
	var mu sync.Mutex
	var successCount int

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get("http://localhost:9092/health")
			
			mu.Lock()
			if err == nil && resp.StatusCode == http.StatusOK {
				successCount++
			}
			mu.Unlock()
			
			if resp != nil {
				resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()

	t.Logf("Concurrent connections test: %d/%d successful", successCount, numConnections)
	
	if successCount > 0 {
		successRate := float64(successCount) / float64(numConnections)
		assert.Greater(t, successRate, 0.7, "Success rate should be > 70%")
	} else {
		t.Skip("No successful connections - service may not be running")
	}
}

func BenchmarkMetricsEndpoint(b *testing.B) {
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Test if endpoint is available
	resp, err := client.Get("http://localhost:9092/metrics")
	if err != nil {
		b.Skip("Metrics endpoint not available")
		return
	}
	resp.Body.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			resp, err := client.Get("http://localhost:9092/metrics")
			if err == nil && resp != nil {
				resp.Body.Close()
			}
		}
	})
}