// Package main implements a concurrent TCP echo server.
// This exercise teaches TCP socket fundamentals in Go.
//
// Learning objectives:
// - Create a TCP listener
// - Accept connections concurrently
// - Handle client data with proper error handling
// - Graceful shutdown with signals
//
// Run: go run main.go
// Test: nc localhost 8080 (then type messages)
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const addr = ":8080"

func main() {
	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start TCP listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", addr, err)
	}
	defer listener.Close()

	log.Printf("🚀 TCP Echo Server listening on %s", addr)
	log.Println("   Connect with: nc localhost 8080")
	log.Println("   Press Ctrl+C to shutdown")

	// Track active connections for graceful shutdown
	var wg sync.WaitGroup

	// Handle shutdown in goroutine
	go func() {
		<-sigChan
		log.Println("\n🛑 Shutting down...")
		cancel()
		listener.Close()
	}()

	// Accept connections loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				// Context cancelled, graceful shutdown
				wg.Wait()
				log.Println("✅ Server shutdown complete")
				return
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConnection(ctx, c)
		}(conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	log.Printf("📥 Client connected: %s", clientAddr)

	// Send welcome message
	fmt.Fprintf(conn, "Welcome to TCP Echo Server!\n")
	fmt.Fprintf(conn, "Type messages and I'll echo them back.\n")
	fmt.Fprintf(conn, "Type 'quit' to disconnect.\n\n")

	reader := bufio.NewReader(conn)

	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			fmt.Fprintf(conn, "Server shutting down. Goodbye!\n")
			return
		default:
		}

		// Read line from client
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("📤 Client disconnected: %s", clientAddr)
			return
		}

		// Trim and check for quit command
		message = message[:len(message)-1] // Remove newline
		if message == "quit" {
			fmt.Fprintf(conn, "Goodbye!\n")
			log.Printf("📤 Client quit: %s", clientAddr)
			return
		}

		// Echo back with prefix
		response := fmt.Sprintf("Echo: %s\n", message)
		conn.Write([]byte(response))

		log.Printf("💬 [%s] %s", clientAddr, message)
	}
}
