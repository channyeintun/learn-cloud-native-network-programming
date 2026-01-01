# Module 6: eBPF Fundamentals

## 📊 Visual Learning

![eBPF Architecture](./images/ebpf_architecture.png)

![XDP Hook Points](./images/xdp_hook_points.png)

## What is the Linux Kernel?

Before understanding eBPF, it's essential to understand **what the kernel is** and why running code inside it is so powerful.

### 📊 Visual Overview

![Linux Kernel Architecture](./images/linux_kernel_overview.png)

### The Kernel Explained

The **Linux kernel** is the core of the operating system—it's the **bridge between your applications and the hardware**. Think of it as the central control system that manages everything.

| Layer | Description | Examples |
|-------|-------------|----------|
| **User Space** | Where your applications run | Browsers, servers, containers, your Go programs |
| **Kernel Space** | Privileged core of the OS | Process scheduler, network stack, device drivers |
| **Hardware** | Physical components | CPU, RAM, Network cards, Storage |

### What the Kernel Does

1. **Process Management** - Schedules which programs run on the CPU
2. **Memory Management** - Allocates and protects memory for each process
3. **Device Drivers** - Communicates with hardware (disks, network cards, etc.)
4. **Network Stack** - Handles all network traffic (the focus of this course!)
5. **File Systems** - Manages how data is stored and retrieved

### Why Kernel Access Matters for Networking

```
Traditional networking:

    Packet → Network Card → Kernel → iptables → Routing → Socket → Your App
                                ↑
                        Every packet goes through
                        kernel's network stack
```

**The problem:** When you need to customize network behavior, you traditionally had two options:

1. **User space processing** - Slow context switches between kernel and your app
2. **Kernel modules** - Dangerous, requires recompilation, can crash the system

**eBPF changes this** - it lets you run **safe**, **fast** code inside the kernel!

---

## What is eBPF?

eBPF (extended Berkeley Packet Filter) is a **revolutionary technology** that allows you to run sandboxed programs inside the Linux kernel without modifying kernel source code or loading kernel modules.

### Why eBPF Matters

```mermaid
flowchart TB
    subgraph traditional["Traditional Approach"]
        direction TB
        T_USER["User Space"] -->|"context switch - SLOW"| T_KERNEL["Kernel Space"]
        T_KERNEL -->|"Modify kernel or load module<br>(dangerous, requires reboot)"| T_NET["Network Stack"]
    end
    
    subgraph ebpf["eBPF Approach"]
        direction TB
        E_USER["User Space<br>(Your Go program)"] -->|"Load eBPF bytecode"| E_VERIFY["eBPF Verifier"]
        E_VERIFY -->|"Validates safety"| E_JIT["JIT Compiler"]
        E_JIT -->|"Native machine code"| E_HOOKS["Kernel Hooks<br>(XDP, TC, Socket, etc.)"]
    end
    
    style traditional fill:#ffebee
    style ebpf fill:#e8f5e9
    style E_HOOKS fill:#c8e6c9
```

---

## eBPF Architecture

### The Big Picture

```mermaid
flowchart TB
    subgraph userspace["USER SPACE"]
        GO["Your Go Program"] --> LOADER["cilium/ebpf loader"]
        BPFTOOL["bpftool (debug)"]
        PROM["Prometheus/Grafana"]
        MAPS_US["eBPF Maps<br>(shared memory)"]
        PROM -.->|"Read maps"| MAPS_US
    end
    
    subgraph kernel["KERNEL SPACE"]
        LOADER --> VERIFIER["Verifier"]
        VERIFIER -->|"Rejects unsafe"| JIT["JIT Compiler"]
        JIT --> VM["eBPF Virtual Machine"]
        
        subgraph hooks["Hook Points"]
            XDP["XDP"]
            TC["TC"]
            SOCKET["Socket"]
            KPROBE["kprobe"]
        end
        
        VM --> hooks
        VM <-->|"Read/Write"| MAPS_K["eBPF Maps"]
        MAPS_K <-.-> MAPS_US
    end
    
    subgraph netstack["Network Stack"]
        PACKET["Packet"] --> XDP2["XDP"] --> TC2["TC"] --> NETFILTER["Netfilter"] --> ROUTING["Routing"] --> SOCK["Socket"] --> APP["App"]
    end
    
    style userspace fill:#e3f2fd
    style kernel fill:#fff3e0
    style hooks fill:#c8e6c9
```

