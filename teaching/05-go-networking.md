# Module 5: Go Network Programming

## Essential Packages Overview

```go
import (
    "net"           // Core networking (TCP/UDP/IP)
    "net/http"      // HTTP client/server
    "context"       // Timeouts and cancellation
    "io"            // I/O primitives
    "bufio"         // Buffered I/O
    "time"          // Durations and timeouts
)
```

---

## TCP Socket Programming

### Simple TCP Server

```go
package main

import (
    "bufio"
    "fmt"
    "net"
)

func main() {
    // Listen on port 8080
    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }
    defer listener.Close()
    
    fmt.Println("Server listening on :8080")
    
    for {
        // Accept new connection
        conn, err := listener.Accept()
        if err != nil {
            fmt.Printf("Accept error: %v\n", err)
            continue
        }
        
        // Handle in goroutine (concurrent)
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) {
    defer conn.Close()
    
    // Get client address
    clientAddr := conn.RemoteAddr().String()
    fmt.Printf("Client connected: %s\n", clientAddr)
    
    // Read data
    reader := bufio.NewReader(conn)
    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            fmt.Printf("Client %s disconnected\n", clientAddr)
            return
        }
        
        fmt.Printf("Received from %s: %s", clientAddr, message)
        
        // Echo back
        conn.Write([]byte("Echo: " + message))
    }
}
```

### TCP Client

```go
package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
)

func main() {
    // Connect to server
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    fmt.Println("Connected to server")
    
    // Send messages from stdin
    reader := bufio.NewReader(os.Stdin)
    serverReader := bufio.NewReader(conn)
    
    for {
        fmt.Print("Enter message: ")
        message, _ := reader.ReadString('\n')
        
        // Send to server
        conn.Write([]byte(message))
        
        // Read response
        response, _ := serverReader.ReadString('\n')
        fmt.Printf("Server: %s", response)
    }
}
```

---

## UDP Socket Programming

### UDP Server

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    // Create UDP address
    addr, err := net.ResolveUDPAddr("udp", ":9999")
    if err != nil {
        panic(err)
    }
    
    // Start listening
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    fmt.Println("UDP Server listening on :9999")
    
    buffer := make([]byte, 1024)
    
    for {
        // Read packet
        n, clientAddr, err := conn.ReadFromUDP(buffer)
        if err != nil {
            fmt.Printf("Read error: %v\n", err)
            continue
        }
        
        message := string(buffer[:n])
        fmt.Printf("Received from %s: %s\n", clientAddr, message)
        
        // Send response
        response := []byte("ACK: " + message)
        conn.WriteToUDP(response, clientAddr)
    }
}
```

---

## Connection Timeouts and Context

```go
package main

import (
    "context"
    "fmt"
    "net"
    "time"
)

func dialWithTimeout(address string) (net.Conn, error) {
    // Method 1: Simple timeout
    conn, err := net.DialTimeout("tcp", address, 5*time.Second)
    if err != nil {
        return nil, fmt.Errorf("dial timeout: %w", err)
    }
    return conn, nil
}

func dialWithContext(ctx context.Context, address string) (net.Conn, error) {
    // Method 2: Context-based (recommended)
    var d net.Dialer
    conn, err := d.DialContext(ctx, "tcp", address)
    if err != nil {
        return nil, fmt.Errorf("dial failed: %w", err)
    }
    return conn, nil
}

func main() {
    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    conn, err := dialWithContext(ctx, "google.com:443")
    if err != nil {
        fmt.Printf("Failed: %v\n", err)
        return
    }
    defer conn.Close()
    
    fmt.Printf("Connected to: %s\n", conn.RemoteAddr())
}
```

---

## ICMP Ping Implementation

This is core to your health monitoring!

```go
package main

import (
    "fmt"
    "net"
    "os"
    "time"
    
    "golang.org/x/net/icmp"
    "golang.org/x/net/ipv4"
)

func ping(host string, timeout time.Duration) (time.Duration, error) {
    // Resolve address
    dst, err := net.ResolveIPAddr("ip4", host)
    if err != nil {
        return 0, err
    }
    
    // Create ICMP connection (requires root)
    conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
    if err != nil {
        return 0, err
    }
    defer conn.Close()
    
    // Build ICMP message
    msg := icmp.Message{
        Type: ipv4.ICMPTypeEcho,
        Code: 0,
        Body: &icmp.Echo{
            ID:   os.Getpid() & 0xffff,
            Seq:  1,
            Data: []byte("HELLO"),
        },
    }
    
    msgBytes, err := msg.Marshal(nil)
    if err != nil {
        return 0, err
    }
    
    // Set deadline
    conn.SetDeadline(time.Now().Add(timeout))
    
    // Send
    start := time.Now()
    _, err = conn.WriteTo(msgBytes, dst)
    if err != nil {
        return 0, err
    }
    
    // Receive reply
    reply := make([]byte, 1500)
    _, _, err = conn.ReadFrom(reply)
    if err != nil {
        return 0, err
    }
    
    rtt := time.Since(start)
    return rtt, nil
}

