package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func main() {
	mode := flag.String("mode", "client", "Mode: client or server")
	addr := flag.String("addr", ":8081", "Address to listen on or connect to")
	duration := flag.Duration("duration", 10*time.Second, "Test duration (client mode)")
	flag.Parse()

	if *mode == "server" {
		runServer(*addr)
	} else {
		runClient(*addr, *duration)
	}
}

func runServer(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Printf("Load test server listening on %s", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// Discard all data
	n, err := io.Copy(io.Discard, conn)
	if err != nil {
		log.Printf("Connection error: %v", err)
	}
	log.Printf("Received %d bytes from %s", n, conn.RemoteAddr())
}

func runClient(addr string, duration time.Duration) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	log.Printf("Connected to %s, sending data for %v...", addr, duration)

	start := time.Now()
	deadline := start.Add(duration)
	buf := make([]byte, 32*1024) // 32KB buffer
	var totalBytes int64

	for time.Now().Before(deadline) {
		n, err := conn.Write(buf)
		if err != nil {
			log.Fatalf("Write error: %v", err)
		}
		totalBytes += int64(n)
	}

	elapsed := time.Since(start)
	mbps := (float64(totalBytes) * 8 / 1000 / 1000) / elapsed.Seconds()

	fmt.Printf("Test complete:\n")
	fmt.Printf("  Duration: %v\n", elapsed)
	fmt.Printf("  Total Data: %d bytes (%.2f MB)\n", totalBytes, float64(totalBytes)/1024/1024)
	fmt.Printf("  Throughput: %.2f Mbps\n", mbps)
}
