// Package main implements a concurrent port scanner.
// This exercise teaches concurrent networking and goroutine pools.
//
// Learning objectives:
// - Dial with timeouts
// - Use goroutine pools for controlled concurrency
// - Aggregate results across goroutines
//
// Run: go run main.go -host scanme.nmap.org -start 1 -end 100
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"
)

// ScanResult holds the result of scanning a port
type ScanResult struct {
	Port int
	Open bool
}

func main() {
	// Parse command line flags
	host := flag.String("host", "localhost", "Target host to scan")
	startPort := flag.Int("start", 1, "Start port")
	endPort := flag.Int("end", 1024, "End port")
	timeout := flag.Duration("timeout", 500*time.Millisecond, "Connection timeout")
	workers := flag.Int("workers", 100, "Number of concurrent workers")
	flag.Parse()

	// Validate the range before scanning: an inverted range makes the summary
	// count negative, and ports outside 1-65535 can never be dialled.
	if *startPort < 1 || *endPort > 65535 || *endPort < *startPort {
		log.Printf("invalid port range %d-%d: start must be >= 1, end must be <= 65535, and end must not be less than start",
			*startPort, *endPort)
		os.Exit(2)
	}

	log.Printf("🔍 Scanning %s ports %d-%d", *host, *startPort, *endPort)
	log.Printf("   Timeout: %v, Workers: %d", *timeout, *workers)

	startTime := time.Now()

	// Scan ports
	results := scanPorts(*host, *startPort, *endPort, *timeout, *workers)

	elapsed := time.Since(startTime)

	// Display results
	fmt.Println("\n📊 Results:")
	fmt.Println("─────────────────────────────────")

	if len(results) == 0 {
		fmt.Println("No open ports found")
	} else {
		// Sort by port number
		sort.Slice(results, func(i, j int) bool {
			return results[i].Port < results[j].Port
		})

		for _, r := range results {
			service := getServiceName(r.Port)
			fmt.Printf("  Port %5d: OPEN  (%s)\n", r.Port, service)
		}
	}

	fmt.Println("─────────────────────────────────")
	fmt.Printf("Scan completed in %v\n", elapsed)
	fmt.Printf("Open ports: %d/%d\n", len(results), *endPort-*startPort+1)
}

func scanPorts(host string, startPort, endPort int, timeout time.Duration, workers int) []ScanResult {
	// Channel for ports to scan
	ports := make(chan int, 100)

	// Channel for results
	results := make(chan ScanResult, 100)

	// WaitGroup for workers
	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range ports {
				result := scanPort(host, port, timeout)
				if result.Open {
					results <- result
				}
			}
		}()
	}

	// Send ports to workers
	go func() {
		for port := startPort; port <= endPort; port++ {
			ports <- port
		}
		close(ports)
	}()

	// Wait for workers and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var openPorts []ScanResult
	for result := range results {
		openPorts = append(openPorts, result)
	}

	return openPorts
}

func scanPort(host string, port int, timeout time.Duration) ScanResult {
	// JoinHostPort brackets IPv6 literals ("::1" -> "[::1]:80");
	// a plain Sprintf("%s:%d") produces an unusable address.
	address := net.JoinHostPort(host, strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return ScanResult{Port: port, Open: false}
	}
	defer conn.Close()

	return ScanResult{Port: port, Open: true}
}

// getServiceName returns common service names for well-known ports
func getServiceName(port int) string {
	services := map[int]string{
		21:    "FTP",
		22:    "SSH",
		23:    "Telnet",
		25:    "SMTP",
		53:    "DNS",
		80:    "HTTP",
		110:   "POP3",
		143:   "IMAP",
		443:   "HTTPS",
		445:   "SMB",
		993:   "IMAPS",
		995:   "POP3S",
		3306:  "MySQL",
		3389:  "RDP",
		5432:  "PostgreSQL",
		6379:  "Redis",
		8080:  "HTTP-Alt",
		8443:  "HTTPS-Alt",
		27017: "MongoDB",
	}

	if name, ok := services[port]; ok {
		return name
	}
	return "unknown"
}
