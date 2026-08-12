// Package main implements a UDP server that echoes messages.
// This exercise teaches UDP socket fundamentals in Go.
//
// Learning objectives:
// - Create UDP listener
// - Handle connectionless protocol
// - Understand differences from TCP
//
// Run: go run main.go
// Test: echo "hello" | nc -u localhost 9999
package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

const addr = ":9999"

// Message stats for monitoring.
// The counters are touched by the receive loop and read by the stats
// goroutine, so they must be atomic (plain ints are a data race).
type Stats struct {
	PacketsReceived atomic.Int64
	BytesReceived   atomic.Int64
	PacketsSent     atomic.Int64
}

func main() {
	// Resolve UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Fatalf("Failed to resolve address: %v", err)
	}

	// Create UDP connection
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer conn.Close()

	log.Printf("🚀 UDP Echo Server listening on %s", addr)
	log.Println("   Test with: echo 'hello' | nc -u localhost 9999")
	log.Println("   Press Ctrl+C to shutdown")

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Stats tracking (pointer: atomic types must never be copied)
	stats := &Stats{}

	// Stats printer goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("📊 Stats: %d packets received, %d bytes, %d responses sent",
					stats.PacketsReceived.Load(), stats.BytesReceived.Load(), stats.PacketsSent.Load())
			case <-sigChan:
				log.Println("\n🛑 Shutting down...")
				log.Printf("📊 Final Stats: %d packets, %d bytes, %d responses sent",
					stats.PacketsReceived.Load(), stats.BytesReceived.Load(), stats.PacketsSent.Load())
				conn.Close()
				os.Exit(0)
			}
		}
	}()

	// Buffer for incoming data
	buffer := make([]byte, 1024)

	// Main receive loop
	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			// The shutdown goroutine closes the socket, which makes every
			// further read fail immediately - stop instead of busy-looping.
			if errors.Is(err, net.ErrClosed) {
				return // listener closed during shutdown
			}
			log.Printf("Read error: %v", err)
			continue
		}

		// Update stats
		stats.PacketsReceived.Add(1)
		stats.BytesReceived.Add(int64(n))

		// Get message content
		message := string(buffer[:n])
		log.Printf("📨 Received from %s: %s", clientAddr, message)

		// Send response
		response := fmt.Sprintf("Echo: %s", message)
		_, err = conn.WriteToUDP([]byte(response), clientAddr)
		if err != nil {
			log.Printf("Write error: %v", err)
			continue
		}
		stats.PacketsSent.Add(1)
	}
}