func main() {
    hosts := []string{"8.8.8.8", "1.1.1.1", "9.9.9.9"}
    
    for _, host := range hosts {
        rtt, err := ping(host, 2*time.Second)
        if err != nil {
            fmt.Printf("%s: FAILED (%v)\n", host, err)
        } else {
            fmt.Printf("%s: %.2fms\n", host, float64(rtt.Microseconds())/1000)
        }
    }
}
```

---

## HTTP Health Checks

```go
package main

import (
    "context"
    "fmt"
    "net"
    "net/http"
    "time"
)

type HealthChecker struct {
    client *http.Client
}

func NewHealthChecker(sourceIP string) *HealthChecker {
    // Create custom dialer bound to specific interface
    dialer := &net.Dialer{
        Timeout:   5 * time.Second,
        LocalAddr: &net.TCPAddr{IP: net.ParseIP(sourceIP)},
    }
    
    transport := &http.Transport{
        DialContext: dialer.DialContext,
    }
    
    return &HealthChecker{
        client: &http.Client{
            Transport: transport,
            Timeout:   10 * time.Second,
        },
    }
}

func (hc *HealthChecker) Check(url string) (time.Duration, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return 0, err
    }
    
    start := time.Now()
    resp, err := hc.client.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()
    
    latency := time.Since(start)
    
    if resp.StatusCode < 200 || resp.StatusCode >= 400 {
        return latency, fmt.Errorf("bad status: %d", resp.StatusCode)
    }
    
    return latency, nil
}

func main() {
    checker := NewHealthChecker("0.0.0.0") // Or specific interface IP
    
    urls := []string{
        "https://www.google.com",
        "https://www.cloudflare.com",
        "https://www.amazon.com",
    }
    
    for _, url := range urls {
        latency, err := checker.Check(url)
        if err != nil {
            fmt.Printf("%s: FAILED (%v)\n", url, err)
        } else {
            fmt.Printf("%s: %v\n", url, latency)
        }
    }
}
```

---

## Binding to Specific Interface

Critical for multi-WAN - send traffic through specific ISP!

```go
package main

import (
    "fmt"
    "net"
    "time"
)

// Get interface IP
func getInterfaceIP(ifaceName string) (net.IP, error) {
    iface, err := net.InterfaceByName(ifaceName)
    if err != nil {
        return nil, err
    }
    
    addrs, err := iface.Addrs()
    if err != nil {
        return nil, err
    }
    
    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok {
            if ipv4 := ipnet.IP.To4(); ipv4 != nil {
                return ipv4, nil
            }
        }
    }
    
    return nil, fmt.Errorf("no IPv4 address found")
}

// Dial through specific interface
func dialViaInterface(ifaceName, address string) (net.Conn, error) {
    ip, err := getInterfaceIP(ifaceName)
    if err != nil {
        return nil, err
    }
    
    dialer := net.Dialer{
        LocalAddr: &net.TCPAddr{IP: ip},
        Timeout:   5 * time.Second,
    }
    
    return dialer.Dial("tcp", address)
}

func main() {
    // Connect to google via eth0 (ISP1)
    conn1, err := dialViaInterface("eth0", "google.com:443")
    if err != nil {
        fmt.Printf("eth0 failed: %v\n", err)
    } else {
        fmt.Printf("Connected via eth0: local=%s remote=%s\n",
            conn1.LocalAddr(), conn1.RemoteAddr())
        conn1.Close()
    }
    
    // Connect via eth1 (ISP2)
    conn2, err := dialViaInterface("eth1", "google.com:443")
    if err != nil {
        fmt.Printf("eth1 failed: %v\n", err)
    } else {
        fmt.Printf("Connected via eth1: local=%s remote=%s\n",
            conn2.LocalAddr(), conn2.RemoteAddr())
        conn2.Close()
    }
}
```

---

## Project: Port Scanner

```go
package main

import (
    "fmt"
    "net"
    "sync"
    "time"
)

func scanPort(host string, port int, timeout time.Duration) bool {
    address := fmt.Sprintf("%s:%d", host, port)
    conn, err := net.DialTimeout("tcp", address, timeout)
    if err != nil {
        return false
    }
    conn.Close()
    return true
}

func main() {
    host := "scanme.nmap.org"
    timeout := 500 * time.Millisecond
    
    var wg sync.WaitGroup
    openPorts := make(chan int, 100)
    
    // Scan ports 1-1024
    for port := 1; port <= 1024; port++ {
        wg.Add(1)
        go func(p int) {
            defer wg.Done()
            if scanPort(host, p, timeout) {
                openPorts <- p
            }
        }(port)
    }
    
    // Wait and close channel
    go func() {
        wg.Wait()
        close(openPorts)
    }()
    
    // Collect results
    fmt.Printf("Open ports on %s:\n", host)
    for port := range openPorts {
        fmt.Printf("  Port %d: OPEN\n", port)
    }
}
```

---

## Key Takeaways

1. **net.Listen** / **net.Dial** for TCP connections
2. **net.ListenUDP** / **ReadFromUDP** for UDP
3. **Context** for timeouts and cancellation
4. **Bind to interface** using LocalAddr in Dialer
5. **ICMP** for ping (requires root/raw sockets)
6. **Goroutines** for concurrent connections

---

## Next Module
→ [06-ebpf-fundamentals.md](./06-ebpf-fundamentals.md): eBPF for high-performance packet processing