### eBPF Hook Points for Networking

| Hook | Location | Direction | Best For |
|------|----------|-----------|----------|
| **XDP** | NIC driver | Ingress only | DDoS, Load balancing, Firewall |
| **TC ingress** | After XDP | Ingress | Traffic shaping, Policing |
| **TC egress** | Before NIC | Egress | Rate limiting, Shaping |
| **Socket** | L4 | Both | Per-connection filtering |
| **cgroup/sock** | cgroup | Both | Container networking |
| **sk_msg** | Socket | Both | Proxy, TLS interception |

---

## eBPF Program Types

### 1. XDP (eXpress Data Path)

The fastest hook point - runs before the kernel allocates any memory for the packet.

```c
SEC("xdp")
int xdp_drop_icmp(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    // Parse Ethernet header
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    // Only process IPv4
    if (eth->h_proto != htons(ETH_P_IP))
        return XDP_PASS;
    
    // Parse IP header
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    // Drop ICMP packets
    if (ip->protocol == IPPROTO_ICMP) {
        return XDP_DROP;
    }
    
    return XDP_PASS;
}
```

**XDP Return Values:**
- `XDP_DROP` - Drop packet silently
- `XDP_PASS` - Pass to normal network stack
- `XDP_TX` - Bounce back out same interface
- `XDP_REDIRECT` - Send to another interface/CPU
- `XDP_ABORTED` - Drop with error trace

### 2. TC (Traffic Control)

More features than XDP (can modify packets, access full sk_buff).

```c
SEC("tc")
int tc_egress(struct __sk_buff *skb) {
    void *data = (void *)(long)skb->data;
    void *data_end = (void *)(long)skb->data_end;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return TC_ACT_OK;
    
    // Can modify packet
    // Can access skb->mark, skb->priority, etc.
    
    return TC_ACT_OK;  // Pass packet
}
```

**TC Return Values:**
- `TC_ACT_OK` - Continue processing
- `TC_ACT_SHOT` - Drop packet
- `TC_ACT_REDIRECT` - Redirect packet

### 3. Socket Filter

Attach to a specific socket.

```c
SEC("socket")
int socket_filter(struct __sk_buff *skb) {
    // Return 0 to drop, non-zero to keep
    return skb->len;
}
```

---

## eBPF Maps

Maps are key-value stores that allow:
- eBPF programs to share data with each other
- User space to communicate with eBPF programs
- Persistent state between packet processing

### Map Types

```c
// Hash map - O(1) lookup
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10000);
    __type(key, __u32);
    __type(value, __u64);
} connection_count SEC(".maps");

// Array - O(1) lookup by index
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 256);
    __type(key, __u32);
    __type(value, struct stats);
} protocol_stats SEC(".maps");

// LRU Hash - Automatic eviction
struct {
    __uint(type, BPF_MAP_TYPE_LRU_HASH);
    __uint(max_entries, 100000);
    __type(key, struct conn_key);
    __type(value, struct conn_info);
} connections SEC(".maps");

// Per-CPU Array - No locking needed
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct counters);
} counters SEC(".maps");

// Ring buffer - Efficient user space communication
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 256 * 1024);  // 256KB
} events SEC(".maps");
```

### Common Map Types for Networking

| Map Type | Use Case | Performance |
|----------|----------|-------------|
| `HASH` | Connection tracking | O(1) |
| `LRU_HASH` | Auto-evicting cache | O(1) |
| `ARRAY` | Protocol counters | O(1) |
| `PERCPU_ARRAY` | Per-CPU stats | Lock-free |
| `RINGBUF` | Event streaming | Efficient |
| `DEVMAP` | XDP_REDIRECT targets | Fast lookup |
| `CPUMAP` | XDP CPU steering | CPU affinity |

---

## Development with Go (cilium/ebpf)

### Project Structure

