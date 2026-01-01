// Package main implements an HTTP health checker for multiple endpoints.
// This exercise teaches HTTP client usage and concurrent health monitoring.
//
// Learning objectives:
// - Configure HTTP clients with timeouts
// - Bind to specific network interfaces
// - Monitor multiple endpoints concurrently
// - Parse and validate responses
//
// Run: go run main.go
// Or:  go run main.go -config endpoints.json
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Endpoint represents a health check target
type Endpoint struct {
	Name           string        `json:"name"`
	URL            string        `json:"url"`
	Interval       time.Duration `json:"interval"`
	Timeout        time.Duration `json:"timeout"`
	ExpectedStatus int           `json:"expected_status"`
}

// HealthStatus represents the current health of an endpoint
type HealthStatus struct {
	Endpoint  *Endpoint
	Healthy   bool
	Latency   time.Duration
	LastCheck time.Time
	Error     string
}

// Default endpoints if no config file provided
var defaultEndpoints = []Endpoint{
	{
		Name:           "Google",
		URL:            "https://www.google.com",
		Interval:       5 * time.Second,
		Timeout:        3 * time.Second,
		ExpectedStatus: 200,
	},
	{
		Name:           "Cloudflare",
		URL:            "https://www.cloudflare.com",
		Interval:       5 * time.Second,
		Timeout:        3 * time.Second,
		ExpectedStatus: 200,
	},
	{
		Name:           "GitHub",
		URL:            "https://api.github.com",
		Interval:       5 * time.Second,
		Timeout:        3 * time.Second,
		ExpectedStatus: 200,
	},
	{
		Name:           "Example (should work)",
		URL:            "https://example.com",
		Interval:       5 * time.Second,
		Timeout:        3 * time.Second,
		ExpectedStatus: 200,
	},
	{
		Name:           "Bad Endpoint (should fail)",
		URL:            "https://this-does-not-exist-12345.com",
		Interval:       10 * time.Second,
		Timeout:        2 * time.Second,
		ExpectedStatus: 200,
	},
}

// HealthChecker manages health checks for multiple endpoints
type HealthChecker struct {
	endpoints []Endpoint
	client    *http.Client
	statuses  map[string]*HealthStatus
	mu        sync.RWMutex
}

func main() {
	configFile := flag.String("config", "", "JSON config file with endpoints")
	interfaceName := flag.String("interface", "", "Network interface to bind to (optional)")
	flag.Parse()

	// Load endpoints
	endpoints := defaultEndpoints
	if *configFile != "" {
		loaded, err := loadEndpoints(*configFile)
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
		endpoints = loaded
	}

	// Create HTTP client
	client := createClient(*interfaceName)

	// Initialize health checker
	hc := &HealthChecker{
		endpoints: endpoints,
		client:    client,
		statuses:  make(map[string]*HealthStatus),
	}

	// Setup context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n🛑 Shutting down...")
		cancel()
	}()

	fmt.Println("🏥 Health Checker Starting")
	fmt.Println("─────────────────────────────────────────────────")
	fmt.Printf("   Monitoring %d endpoints\n", len(endpoints))
	fmt.Println("   Press Ctrl+C to stop")
	fmt.Println("─────────────────────────────────────────────────")

	// Start health checks
	var wg sync.WaitGroup
	for i := range endpoints {
		wg.Add(1)
		go func(ep *Endpoint) {
			defer wg.Done()
			hc.monitorEndpoint(ctx, ep)
		}(&endpoints[i])
	}

	// Start status display
	go hc.displayStatus(ctx)

	wg.Wait()
	fmt.Println("✅ Health checker stopped")
}

func (hc *HealthChecker) monitorEndpoint(ctx context.Context, ep *Endpoint) {
	ticker := time.NewTicker(ep.Interval)
	defer ticker.Stop()

	// Initial check
	hc.checkEndpoint(ctx, ep)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.checkEndpoint(ctx, ep)
		}
	}
}

func (hc *HealthChecker) checkEndpoint(ctx context.Context, ep *Endpoint) {
	reqCtx, cancel := context.WithTimeout(ctx, ep.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "GET", ep.URL, nil)
	if err != nil {
		hc.updateStatus(ep, false, 0, err.Error())
		return
	}

	start := time.Now()
	resp, err := hc.client.Do(req)
	latency := time.Since(start)

	if err != nil {
		hc.updateStatus(ep, false, latency, err.Error())
		return
	}
	defer resp.Body.Close()

	healthy := resp.StatusCode == ep.ExpectedStatus
	errMsg := ""
	if !healthy {
		errMsg = fmt.Sprintf("status %d (expected %d)", resp.StatusCode, ep.ExpectedStatus)
	}

	hc.updateStatus(ep, healthy, latency, errMsg)
}

func (hc *HealthChecker) updateStatus(ep *Endpoint, healthy bool, latency time.Duration, errMsg string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.statuses[ep.Name] = &HealthStatus{
		Endpoint:  ep,
		Healthy:   healthy,
		Latency:   latency,
		LastCheck: time.Now(),
		Error:     errMsg,
	}
}

func (hc *HealthChecker) displayStatus(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hc.printStatus()
		}
	}
}

func (hc *HealthChecker) printStatus() {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	fmt.Println("\n📊 Health Status:")
	for _, ep := range hc.endpoints {
		status, ok := hc.statuses[ep.Name]
		if !ok {
			fmt.Printf("   ⏳ %-25s checking...\n", ep.Name)
			continue
		}

		icon := "✅"
		if !status.Healthy {
			icon = "❌"
		}

		latencyStr := fmt.Sprintf("%.0fms", float64(status.Latency.Microseconds())/1000)
		if status.Error != "" {
			fmt.Printf("   %s %-25s %s (error: %s)\n", icon, ep.Name, latencyStr, status.Error)
		} else {
			fmt.Printf("   %s %-25s %s\n", icon, ep.Name, latencyStr)
		}
	}
}

func createClient(interfaceName string) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	// Bind to specific interface if provided
	if interfaceName != "" {
		localAddr := getInterfaceAddr(interfaceName)
		if localAddr != nil {
			dialer := &net.Dialer{
				LocalAddr: localAddr,
				Timeout:   5 * time.Second,
			}
			transport.DialContext = dialer.DialContext
			log.Printf("Bound to interface: %s (%s)", interfaceName, localAddr)
		}
	}

	return &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}
}

func getInterfaceAddr(name string) *net.TCPAddr {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		log.Printf("Interface %s not found: %v", name, err)
		return nil
	}

	addrs, err := iface.Addrs()
	if err != nil {
		log.Printf("Failed to get addresses for %s: %v", name, err)
		return nil
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.To4() != nil {
			return &net.TCPAddr{IP: ipnet.IP}
		}
	}

	return nil
}

func loadEndpoints(filename string) ([]Endpoint, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var endpoints []Endpoint
	if err := json.Unmarshal(data, &endpoints); err != nil {
		return nil, err
	}

	// Set defaults
	for i := range endpoints {
		if endpoints[i].Interval == 0 {
			endpoints[i].Interval = 5 * time.Second
		}
		if endpoints[i].Timeout == 0 {
			endpoints[i].Timeout = 3 * time.Second
		}
		if endpoints[i].ExpectedStatus == 0 {
			endpoints[i].ExpectedStatus = 200
		}
	}

	return endpoints, nil
}
