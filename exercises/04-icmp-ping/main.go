// Package main implements ICMP ping functionality.
// This exercise teaches raw socket programming and ICMP protocol.
//
// Learning objectives:
// - Work with ICMP protocol
// - Parse network addresses
// - Measure round-trip time
// - Handle privileged operations (requires root/sudo)
//
// Run: sudo go run main.go -host 8.8.8.8 -count 4
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	protocolICMP = 1
)

// PingResult holds statistics for a ping session
type PingResult struct {
	Host        string
	PacketsSent int
	PacketsRecv int
	MinRTT      time.Duration
	MaxRTT      time.Duration
	AvgRTT      time.Duration
	TotalRTT    time.Duration
}

func main() {
	// Parse flags
	host := flag.String("host", "8.8.8.8", "Host to ping")
	count := flag.Int("count", 4, "Number of pings to send")
	timeout := flag.Duration("timeout", 2*time.Second, "Timeout per ping")
	interval := flag.Duration("interval", 1*time.Second, "Interval between pings")
	flag.Parse()

	// Check for root privileges
	if os.Geteuid() != 0 {
		log.Println("⚠️  Warning: ICMP requires root privileges")
		log.Println("   Run with: sudo go run main.go")
		os.Exit(1)
	}

	// Resolve host
	dst, err := net.ResolveIPAddr("ip4", *host)
	if err != nil {
		log.Fatalf("Failed to resolve %s: %v", *host, err)
	}

	fmt.Printf("PING %s (%s)\n", *host, dst.IP)
	fmt.Println("─────────────────────────────────")

	result := &PingResult{
		Host:   *host,
		MinRTT: time.Hour, // Start with large value
	}

	// Send pings
	for i := 0; i < *count; i++ {
		rtt, err := ping(dst, i+1, *timeout)
		result.PacketsSent++

		if err != nil {
			fmt.Printf("Request timeout for seq %d\n", i+1)
		} else {
			result.PacketsRecv++
			result.TotalRTT += rtt

			if rtt < result.MinRTT {
				result.MinRTT = rtt
			}
			if rtt > result.MaxRTT {
				result.MaxRTT = rtt
			}

			fmt.Printf("Reply from %s: seq=%d time=%.2fms\n",
				dst.IP, i+1, float64(rtt.Microseconds())/1000)
		}

		// Wait between pings (except for last one)
		if i < *count-1 {
			time.Sleep(*interval)
		}
	}

	// Print statistics
	fmt.Println("─────────────────────────────────")
	fmt.Printf("\n--- %s ping statistics ---\n", *host)

	// Guard the division: with -count 0 nothing was sent and 0/0 would print NaN.
	lossPercent := 0.0
	if result.PacketsSent > 0 {
		lossPercent = float64(result.PacketsSent-result.PacketsRecv) / float64(result.PacketsSent) * 100
	}
	fmt.Printf("%d packets transmitted, %d received, %.1f%% packet loss\n",
		result.PacketsSent, result.PacketsRecv, lossPercent)

	if result.PacketsRecv > 0 {
		result.AvgRTT = result.TotalRTT / time.Duration(result.PacketsRecv)
		fmt.Printf("rtt min/avg/max = %.2f/%.2f/%.2f ms\n",
			float64(result.MinRTT.Microseconds())/1000,
			float64(result.AvgRTT.Microseconds())/1000,
			float64(result.MaxRTT.Microseconds())/1000)
	}
}

func ping(dst *net.IPAddr, seq int, timeout time.Duration) (time.Duration, error) {
	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return 0, fmt.Errorf("listen error: %w", err)
	}
	defer conn.Close()

	// Build ICMP echo request
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   os.Getpid() & 0xffff,
			Seq:  seq,
			Data: []byte("PING from Go exercise!"),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return 0, fmt.Errorf("marshal error: %w", err)
	}

	// Set deadline
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return 0, fmt.Errorf("set deadline: %w", err)
	}

	// Send
	start := time.Now()
	if _, err := conn.WriteTo(msgBytes, dst); err != nil {
		return 0, fmt.Errorf("write error: %w", err)
	}

	// Receive reply.
	// A raw ICMP socket receives *every* ICMP message the host gets, not just
	// ours - a system `ping` running in another terminal delivers its replies
	// here too. So keep reading until a packet matches all four criteria:
	// echo reply, our ID, our sequence number, and the host we pinged.
	// The deadline set above turns a genuinely lost reply into a read error.
	reply := make([]byte, 1500)
	for {
		n, peer, err := conn.ReadFrom(reply)
		if err != nil {
			return 0, fmt.Errorf("read error: %w", err)
		}

		rm, err := icmp.ParseMessage(protocolICMP, reply[:n])
		if err != nil {
			continue // not a well-formed ICMP message
		}

		if rm.Type != ipv4.ICMPTypeEchoReply {
			continue // e.g. Destination Unreachable or Time Exceeded
		}

		echo, ok := rm.Body.(*icmp.Echo)
		if !ok {
			continue
		}

		if echo.ID != os.Getpid()&0xffff || echo.Seq != seq {
			continue // another process's ping, or a stale sequence of ours
		}

		if peer.String() != dst.IP.String() {
			continue // right identifier, wrong host
		}

		return time.Since(start), nil
	}
}