```
myproject/
├── bpf/
│   ├── headers/
│   │   ├── vmlinux.h          # Kernel types (generated)
│   │   └── bpf_helpers.h      # BPF helper definitions
│   └── program.c              # Your eBPF C code
├── main.go                     # Go loader and user space logic
├── program_bpfel.go            # Generated (little endian)
├── program_bpfel.o             # Compiled eBPF object
└── go.mod
```

### Step 1: Write eBPF Program (C)

```c
// bpf/program.c
//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

// Map to count packets per IP
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10000);
    __type(key, __u32);    // Source IP
    __type(value, __u64);  // Packet count
} packet_count SEC(".maps");

SEC("xdp")
int count_packets(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    // Parse Ethernet
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;
    
    // Parse IP
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    __u32 src_ip = ip->saddr;
    __u64 *count = bpf_map_lookup_elem(&packet_count, &src_ip);
    
    if (count) {
        __sync_fetch_and_add(count, 1);
    } else {
        __u64 initial = 1;
        bpf_map_update_elem(&packet_count, &src_ip, &initial, BPF_ANY);
    }
    
    return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";
```

### Step 2: Generate Go Bindings

```go
// main.go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang -cflags "-O2 -g -Wall" program ./bpf/program.c -- -I./bpf/headers
```

Run:
```bash
go generate ./...
```

This creates `program_bpfel.go` with:
- `programObjects` - compiled eBPF programs
- `programMaps` - map references
- `loadProgram()` - loader function

### Step 3: Load and Attach (Go)

```go
// main.go
package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang program ./bpf/program.c -- -I./bpf/headers

func main() {
    // Load compiled eBPF programs
    objs := programObjects{}
    if err := loadProgramObjects(&objs, nil); err != nil {
        log.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()
    
    // Get network interface
    iface, err := net.InterfaceByName("eth0")
    if err != nil {
        log.Fatalf("getting interface: %v", err)
    }
    
    // Attach XDP program
    xdpLink, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.CountPackets,
        Interface: iface.Index,
    })
    if err != nil {
        log.Fatalf("attaching XDP: %v", err)
    }
    defer xdpLink.Close()
    
    log.Printf("XDP program attached to %s", iface.Name)
    
    // Read map periodically
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()
    
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    
    for {
        select {
        case <-ticker.C:
            printStats(objs.PacketCount)
        case <-sig:
            log.Println("Shutting down...")
            return
        }
    }
}

func printStats(m *ebpf.Map) {
    var (
        key   uint32
        value uint64
    )
    
    iter := m.Iterate()
    for iter.Next(&key, &value) {
        ip := net.IPv4(
            byte(key), byte(key>>8), 
            byte(key>>16), byte(key>>24),
        )
        fmt.Printf("IP: %s, Packets: %d\n", ip, value)
    }
    fmt.Println("---")
}
```

---

## The eBPF Verifier

The verifier ensures your program is safe:

### What it Checks

1. **No infinite loops** - All loops must have bounded iterations
2. **Valid memory access** - All pointers must be bounds-checked
3. **Valid helper calls** - Only allowed helpers for program type
4. **Stack size limit** - Max 512 bytes
5. **Instruction count** - Max ~1 million instructions

### Common Verifier Errors

```c
// ERROR: Unbounded loop
for (int i = 0; i < n; i++) { }  // n is unknown

// FIX: Use bounded loop
#pragma unroll
for (int i = 0; i < 10; i++) { }

// Or use bpf_loop() helper (kernel 5.17+)
bpf_loop(100, callback_fn, &ctx, 0);
```

```c
// ERROR: Invalid memory access
struct iphdr *ip = data + sizeof(struct ethhdr);
__u32 src = ip->saddr;  // Might be out of bounds!

// FIX: Always bounds check
struct iphdr *ip = data + sizeof(struct ethhdr);
if ((void *)(ip + 1) > data_end)
    return XDP_PASS;
__u32 src = ip->saddr;  // Now safe
```

---

## Helper Functions

eBPF programs can't call arbitrary kernel functions. Instead, use **BPF helpers**:

### Networking Helpers

