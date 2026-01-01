# Quick Reference - Cloud-Native Network Programming

## 📊 Complete Concept Overview

![Network Concepts Overview](./images/network_concepts_overview.png)

## 📚 Primary eBPF Resource

> **"Learning eBPF" by Liz Rice** → `Learning eBPF New Version.pdf` (project root)

---

## Linux Networking Commands

### Network Namespaces
```bash
# Create namespace
ip netns add ns1

# Execute in namespace
ip netns exec ns1 ip link

# Create veth pair
ip link add veth0 type veth peer name veth1

# Move interface to namespace
ip link set veth1 netns ns1

# Bring up interface
ip netns exec ns1 ip link set veth1 up
ip netns exec ns1 ip addr add 10.0.0.1/24 dev veth1

# Connect namespaces via bridge
ip link add br0 type bridge
ip link set br0 up
ip link set veth0 master br0
```

### Packet Capture
```bash
# Basic capture
tcpdump -i any -nn 'tcp port 80'

# Save to file
tcpdump -i eth0 -w capture.pcap

# Read from file
tcpdump -r capture.pcap

# Filter expressions
tcpdump 'host 192.168.1.1'
tcpdump 'tcp and port 443'
tcpdump 'icmp'

# Verbose output
tcpdump -i eth0 -vvv

# Show packet contents
tcpdump -i eth0 -X
```

### Routing
```bash
# Show routes
ip route show
ip route show table all
ip rule list

# Add route
ip route add 10.0.0.0/8 via 192.168.1.1

# Add policy route
ip rule add from 10.0.0.0/8 table 100
ip route add default via 192.168.1.1 table 100

# Delete route
ip route del 10.0.0.0/8
```

### iptables (Foundation)
```bash
# List rules
iptables -L -v -n
iptables -t nat -L -v -n
iptables -t mangle -L -v -n

# Add rules
iptables -A INPUT -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -s 10.0.0.0/8 -j DROP

# NAT
iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
iptables -t nat -A PREROUTING -p tcp --dport 80 -j DNAT --to 10.0.0.5:8080

# Flush rules
iptables -F
```

### Connection Tracking
```bash
# List connections
conntrack -L
conntrack -L -p tcp
conntrack -L -s 192.168.1.100

# Count connections
conntrack -C

# Clear connections
conntrack -F

# Watch new connections
conntrack -E
```

---

## eBPF/XDP Commands

### bpftool
```bash
# List programs
sudo bpftool prog list
sudo bpftool prog list --json | jq

# Show program details
sudo bpftool prog show id 123

# Dump bytecode
sudo bpftool prog dump xlated id 123
sudo bpftool prog dump jited id 123

# List maps
sudo bpftool map list

# Show map contents
sudo bpftool map dump id 45

# Lookup map entry
sudo bpftool map lookup id 45 key 0x01 0x00 0x00 0x00

# Update map entry
sudo bpftool map update id 45 key 0x01 0x00 0x00 0x00 value 0x64 0x00 0x00 0x00

# List network attachments
sudo bpftool net list

# Show BTF info
sudo bpftool btf list
```

### bpftrace
```bash
# List tracepoints
sudo bpftrace -l 'tracepoint:*'

# Count syscalls
sudo bpftrace -e 'tracepoint:syscalls:sys_enter_* { @[comm] = count(); }'

# Trace TCP connects
sudo bpftrace -e 'kprobe:tcp_connect { printf("%s -> connect\n", comm); }'

# Network latency
sudo bpftrace -e 'kprobe:tcp_rcv_established { @ns = hist(nsecs); }'

# Trace packets
sudo bpftrace -e 'kprobe:netif_receive_skb { @packets = count(); }'
```

### XDP Attach/Detach
```bash
# Attach XDP (using ip)
ip link set dev eth0 xdpgeneric obj program.o sec xdp

# Detach XDP
ip link set dev eth0 xdpgeneric off

# Check XDP status
ip link show eth0

# Using bpftool
bpftool net attach xdpgeneric id 123 dev eth0
bpftool net detach xdpgeneric dev eth0
```

---

## Go Code Snippets

### cilium/ebpf Setup
```bash
# Install bpf2go
go install github.com/cilium/ebpf/cmd/bpf2go@latest

# Generate Go bindings
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang program ./bpf/program.c -- -I./bpf/headers
go generate ./...
```

