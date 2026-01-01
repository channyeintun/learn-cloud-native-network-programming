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
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const addr = ":9999"

// Message stats for monitoring
type Stats struct {
	PacketsReceived int
	BytesReceived   int
	PacketsSent     int
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

	// Stats tracking
	stats := &Stats{}

	// Stats printer goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Printf("📊 Stats: %d packets received, %d bytes, %d responses sent",
					stats.PacketsReceived, stats.BytesReceived, stats.PacketsSent)
			case <-sigChan:
				log.Println("\n🛑 Shutting down...")
				log.Printf("📊 Final Stats: %d packets, %d bytes",
					stats.PacketsReceived, stats.BytesReceived)
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
			// Check if it's a shutdown-related error (connection closed)
			select {
			case <-sigChan:
				// Already handled in goroutine
				return
			default:
				log.Printf("Read error: %v", err)
				continue
			}
		}

		// Update stats
		stats.PacketsReceived++
		stats.BytesReceived += n

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
		stats.PacketsSent++
	}
}