```c
// Map operations
bpf_map_lookup_elem(map, key)      // Get value
bpf_map_update_elem(map, key, val) // Set value
bpf_map_delete_elem(map, key)      // Delete entry

// Packet manipulation (TC only)
bpf_skb_store_bytes(skb, offset, from, len, flags)
bpf_skb_load_bytes(skb, offset, to, len)
bpf_skb_change_head(skb, len, flags)  // Add/remove header space

// XDP helpers
bpf_xdp_adjust_head(ctx, delta)    // Move data pointer
bpf_xdp_adjust_tail(ctx, delta)    // Change packet size
bpf_redirect(ifindex, flags)       // Redirect to interface
bpf_redirect_map(map, key, flags)  // Redirect via DEVMAP

// Checksum
bpf_csum_diff(from, from_size, to, to_size, seed)
bpf_l3_csum_replace(skb, offset, from, to, size)
bpf_l4_csum_replace(skb, offset, from, to, flags)

// Time
bpf_ktime_get_ns()  // Current time in nanoseconds

// Debug
bpf_trace_printk(fmt, fmt_size, ...)  // Print to trace_pipe
```

---

## Practical Exercise: Build a Firewall

### Goal
Create an XDP firewall that:
1. Blocks specific IP addresses (from a map)
2. Blocks specific ports (from a map)
3. Counts blocked/passed packets
4. Exposes metrics to user space

### Step 1: eBPF Program

```c
// bpf/firewall.c
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_endian.h>

#define ETH_P_IP 0x0800
#define IPPROTO_TCP 6
#define IPPROTO_UDP 17

// Blocked IPs
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 10000);
    __type(key, __u32);    // IP address
    __type(value, __u8);   // 1 = blocked
} blocked_ips SEC(".maps");

// Blocked ports
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1000);
    __type(key, __u16);    // Port
    __type(value, __u8);   // 1 = blocked
} blocked_ports SEC(".maps");

// Statistics
struct stats {
    __u64 passed;
    __u64 dropped;
};

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct stats);
} statistics SEC(".maps");

SEC("xdp")
int firewall(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    __u32 stats_key = 0;
    struct stats *stats = bpf_map_lookup_elem(&statistics, &stats_key);
    if (!stats)
        return XDP_PASS;
    
    // Parse Ethernet
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;
    
    // Parse IP
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    // Check blocked IPs
    __u32 src_ip = ip->saddr;
    __u8 *blocked = bpf_map_lookup_elem(&blocked_ips, &src_ip);
    if (blocked && *blocked) {
        stats->dropped++;
        return XDP_DROP;
    }
    
    // Check blocked ports (TCP/UDP)
    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
        if ((void *)(tcp + 1) > data_end)
            return XDP_PASS;
        
        __u16 dst_port = bpf_ntohs(tcp->dest);
        blocked = bpf_map_lookup_elem(&blocked_ports, &dst_port);
        if (blocked && *blocked) {
            stats->dropped++;
            return XDP_DROP;
        }
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)ip + (ip->ihl * 4);
        if ((void *)(udp + 1) > data_end)
            return XDP_PASS;
        
        __u16 dst_port = bpf_ntohs(udp->dest);
        blocked = bpf_map_lookup_elem(&blocked_ports, &dst_port);
        if (blocked && *blocked) {
            stats->dropped++;
            return XDP_DROP;
        }
    }
    
    stats->passed++;
    return XDP_PASS;
}

char LICENSE[] SEC("license") = "GPL";
```

### Step 2: Go User Space Program