### Load and Attach XDP
```go
import (
    "github.com/cilium/ebpf"
    "github.com/cilium/ebpf/link"
)

// Load
objs := programObjects{}
loadProgramObjects(&objs, nil)
defer objs.Close()

// Attach
iface, _ := net.InterfaceByName("eth0")
xdpLink, _ := link.AttachXDP(link.XDPOptions{
    Program:   objs.XdpProg,
    Interface: iface.Index,
})
defer xdpLink.Close()
```

### Map Operations
```go
// Put
objs.MyMap.Put(key, value)

// Get
var value uint64
objs.MyMap.Lookup(key, &value)

// Delete
objs.MyMap.Delete(key)

// Iterate
iter := objs.MyMap.Iterate()
for iter.Next(&key, &value) {
    fmt.Printf("%v: %v\n", key, value)
}
```

### Network Namespaces
```go
import (
    "github.com/vishvananda/netns"
    "runtime"
)

// Switch namespace
runtime.LockOSThread()
defer runtime.UnlockOSThread()

ns, _ := netns.GetFromName("myns")
defer ns.Close()

netns.Set(ns)
// Do network operations
netns.Set(origNs)
```

### Netlink Routes
```go
import "github.com/vishvananda/netlink"

// List routes
routes, _ := netlink.RouteList(nil, netlink.FAMILY_V4)

// Add route
route := &netlink.Route{
    Dst: &net.IPNet{
        IP:   net.ParseIP("10.0.0.0"),
        Mask: net.CIDRMask(24, 32),
    },
    Gw: net.ParseIP("192.168.1.1"),
}
netlink.RouteAdd(route)
```

---

## Common eBPF Patterns

### Bounds Checking (Required!)
```c
// Always bounds check before access
struct ethhdr *eth = data;
if ((void *)(eth + 1) > data_end)
    return XDP_PASS;

struct iphdr *ip = (void *)(eth + 1);
if ((void *)(ip + 1) > data_end)
    return XDP_PASS;
```

### Map Lookup Pattern
```c
__u32 key = ip->saddr;
__u64 *value = bpf_map_lookup_elem(&my_map, &key);
if (value) {
    __sync_fetch_and_add(value, 1);
} else {
    __u64 initial = 1;
    bpf_map_update_elem(&my_map, &key, &initial, BPF_ANY);
}
```

### Packet Header Parsing
```c
// Ethernet
struct ethhdr *eth = data;
if (eth->h_proto != bpf_htons(ETH_P_IP))
    return XDP_PASS;

// IP
struct iphdr *ip = (void *)(eth + 1);
__u8 protocol = ip->protocol;
__u32 src_ip = ip->saddr;

// TCP (variable IP header length!)
struct tcphdr *tcp = (void *)ip + (ip->ihl * 4);
if ((void *)(tcp + 1) > data_end)
    return XDP_PASS;
__u16 dst_port = bpf_ntohs(tcp->dest);
```

---

## Kernel Requirements

| Feature | Minimum | Recommended |
|---------|---------|-------------|
| Basic eBPF | 4.4 | 5.10+ |
| XDP | 4.8 | 5.10+ |
| BTF (CO-RE) | 5.2 | 5.10+ |
| Ring buffer | 5.8 | 5.10+ |
| bpf_loop | 5.17 | 5.17+ |

Check your kernel:
```bash
uname -r
cat /boot/config-$(uname -r) | grep BPF
```

---

## Debug Tips

### Verifier Errors
```bash
# Get detailed verifier output
sudo cat /sys/kernel/debug/tracing/trace_pipe
```

### eBPF Logs
```c
// In eBPF program
bpf_printk("debug: ip=%u port=%u\n", ip, port);

// Read logs
sudo cat /sys/kernel/debug/tracing/trace_pipe
```

### Check XDP Stats
```bash
# Interface stats
ip -s link show eth0

# XDP stats
ethtool -S eth0 | grep xdp
```

---

## File Locations

| What | Location |
|------|----------|
| BTF vmlinux | `/sys/kernel/btf/vmlinux` |
| BPF filesystem | `/sys/fs/bpf/` |
| Trace pipe | `/sys/kernel/debug/tracing/trace_pipe` |
| Network stats | `/proc/net/*` |
| Interface info | `/sys/class/net/*` |
