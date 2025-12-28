# Module 4: iptables Mastery for Load Balancing

## 📊 Visual Learning

![iptables Chains and Tables](./images/iptables_chains.png)

## iptables Architecture Overview

iptables is organized into **tables** containing **chains** of **rules**.

```
┌─────────────────────────────────────────────────────────────────┐
│                        iptables TABLES                          │
├───────────────┬──────────────────────────────────────────────────┤
│    filter     │ Default table. Accept, drop, reject packets     │
├───────────────┼──────────────────────────────────────────────────┤
│     nat       │ Network Address Translation                      │
├───────────────┼──────────────────────────────────────────────────┤
│    mangle     │ Packet modification (marks, TTL, TOS) ⭐        │
├───────────────┼──────────────────────────────────────────────────┤
│     raw       │ Bypass connection tracking                       │
├───────────────┼──────────────────────────────────────────────────┤
│   security    │ SELinux security marking                         │
└───────────────┴──────────────────────────────────────────────────┘

⭐ mangle is WHERE YOUR LOAD BALANCER LOGIC LIVES
```

---

## Table & Chain Reference

### Which Table for What?

| Table | Purpose | Your Use |
|-------|---------|----------|
| **mangle** | Packet marking | Mark packets for routing |
| **nat** | Address translation | Masquerading |
| **filter** | Firewalling | (Not primary focus) |

### Chain Flow for Forwarded Packets

```mermaid
flowchart TD
    IN["Incoming packet"] --> RAW["raw PREROUTING<br>(skip conntrack if needed)"]
    RAW --> MANGLE_PRE["mangle PREROUTING<br>⭐ MARK packets here"]
    MANGLE_PRE --> NAT_PRE["nat PREROUTING<br>(DNAT happens here)"]
    NAT_PRE --> ROUTE{"Routing Decision<br>ip rules check MARK here"}
    
    ROUTE --> MANGLE_FWD["mangle FORWARD"]
    MANGLE_FWD --> FILTER_FWD["filter FORWARD<br>(firewall decision)"]
    FILTER_FWD --> MANGLE_POST["mangle POSTROUTING"]
    MANGLE_POST --> NAT_POST["nat POSTROUTING<br>⭐ MASQUERADE here"]
    NAT_POST --> OUT["Outgoing packet"]
    
    style MANGLE_PRE fill:#e8f5e9
    style NAT_POST fill:#e8f5e9
    style ROUTE fill:#f3e5f5
```

---

## Essential iptables Commands

### Viewing Rules

```bash
# List all rules in all tables
iptables -L -v -n

# Specific table
iptables -t nat -L -v -n
iptables -t mangle -L -v -n

# With line numbers (for deletion)
iptables -t mangle -L PREROUTING -v -n --line-numbers

# Show rules as commands (for backup)
iptables-save
```

### Managing Rules

```bash
# Append rule to chain
iptables -t mangle -A PREROUTING ...

# Insert at position
iptables -t mangle -I PREROUTING 1 ...

# Delete by specification
iptables -t mangle -D PREROUTING ...

# Delete by line number
iptables -t mangle -D PREROUTING 3

# Flush (delete all rules in chain)
iptables -t mangle -F PREROUTING

# Flush entire table
iptables -t mangle -F
```

---

## Multi-WAN iptables Configuration

### Complete Setup Script

```bash
#!/bin/bash

# ============================================
# MULTI-WAN LOAD BALANCER - iptables Setup
# ============================================

# Configuration
ISP1_IF="eth0"        # ISP1 interface
ISP2_IF="eth1"        # ISP2 interface
LAN_IF="eth2"         # LAN interface
ISP1_MARK=1
ISP2_MARK=2
ISP1_WEIGHT=70        # 70% to ISP1
ISP2_WEIGHT=30        # 30% to ISP2

# Clear existing rules
iptables -t mangle -F
iptables -t nat -F

# ============================================
# STEP 1: Restore marks for existing connections
# ============================================
# This MUST come first - existing connections keep their ISP
iptables -t mangle -A PREROUTING \
    -m conntrack --ctstate ESTABLISHED,RELATED \
    -j CONNMARK --restore-mark

# ============================================
# STEP 2: Mark NEW connections (load balancing)
# ============================================
# Calculate probability (weight/100)
# For 70/30 split: mark 1 with p=0.70, rest get mark 2

# Method: Statistical random marking
iptables -t mangle -A PREROUTING \
    -i $LAN_IF \
    -m conntrack --ctstate NEW \
    -m statistic --mode random --probability 0.70 \
    -j MARK --set-mark $ISP1_MARK

# Everything not marked yet gets ISP2
iptables -t mangle -A PREROUTING \
    -i $LAN_IF \
    -m conntrack --ctstate NEW \
    -m mark --mark 0 \
    -j MARK --set-mark $ISP2_MARK

# ============================================
# STEP 3: Save marks to connection tracking
# ============================================
iptables -t mangle -A PREROUTING -j CONNMARK --save-mark

# ============================================
# STEP 4: NAT (Masquerading)
# ============================================
iptables -t nat -A POSTROUTING -o $ISP1_IF -j MASQUERADE
iptables -t nat -A POSTROUTING -o $ISP2_IF -j MASQUERADE

# ============================================
# STEP 5: Policy routing rules
# ============================================
ip rule add fwmark $ISP1_MARK table isp1 priority 100 2>/dev/null
ip rule add fwmark $ISP2_MARK table isp2 priority 100 2>/dev/null

echo "Multi-WAN iptables configured!"
echo "ISP1 (mark=$ISP1_MARK): $ISP1_WEIGHT%"
echo "ISP2 (mark=$ISP2_MARK): $ISP2_WEIGHT%"
```

---

## Advanced Marking Techniques