```go
package main

import (
    "encoding/binary"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go firewall ./bpf/firewall.c -- -I./bpf/headers

func main() {
    ifaceName := flag.String("iface", "eth0", "Network interface")
    blockIP := flag.String("block-ip", "", "IP to block")
    blockPort := flag.Int("block-port", 0, "Port to block")
    flag.Parse()

    // Load eBPF
    objs := firewallObjects{}
    if err := loadFirewallObjects(&objs, nil); err != nil {
        log.Fatalf("loading objects: %v", err)
    }
    defer objs.Close()

    // Block IP if specified
    if *blockIP != "" {
        ip := net.ParseIP(*blockIP).To4()
        if ip == nil {
            log.Fatalf("invalid IP: %s", *blockIP)
        }
        key := binary.LittleEndian.Uint32(ip)
        val := uint8(1)
        if err := objs.BlockedIps.Put(key, val); err != nil {
            log.Fatalf("blocking IP: %v", err)
        }
        log.Printf("Blocked IP: %s", *blockIP)
    }

    // Block port if specified
    if *blockPort > 0 {
        key := uint16(*blockPort)
        val := uint8(1)
        if err := objs.BlockedPorts.Put(key, val); err != nil {
            log.Fatalf("blocking port: %v", err)
        }
        log.Printf("Blocked port: %d", *blockPort)
    }

    // Attach to interface
    iface, err := net.InterfaceByName(*ifaceName)
    if err != nil {
        log.Fatalf("getting interface: %v", err)
    }

    xdpLink, err := link.AttachXDP(link.XDPOptions{
        Program:   objs.Firewall,
        Interface: iface.Index,
    })
    if err != nil {
        log.Fatalf("attaching XDP: %v", err)
    }
    defer xdpLink.Close()

    log.Printf("Firewall attached to %s, press Ctrl+C to exit", iface.Name)

    // Stats printer
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()

    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

    for {
        select {
        case <-ticker.C:
            printStats(objs.Statistics)
        case <-sig:
            log.Println("Shutting down...")
            return
        }
    }
}

type Stats struct {
    Passed  uint64
    Dropped uint64
}

func printStats(m *ebpf.Map) {
    var (
        key   uint32 = 0
        stats []Stats
    )

    if err := m.Lookup(key, &stats); err != nil {
        log.Printf("lookup stats: %v", err)
        return
    }

    // Aggregate per-CPU stats
    var total Stats
    for _, s := range stats {
        total.Passed += s.Passed
        total.Dropped += s.Dropped
    }

    fmt.Printf("Passed: %d, Dropped: %d\n", total.Passed, total.Dropped)
}
```

---

## Tools for eBPF Development

### bpftool

```bash
# List loaded programs
sudo bpftool prog list

# Show program details
sudo bpftool prog show id 123

# Dump program bytecode
sudo bpftool prog dump xlated id 123

# List maps
sudo bpftool map list

# Dump map contents
sudo bpftool map dump id 45

# Show program attached to interface
sudo bpftool net list
```

### bpftrace

High-level tracing language:

```bash
# Count syscalls by program
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[comm] = count(); }'

# Trace TCP connections
sudo bpftrace -e 'kprobe:tcp_connect { printf("connect: %s\n", comm); }'

# Network latency histogram
sudo bpftrace -e 'kprobe:tcp_rcv_established { @ns = hist(nsecs); }'
```

---

## Key Takeaways

1. **eBPF runs in kernel** - No context switches, line-rate processing
2. **Verifier ensures safety** - Can't crash the kernel
3. **Maps share state** - Between eBPF programs and user space
4. **XDP is fastest** - Use for DDoS, load balancing
5. **TC for more features** - Can modify packets, access sk_buff
6. **cilium/ebpf in Go** - Production-ready library

---

## Next Steps

1. Do the firewall exercise above
2. Add more features: rate limiting, logging
3. Build an L4 load balancer with XDP
4. Study Cilium's eBPF programs
5. Move on to Module 7: XDP Deep Dive

---

## Resources

### 📚 Primary Learning Resource

> **"Learning eBPF" by Liz Rice** - You have this book locally!
> 
> 📖 **Location:** `Learning eBPF New Version.pdf` (in the project root)
> 
> This is the **essential companion** to this module. Work through chapters 1-6 alongside the exercises here.

### Additional Resources

- [Learning eBPF - GitHub Companion](https://github.com/lizrice/learning-ebpf) - Code examples from Liz Rice's book
- [Lima VM](https://github.com/lima-vm/lima) - Linux VMs on macOS (recommended for running eBPF)
- [eBPF.io](https://ebpf.io) - Official eBPF site
- [cilium/ebpf documentation](https://pkg.go.dev/github.com/cilium/ebpf) - Go library docs
- [XDP Tutorial](https://github.com/xdp-project/xdp-tutorial) - Hands-on XDP exercises
- [Brendan Gregg's eBPF page](https://www.brendangregg.com/ebpf.html) - Performance insights
