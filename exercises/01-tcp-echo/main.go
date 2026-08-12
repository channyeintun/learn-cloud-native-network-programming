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
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
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

	// Accept connections loop. retryDelay implements the exponential backoff
	// net/http.Server.Serve uses: without it a persistent transient failure
	// (fd exhaustion) would spin a CPU core and flood the log.
	var retryDelay time.Duration

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
			}

			if isTemporaryAcceptError(err) {
				if retryDelay == 0 {
					retryDelay = 5 * time.Millisecond
				} else {
					retryDelay *= 2
				}
				if retryDelay > time.Second {
					retryDelay = time.Second
				}
				log.Printf("Accept error: %v; retrying in %v", err, retryDelay)
				time.Sleep(retryDelay)
				continue
			}

			// Not recoverable: stop accepting and drain in-flight connections.
			log.Printf("Fatal accept error: %v", err)
			wg.Wait()
			return
		}
		retryDelay = 0

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConnection(ctx, c)
		}(conn)
	}
}

// isTemporaryAcceptError reports whether an Accept failure is worth retrying.
// Running out of file descriptors (EMFILE/ENFILE) or losing a half-open
// connection before it is accepted (ECONNABORTED) clears up on its own; a
// timeout does too. Anything else means the listener is unusable.
func isTemporaryAcceptError(err error) bool {
	if errors.Is(err, syscall.EMFILE) || errors.Is(err, syscall.ENFILE) ||
		errors.Is(err, syscall.ECONNABORTED) || errors.Is(err, syscall.EINTR) {
		return true
	}
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Say goodbye and unblock the blocking Read below when the server is
	// shutting down. The farewell must be written here, before Close: the
	// read loop is parked in ReadString and cannot send it itself.
	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			fmt.Fprintf(conn, "Server shutting down. Goodbye!\n")
			conn.Close()
		case <-done:
		}
	}()

	clientAddr := conn.RemoteAddr().String()
	log.Printf("📥 Client connected: %s", clientAddr)

	// Send welcome message
	fmt.Fprintf(conn, "Welcome to TCP Echo Server!\n")
	fmt.Fprintf(conn, "Type messages and I'll echo them back.\n")
	fmt.Fprintf(conn, "Type 'quit' to disconnect.\n\n")

	reader := bufio.NewReader(conn)

	for {
		// Read line from client. On shutdown the watcher goroutine above
		// closes the socket, which makes this call return an error.
		message, err := reader.ReadString('\n')
		if err != nil {
			// ReadString returns whatever it read *together with* the error,
			// so a final line sent without a trailing newline still has to be
			// echoed before we give up on the connection.
			if last := strings.TrimRight(message, "\r\n"); last != "" {
				log.Printf("💬 [%s] %s", clientAddr, last)
				if _, werr := conn.Write([]byte(fmt.Sprintf("Echo: %s\n", last))); werr != nil {
					log.Printf("write to %s: %v", clientAddr, werr)
				}
			}
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				log.Printf("📤 Client disconnected: %s", clientAddr)
			} else {
				log.Printf("📤 Client %s read error: %v", clientAddr, err)
			}
			return
		}

		// Trim the line ending and check for the quit command. TrimRight also
		// strips the '\r' that telnet, PuTTY and other CRLF clients send, so
		// "quit" matches for them too.
		message = strings.TrimRight(message, "\r\n")
		if message == "quit" {
			fmt.Fprintf(conn, "Goodbye!\n")
			log.Printf("📤 Client quit: %s", clientAddr)
			return
		}

		// Echo back with prefix
		response := fmt.Sprintf("Echo: %s\n", message)
		if _, err := conn.Write([]byte(response)); err != nil {
			log.Printf("write to %s: %v", clientAddr, err)
			return
		}

		log.Printf("💬 [%s] %s", clientAddr, message)
	}
}