### 1. Mark by Source IP

```bash
# Gaming PC always uses ISP1 (low latency)
iptables -t mangle -A PREROUTING \
    -s 10.0.0.50 \
    -m conntrack --ctstate NEW \
    -j MARK --set-mark 1

# Work laptop always uses ISP2 (prioritize bandwidth)
iptables -t mangle -A PREROUTING \
    -s 10.0.0.100 \
    -m conntrack --ctstate NEW \
    -j MARK --set-mark 2
```

### 2. Mark by Destination Port

```bash
# Gaming ports → ISP1
iptables -t mangle -A PREROUTING \
    -p udp --dport 27015:27030 \
    -m conntrack --ctstate NEW \
    -j MARK --set-mark 1

# Video streaming (typically HTTPS) → ISP2
iptables -t mangle -A PREROUTING \
    -p tcp --dport 443 \
    -m conntrack --ctstate NEW \
    -j MARK --set-mark 2
```

### 3. Mark by IP Set (Multiple Destinations)

```bash
# Create IP set for Netflix IPs
ipset create netflix hash:net
ipset add netflix 23.246.0.0/18
ipset add netflix 37.77.184.0/21
# ... more Netflix IP ranges

# Route Netflix to ISP2
iptables -t mangle -A PREROUTING \
    -m set --match-set netflix dst \
    -m conntrack --ctstate NEW \
    -j MARK --set-mark 2
```

---

## Managing Rules from Go

### Using go-iptables Library

```go
package main

import (
    "fmt"
    "github.com/coreos/go-iptables/iptables"
)

func main() {
    ipt, err := iptables.New()
    if err != nil {
        panic(err)
    }

    // List current mangle PREROUTING rules
    rules, err := ipt.List("mangle", "PREROUTING")
    if err != nil {
        panic(err)
    }

    fmt.Println("Current mangle PREROUTING rules:")
    for i, rule := range rules {
        fmt.Printf("  %d: %s\n", i, rule)
    }
}
```

### Adding/Removing Rules Programmatically

```go
func (lb *LoadBalancer) addSourceRouting(srcIP string, mark int) error {
    ipt, _ := iptables.New()
    
    return ipt.Insert("mangle", "PREROUTING", 1,
        "-s", srcIP,
        "-m", "conntrack", "--ctstate", "NEW",
        "-j", "MARK", "--set-mark", fmt.Sprintf("%d", mark),
    )
}

func (lb *LoadBalancer) removeSourceRouting(srcIP string, mark int) error {
    ipt, _ := iptables.New()
    
    return ipt.Delete("mangle", "PREROUTING",
        "-s", srcIP,
        "-m", "conntrack", "--ctstate", "NEW",
        "-j", "MARK", "--set-mark", fmt.Sprintf("%d", mark),
    )
}

func (lb *LoadBalancer) updateWeightedDistribution(isp1Prob float64) error {
    ipt, _ := iptables.New()
    
    // Clear existing distribution rules
    ipt.ClearChain("mangle", "LOAD_BALANCE")
    
    // Re-add with new probability
    return ipt.Append("mangle", "LOAD_BALANCE",
        "-m", "statistic", "--mode", "random",
        "--probability", fmt.Sprintf("%.2f", isp1Prob),
        "-j", "MARK", "--set-mark", "1",
    )
}
```

---

## Debugging iptables

### Check Packet Counts

```bash
# View packet/byte counters
iptables -t mangle -L PREROUTING -v -n

# Sample output:
# Chain PREROUTING (policy ACCEPT 0 packets, 0 bytes)
# pkts bytes target     prot opt in     out     source        destination
# 1234 456K CONNMARK   all  --  *      *       0.0.0.0/0     0.0.0.0/0     ctstate ESTABLISHED,RELATED CONNMARK restore
#  567 89K  MARK       all  --  eth2   *       0.0.0.0/0     0.0.0.0/0     ctstate NEW statistic mode random probability 0.70 MARK set 0x1
```

### Log Matched Packets

```bash
# Add logging rule before actual rule
iptables -t mangle -I PREROUTING 1 \
    -m conntrack --ctstate NEW \
    -j LOG --log-prefix "NEW-CONN: " --log-level 4

# Watch logs
tail -f /var/log/kern.log | grep "NEW-CONN"
```

### Trace Packet Path

```bash
# Enable tracing for specific traffic
iptables -t raw -A PREROUTING -p icmp -j TRACE
iptables -t raw -A OUTPUT -p icmp -j TRACE

# View trace
dmesg | grep TRACE

# Remember to remove when done!
iptables -t raw -F
```

---

## Practical Exercises

### Exercise 1: Build and Test Basic Marking

```bash
# 1. Add test marking rule
sudo iptables -t mangle -A PREROUTING \
    -p icmp \
    -j MARK --set-mark 99

# 2. Test
ping -c 1 8.8.8.8

# 3. Check conntrack
conntrack -L -p icmp

# 4. Clean up
sudo iptables -t mangle -D PREROUTING -p icmp -j MARK --set-mark 99
```

### Exercise 2: Monitor Distribution

```bash
# Watch connection distribution in real-time
watch -n 1 'echo "ISP1: $(conntrack -L 2>/dev/null | grep "mark=1" | wc -l) | ISP2: $(conntrack -L 2>/dev/null | grep "mark=2" | wc -l)"'
```

---

## Key Takeaways

1. **mangle table** is where marking happens
2. **PREROUTING** chain catches packets before routing decision
3. **CONNMARK** saves marks for connection persistence
4. Order matters: restore first, then mark NEW, then save
5. **go-iptables** provides programmatic control
6. Verify with counters and conntrack

---

## Next Module
→ [05-go-networking.md](./05-go-networking.md): Go network programming for your load balancer
