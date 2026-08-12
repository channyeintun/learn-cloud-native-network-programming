# Cloud-Native Network Programming Roadmap

**Target:** Frontend developers (JS/TS/React/Next.js) with Go knowledge wanting to build cloud networking tools

**Timeline:** 6-8 months (10-15 hours/week)

**Goal:** Contribute to or build cloud networking tools like Cilium, Traefik, CoreDNS, or your own

---

## Why Cloud-Native Networking?

Cloud-native networking is the **hottest area** in infrastructure today:

- **eBPF** is revolutionizing Linux networking, security, and observability
- Companies like Meta, Google, CloudFlare, Netflix use eBPF in production
- **Kubernetes networking** demands specialized skills
- **High demand, low supply** of engineers who understand this stack
- Open source projects (Cilium, Falco, Pixie) are actively seeking contributors

---

## The Big Picture: Software Defined Networking (SDN)

Before diving into implementation, understand this core abstraction that underpins everything in this roadmap:

**SDN separates the brain (control plane) from the muscle (data plane).**

```mermaid
flowchart TB
    subgraph control["Control Plane (The Brain)"]
        direction TB
        K8S["Kubernetes API Server"]
        CTRL["SDN Controller<br>(Cilium Agent, Calico, etc.)"]
        POLICY["Network Policies<br>& Configuration"]
        K8S --> CTRL
        POLICY --> CTRL
    end
    
    subgraph data["Data Plane (The Muscle)"]
        direction TB
        EBPF["eBPF Programs"]
        XDP["XDP"]
        TC["TC (Traffic Control)"]
        IPTABLES["iptables / nftables"]
        KPROXY["kube-proxy"]
    end
    
    CTRL -->|"Programs & Rules"| data
    data -->|"Metrics & Events"| CTRL
    
    style control fill:#e3f2fd,color:#1a1a2e
    style data fill:#c8e6c9,color:#1a1a2e
```

![SDN Architecture Overview](teaching/images/network_concepts_overview_sdn.png)

### Why This Abstraction Matters

| Traditional Networking | Software Defined Networking |
|------------------------|----------------------------|
| Config lives on each device | Centralized, programmable control |
| Manual, error-prone changes | Automated, API-driven |
| Vendor-specific CLIs | Open, standard interfaces |
| Static, slow to adapt | Dynamic, responds to workload changes |

### How Your Learning Map Connects to SDN

Everything you'll learn fits into this paradigm:

| SDN Layer | What You'll Learn | Purpose |
|-----------|-------------------|---------|
| **Control Plane** | Kubernetes, CNI plugins, Cilium Agent — [16](teaching/16-kubernetes-networking.md), [17](teaching/17-cni-plugin-development.md), [19](teaching/19-control-plane-agent-patterns.md) | Decision-making: "where should traffic go?" |
| **Data Plane** | eBPF, XDP, TC, iptables — [06](teaching/06-ebpf-fundamentals.md), [11](teaching/11-ebpf-networking-guide.md), [18](teaching/18-network-policy-enforcement.md), [20](teaching/20-production-load-balancer-datapath.md) | Execution: actually moving packets |
| **Programmability** | Go + cilium/ebpf library — [14](teaching/14-go-development.md), [19](teaching/19-control-plane-agent-patterns.md) | How control plane programs data plane |
| **Observability** | eBPF tracing, metrics — [12](teaching/12-ebpf-security.md), [19 §8](teaching/19-control-plane-agent-patterns.md) | Feedback loop from data plane to control |

> **Key Insight:** When you write an XDP program, you're programming the SDN data plane. When you build a CNI plugin, you're building SDN control plane logic. The implementations fade from memory, but this mental model stays.

### How SDN Came to Be

![SDN Evolution Timeline](teaching/images/sdn_evolution_history.png)

**The Problem SDN Solved:**

In traditional networking, every device (router, switch, firewall) had its own brain bundled with its muscle. This meant:
- Vendor lock-in with proprietary CLIs
- Manual configuration across dozens of devices
- Slow, error-prone changes

**The Three Eras:**

| Era | Years | Key Idea | Limitation |
|-----|-------|----------|------------|
| **Traditional** | 1990s-2007 | Brain + muscle in same box | Vendor lock-in, manual config |
| **OpenFlow SDN** | 2008-2015 | Extract brain to central controller | Required special hardware |
| **eBPF SDN** | 2016+ | Data plane programmable in the Linux kernel (brain stays in user space) | None—this is what you're learning! |

**Key Milestones:**
- **2008:** Stanford creates OpenFlow protocol
- **2011:** Open Networking Foundation formed (Google, Facebook, Microsoft)
- **2012:** Google reveals they run SDN in production (B4 network)
- **2016+:** eBPF matures, Cilium brings SDN to Kubernetes

> **Why eBPF wins:** Original SDN needed special OpenFlow hardware. eBPF runs on any reasonably modern Linux kernel (4.8+ for XDP, 5.10+ recommended — see Kernel Version Requirements below) on commodity NICs—no special hardware required. This is why your roadmap focuses on eBPF.

### Why Containers Demanded eBPF

The evolution of compute (physical → VMs → containers) *drove* the evolution of networking:

![Compute and Networking Evolution](teaching/images/compute_networking_evolution.png)

| Compute Era | Scale | Networking Challenge | Solution |
|-------------|-------|---------------------|----------|
| **Physical** (1990s) | 1 IP per server | Few machines, static | Traditional switches |
| **VMs** (2005+) | 50 VMs per host, migrate | Network must follow VMs | SDN, vSwitch (software) |
| **Containers** (2013+) | 1000s per host, ephemeral | Instant networking, massive scale | eBPF in kernel |

> **The Pattern:** Networking moved *up the stack* as compute virtualized deeper.  
> Hardware switches → Hypervisor vSwitches → Linux kernel (eBPF)

This is why Kubernetes networking uses eBPF (Cilium)—containers are too fast and too many for anything else.

---

## PART I: LEARNING ROADMAP

### Two Tracks, One Journey

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        CORE NETWORKING (Phases 1-4)                     │
│  Foundation skills that work anywhere—Linux, bare metal, VMs, or cloud  │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 1          Phase 2          Phase 3          Phase 4             │
│  Linux Net    →   Go Sockets   →   eBPF         →   XDP                 │
│  (2-3 weeks)      (2-3 weeks)      (4-5 weeks)      (3-4 weeks)         │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                    CLOUD-NATIVE EXTENSION (Phases 5-6)                  │
│         Apply your skills to Kubernetes and production systems          │
├─────────────────────────────────────────────────────────────────────────┤
│  Phase 5                              Phase 6                           │
│  Kubernetes & CNI                 →   Real Projects                     │
│  (4-5 weeks)                          (5-6 weeks)                       │
└─────────────────────────────────────────────────────────────────────────┘
```

| Track | Phases | What You'll Build | Standalone Value? |
|-------|--------|-------------------|-------------------|
| **Core** | 1-4 | Load balancers, firewalls, packet tools | ✅ Yes—valuable without K8s |
| **Cloud-Native** | 5-6 | CNI plugins, observability, Cilium | Extends Core to K8s |

> **Your goal:** Complete both tracks to become a **cloud-native network programmer**.

### Where the Modules Fit

This repo ships 21 teaching modules ([`teaching/`](teaching/), numbered 00-20) and 5 runnable Go
exercises ([`exercises/`](exercises/)). Every phase is now covered by at least one module:

| Phase | Material in this repo |
|-------|----------------------|
| **Phase 1: Linux Networking** | [01 OSI Model](teaching/01-osi-model-deep-dive.md), [02 NAT & Routing](teaching/02-nat-and-routing.md), [03 Connection Tracking](teaching/03-connection-tracking.md), [04 iptables Mastery](teaching/04-iptables-mastery.md), [15 Network Namespaces & Virtual Devices](teaching/15-network-namespaces-and-virtual-devices.md) |
| **Phase 2: Go Network Programming** | [05 Go Networking](teaching/05-go-networking.md) + the 5 [hands-on exercises](exercises/) |
| **Phase 3: eBPF Fundamentals** | [06 eBPF Fundamentals](teaching/06-ebpf-fundamentals.md), [08 eBPF VM Deep Dive](teaching/08-ebpf-vm-deep-dive.md), [09 eBPF Maps Mastery](teaching/09-ebpf-maps-mastery.md), [10 CO-RE & BTF](teaching/10-core-btf-portability.md) |
| **Phase 4: XDP & Packet Processing** | [11 eBPF Networking Guide](teaching/11-ebpf-networking-guide.md), [20 Production Load Balancer Datapath](teaching/20-production-load-balancer-datapath.md) |
| **Phase 5: Kubernetes & CNI** | [16 Kubernetes Networking](teaching/16-kubernetes-networking.md), [17 CNI Plugin Development](teaching/17-cni-plugin-development.md), [18 Network Policy Enforcement](teaching/18-network-policy-enforcement.md) |
| **Phase 6: Building Real Tools** | [12 eBPF Security](teaching/12-ebpf-security.md), [13 Socket Programming](teaching/13-socket-programming.md), [14 Go Development](teaching/14-go-development.md), [19 Control Plane Agent Patterns](teaching/19-control-plane-agent-patterns.md), [20 Production Load Balancer Datapath](teaching/20-production-load-balancer-datapath.md) |

Start from [00 Roadmap Validation](teaching/00-roadmap-validation.md) to assess where you are, and
keep [07 Quick Reference](teaching/07-quick-reference.md) open throughout.

> **Read Module 15 out of order.** It is numbered after the eBPF modules but belongs to Phase 1: it
> builds the namespace/veth/bridge lab that the XDP and TC exercises in Modules 06 and 11 attach to,
> and it is the substrate Phase 5 assumes (a pod *is* a network namespace).

---

### Phase 1: Linux Networking Deep Dive (2-3 weeks)

> **Modules:** [01 OSI Model](teaching/01-osi-model-deep-dive.md) · [02 NAT & Routing](teaching/02-nat-and-routing.md) · [03 Connection Tracking](teaching/03-connection-tracking.md) · [04 iptables Mastery](teaching/04-iptables-mastery.md) · [15 Network Namespaces & Virtual Devices](teaching/15-network-namespaces-and-virtual-devices.md)

**Objectives:**
- Master OSI model layers 2-4 (Ethernet, IP, TCP/UDP)
- Understand Linux kernel networking stack
- Master packet flow through Linux
- Understand namespaces and virtual networking

**Key Concepts:**

```mermaid
flowchart TD
    subgraph ingress["Ingress Path"]
        NIC["NIC"] --> Driver --> XDP["XDP Hook"] --> TC_IN["tc ingress"] --> NF["netfilter/iptables"]
    end
    
    NF --> ROUTE{"Routing Decision"}
    
    ROUTE -->|"Local"| LOCAL["Local Delivery"]
    ROUTE -->|"Forward"| FWD["Forward"]
    
    LOCAL --> SOCKET["Socket Layer"] --> APP["Application"]
    FWD --> TC_OUT["tc egress"] --> NIC_OUT["NIC"]
    
    style XDP fill:#e1f5fe,color:#1a1a2e
    style NF fill:#fff3e0,color:#1a1a2e
    style ROUTE fill:#f3e5f5,color:#1a1a2e
```

**Topics:**
- IP addressing, subnetting, CIDR
- TCP: 3-way handshake, state machine, congestion control
- UDP: connectionless, use cases (DNS, QUIC)
- ARP, MAC addresses, VLANs
- Network namespaces, veth pairs, bridges
- NAT: SNAT, DNAT, masquerading
- Routing: tables, metrics, policy routing

**Exercises:** (1 and 2 are worked end to end, with a reusable `setup-lab.sh`, in
[Module 15](teaching/15-network-namespaces-and-virtual-devices.md))
1. Create network namespaces and connect with veth pairs
2. Build a virtual network with bridges
3. Trace a packet through iptables with logging
4. Use `tcpdump` to capture and analyze traffic
5. Explore `/proc/net/*` and `/sys/class/net/*`

**Commands to Master:**
```bash
# Network namespaces
ip netns add ns1
ip netns exec ns1 ip link
ip link add veth0 type veth peer name veth1
ip link set veth1 netns ns1

# Packet capture
tcpdump -i any -nn 'tcp port 80'
tcpdump -i eth0 -w capture.pcap

# Routing
ip route show table all
ip rule list
cat /proc/net/route

# iptables (foundation before eBPF)
iptables -t filter -L -v -n
iptables -t nat -L -v -n
conntrack -L
```

---

### Phase 2: Go Network Programming (2-3 weeks)

> **Modules:** [05 Go Networking](teaching/05-go-networking.md) · [hands-on Go exercises](exercises/)

**Objectives:**
- Master Go's `net` package
- TCP/UDP socket programming
- HTTP internals and proxies
- Understand syscalls for networking

**Projects:**
1. **Port scanner** - Scan ports with goroutine pool
2. **TCP echo server** - Concurrent connection handling
3. **HTTP proxy** - Forward requests, modify headers
4. **DNS client** - Parse DNS wire format
5. **Packet sniffer** - Using raw sockets

**Key Code Patterns:**

```go
// Raw socket for packet capture
fd, err := syscall.Socket(syscall.AF_PACKET, syscall.SOCK_RAW, int(htons(syscall.ETH_P_ALL)))

// Network namespace switching
runtime.LockOSThread()
defer runtime.UnlockOSThread()
netns.Set(targetNs)

// Efficient connection handling
listener, _ := net.Listen("tcp", ":8080")
for {
    conn, _ := listener.Accept()
    go handle(conn)
}
```

**Key Packages:**
- `net` - TCP/UDP sockets
- `net/http` - HTTP client/server
- `syscall` / `golang.org/x/sys/unix` - Low-level syscalls
- `github.com/google/gopacket` - Packet parsing
- `github.com/vishvananda/netlink` - Netlink API
- `github.com/vishvananda/netns` - Network namespaces

---

### Phase 3: eBPF Fundamentals (4-5 weeks)

> **Modules:** [06 eBPF Fundamentals](teaching/06-ebpf-fundamentals.md) · [08 eBPF VM Deep Dive](teaching/08-ebpf-vm-deep-dive.md) · [09 eBPF Maps Mastery](teaching/09-ebpf-maps-mastery.md) · [10 CO-RE & BTF](teaching/10-core-btf-portability.md)

**This is the core skill that makes cloud-native networking possible.**

**Objectives:**
- Understand eBPF architecture and verifier
- Write eBPF programs in C
- Use cilium/ebpf or libbpf-go from Go
- Master eBPF maps for data sharing
- Attach to various hook points

**What is eBPF?**

```mermaid
flowchart TB
    subgraph userspace["User Space"]
        GO["Your Go Program"] --> LOADER["eBPF Loader<br>(cilium/ebpf)"]
        LOADER <--> MAPS_US["eBPF Maps<br>(shared data)"]
    end
    
    subgraph kernel["Kernel"]
        LOADER --> VM["eBPF Virtual Machine"]
        
        subgraph hooks["Hook Points"]
            XDP["XDP<br>(L2)"]
            TC["TC<br>(L3)"]
            SOCK["Socket<br>Filters"]
            TRACE["Tracepoints<br>& kprobes"]
        end
        
        VM --> hooks
        VM <--> MAPS_K["eBPF Maps"]
        MAPS_K <-.-> MAPS_US
    end
    
    style userspace fill:#e3f2fd,color:#1a1a2e
    style kernel fill:#fff3e0,color:#1a1a2e
    style hooks fill:#c8e6c9,color:#1a1a2e
```

**eBPF Hook Points for Networking:**

| Hook | Layer | Use Case | Performance |
|------|-------|----------|-------------|
| XDP | L2/Driver | Packet filtering, DDoS mitigation | Fastest |
| TC (ingress/egress) | L2/L3 (sk_buff, after skb alloc) | Traffic shaping, load balancing | Fast |
| Socket filters | L4 | Per-socket filtering | Medium |
| cgroup/sock | Socket | Container networking | Medium |
| kprobes | Any | Debugging, tracing | Flexible |

**eBPF Maps (Key-Value Stores):**

eBPF maps are key-value stores shared between kernel and user space — see [Module 09: eBPF Maps Mastery](teaching/09-ebpf-maps-mastery.md) for every map type and its Go-side API.

**Projects:**
1. **Packet counter** - Count packets per IP with XDP
2. **Simple firewall** - Drop packets by IP/port
3. **Connection tracker** - Track TCP connections
4. **Latency tracer** - Measure packet latency with kprobes

**Tools:**
- `bpftool` - Inspect loaded eBPF programs
- `bpftrace` - High-level eBPF tracing
- `libbpf` / `cilium/ebpf` - Libraries for loading eBPF

**Resources:**
- "Learning eBPF" by Liz Rice (O'Reilly)
- https://ebpf.io - Official eBPF site
- https://docs.cilium.io/en/latest/bpf/
- https://github.com/cilium/ebpf (Go library)

---

### Phase 4: XDP & High-Performance Packet Processing (3-4 weeks)

> **Modules:** [11 eBPF Networking Guide](teaching/11-ebpf-networking-guide.md) · [20 Production Load Balancer Datapath](teaching/20-production-load-balancer-datapath.md) (Maglev, DSR/IPIP/GUE, conntrack at scale, AF_XDP)

**Objectives:**
- Master XDP (eXpress Data Path)
- Process millions of packets per second
- Build load balancers and firewalls
- Understand AF_XDP for user-space processing

**XDP Actions:**

```c
// XDP program return values
XDP_DROP    // Drop packet (DDoS defense)
XDP_PASS    // Pass to normal network stack
XDP_TX      // Bounce packet back out same interface
XDP_REDIRECT // Send to another interface or CPU
XDP_ABORTED // Error, drop with trace
```

**XDP Modes:**

| Mode | Description | Performance |
|------|-------------|-------------|
| Native | Driver support required | Excellent (40M+ pps) |
| Generic | Works everywhere (skb-based, no driver support) | Good (1M+ pps) |
| Offloaded | Runs on the NIC itself (in-tree support: Netronome nfp only) | Best, but almost no hardware supports it |

**Project: XDP Load Balancer**

```mermaid
flowchart LR
    CLIENTS["Clients"] --> LB["XDP Load Balancer<br>- Hash by 5-tuple<br>- IPIP/GUE encap<br>- Connection track"]
    LB --> B1["Backend 1"]
    LB --> B2["Backend 2"]
    LB --> B3["Backend 3"]
    
    style LB fill:#e8f5e9,color:#1a1a2e
```

**Implementation approach:**
1. Parse Ethernet, IP, TCP/UDP headers in XDP
2. Hash connection 5-tuple
3. Select backend from BPF map
4. Rewrite packet headers (DSR or NAT)
5. Track connection state in BPF map

**Sample XDP Code:**

```c
/* Requires #include <bpf/bpf_helpers.h>, #include <bpf/bpf_endian.h>
   (bpf_htons lives there - libc's htons() is not available to clang -target bpf)
   and #include <linux/if_ether.h> for ETH_ALEN / ETH_P_IP - those are
   preprocessor macros, and vmlinux.h carries types and enums, not macros. */

struct backend {
    __be32 ip;
    __u8   mac[ETH_ALEN];   /* next-hop MAC, needed for the XDP_TX rewrite */
};

/* XDP has no bpf_l3_csum_replace()/bpf_l4_csum_replace(), so patch the
   checksums incrementally by hand (RFC 1624). */
static __always_inline void csum_replace4(__u16 *sum, __be32 old, __be32 new) {
    __u32 csum = ~((__u32)*sum) & 0xffff;
    csum += (~old >> 16) & 0xffff; csum += ~old & 0xffff;
    csum += (new >> 16) & 0xffff;  csum += new & 0xffff;
    csum = (csum & 0xffff) + (csum >> 16);
    csum = (csum & 0xffff) + (csum >> 16);
    *sum = ~csum & 0xffff;
}

SEC("xdp")
int xdp_lb(struct xdp_md *ctx) {
    void *data = (void *)(long)ctx->data;
    void *data_end = (void *)(long)ctx->data_end;
    
    struct ethhdr *eth = data;
    if ((void *)(eth + 1) > data_end)
        return XDP_PASS;
    
    if (eth->h_proto != bpf_htons(ETH_P_IP))
        return XDP_PASS;
    
    struct iphdr *ip = (void *)(eth + 1);
    if ((void *)(ip + 1) > data_end)
        return XDP_PASS;
    
    // Look up backend in BPF map
    __u32 key = ip->daddr;
    struct backend *backend = bpf_map_lookup_elem(&backends, &key);
    if (!backend)
        return XDP_PASS;
    
    if (ip->ihl != 5)            /* options not handled */
        return XDP_PASS;
    
    // Rewrite destination IP
    __be32 old_daddr = ip->daddr, new_daddr = backend->ip;
    ip->daddr = new_daddr;
    csum_replace4(&ip->check, old_daddr, new_daddr);
    
    /* L4 checksums cover the IPv4 pseudo-header -> must be fixed too */
    if (ip->protocol == IPPROTO_TCP) {
        struct tcphdr *tcp = (void *)ip + sizeof(*ip);
        if ((void *)(tcp + 1) > data_end) return XDP_DROP;
        csum_replace4(&tcp->check, old_daddr, new_daddr);
    } else if (ip->protocol == IPPROTO_UDP) {
        struct udphdr *udp = (void *)ip + sizeof(*ip);
        if ((void *)(udp + 1) > data_end) return XDP_DROP;
        if (udp->check) {
            csum_replace4(&udp->check, old_daddr, new_daddr);
            if (!udp->check) udp->check = 0xFFFF;  /* a computed 0 must be sent as 0xFFFF */
        }
    }
    
    /* XDP_TX reuses the same L2 header - it must be rewritten */
    __builtin_memcpy(eth->h_source, eth->h_dest, ETH_ALEN);   /* LB's own MAC */
    __builtin_memcpy(eth->h_dest, backend->mac, ETH_ALEN);    /* next-hop MAC */
    return XDP_TX;  // Send back out the same interface
}
```

---

### Phase 5: Kubernetes Networking & CNI (4-5 weeks)

> **Modules:** [16 Kubernetes Networking](teaching/16-kubernetes-networking.md) · [17 CNI Plugin Development](teaching/17-cni-plugin-development.md) · [18 Network Policy Enforcement](teaching/18-network-policy-enforcement.md)
> **Prerequisite:** [15 Network Namespaces & Virtual Devices](teaching/15-network-namespaces-and-virtual-devices.md) — a pod is a network namespace, and a CNI plugin is veth plumbing.

**Objectives:**
- Understand Kubernetes networking model
- Build a CNI plugin from scratch
- Understand how Cilium/Calico work
- Master Network Policies

**Kubernetes Networking Model:**

```mermaid
flowchart TB
    subgraph node["Kubernetes Node"]
        subgraph poda["Pod A (10.0.1.5)"]
            CA["Container<br>eth0"]
        end
        
        subgraph podb["Pod B (10.0.1.6)"]
            CB["Container<br>eth0"]
        end
        
        CA -->|veth| CNI
        CB -->|veth| CNI
        
        CNI["CNI Plugin<br>(Cilium, Calico, Flannel)<br>- IP assign<br>- Routing<br>- Policy"]
        CNI --> NET["Node Network<br>(eBPF/iptables)"]
    end
    
    style poda fill:#e3f2fd,color:#1a1a2e
    style podb fill:#e3f2fd,color:#1a1a2e
    style CNI fill:#c8e6c9,color:#1a1a2e
```

**CNI Plugin Development:**

A CNI plugin is just an executable that handles:
- `ADD` - Set up networking for a container
- `DEL` - Clean up networking
- `CHECK` - Verify networking is correct

**Simple CNI Plugin Structure:** (sketch — [Module 17](teaching/17-cni-plugin-development.md) builds the
complete plugin, with IPAM, `DEL`/`CHECK`, MTU handling and a kind cluster to run it in)

```go
// main.go
func cmdAdd(args *skel.CmdArgs) error {
    // 1. Parse config
    conf, _ := parseConfig(args.StdinData)
    
    // 2. Create veth pair
    netns, err := ns.GetNS(args.Netns)   // args.Netns is a path string
    if err != nil {
        return err
    }
    defer netns.Close()
    
    var containerVeth net.Interface
    err = netns.Do(func(hostNS ns.NetNS) error {   // runs inside the container netns
        // the first return value is the host-side veth; drop it here, but a real
        // plugin puts its name into the CNI result it prints
        _, containerVeth, err = ip.SetupVeth(args.IfName, 1500, "", hostNS)
        return err
    })
    if err != nil {
        return err
    }
    
    // 3. Assign IP address
    ipConfig := &current.IPConfig{
        Address: *ipAddr,
        Gateway: gateway,
    }
    
    // 4. Set up routes (inside the container netns)
    err = netns.Do(func(_ ns.NetNS) error {
        link, err := netlink.LinkByName(containerVeth.Name)
        if err != nil {
            return err
        }
        return ip.AddRoute(defaultRoute, gateway, link)
    })
    if err != nil {
        return err
    }
    
    return types.PrintResult(result, conf.CNIVersion)
}
```

**Projects:**
1. **Simple CNI plugin** - Assign IPs, create veth pairs → [Module 17, Parts 3-4](teaching/17-cni-plugin-development.md)
2. **eBPF-based CNI** - Use eBPF for routing instead of iptables → [Module 17, Part 6](teaching/17-cni-plugin-development.md)
3. **Network Policy enforcer** - Implement K8s NetworkPolicy with eBPF → [Module 18](teaching/18-network-policy-enforcement.md)

**Key Concepts:** (all covered in [Module 16](teaching/16-kubernetes-networking.md), except policy → [Module 18](teaching/18-network-policy-enforcement.md))
- Pod networking (every pod gets an IP)
- Service networking (ClusterIP, NodePort, LoadBalancer)
- kube-proxy modes (iptables, IPVS, nftables, eBPF)
- Gateway API (modern) / Ingress (legacy)
- Network Policies

---

### Phase 6: Building Real Tools (5-6 weeks)

> **Modules:** [12 eBPF Security](teaching/12-ebpf-security.md) · [13 Socket Programming](teaching/13-socket-programming.md) · [14 Go Development](teaching/14-go-development.md) · [19 Control Plane Agent Patterns](teaching/19-control-plane-agent-patterns.md) · [20 Production Load Balancer Datapath](teaching/20-production-load-balancer-datapath.md)

Every project below needs the same skeleton: a long-running agent that watches cluster state and
rewrites BPF maps underneath a data plane that must never stop forwarding.
[Module 19](teaching/19-control-plane-agent-patterns.md) is that skeleton — informers and workqueues,
reconciliation, map lifecycle across agent restarts, DaemonSet packaging, capabilities and metrics.

**Choose one or more projects:**

#### Option A: L4 Load Balancer (like Cilium's)

> Fully covered: [Module 20](teaching/20-production-load-balancer-datapath.md) is the datapath for
> every line of this feature list, and [Module 19](teaching/19-control-plane-agent-patterns.md) feeds
> it real `EndpointSlice` state.

```
Features:
- XDP-based packet processing
- Maglev consistent hashing
- DSR (Direct Server Return)
- Health checking
- Connection tracking
- Graceful backend addition/removal
```

#### Option B: Network Observability Tool (like Pixie/Hubble)

```
Features:
- eBPF-based packet capture
- Protocol parsing (HTTP, gRPC, DNS)
- Latency histograms
- Connection flow visualization
- Kubernetes-aware (pod/service names)
```

#### Option C: DNS Proxy/Server (like CoreDNS)

```
Features:
- DNS protocol implementation
- Plugin architecture
- Kubernetes integration
- Caching
- Load balancing
- Metrics export
```

#### Option D: Service Mesh Data Plane (like Envoy, but simpler)

```
Features:  
- L7 proxy with eBPF acceleration
- mTLS termination
- Traffic routing
- Circuit breaking
- Rate limiting
```

---

## PART II: TECHNOLOGY DEEP DIVES

### Essential Go Packages

```bash
# eBPF
go get github.com/cilium/ebpf

# Networking
go get github.com/vishvananda/netlink
go get github.com/vishvananda/netns  
go get github.com/google/gopacket

# Kubernetes
go get k8s.io/client-go
go get github.com/containernetworking/cni      # pkg/skel, pkg/types (plugin protocol)
go get github.com/containernetworking/plugins  # pkg/ip, pkg/ns (veth, routes, netns)

# Observability
go get github.com/prometheus/client_golang

# Utilities
go get github.com/rs/zerolog
go get github.com/spf13/cobra
```

### eBPF Development Setup

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt install -y clang llvm libelf-dev libbpf-dev \
    linux-headers-$(uname -r) bpftool

# Install bpf2go (generates Go from eBPF C)
go install github.com/cilium/ebpf/cmd/bpf2go@latest

# Project structure
myproject/
├── bpf/
│   └── program.c          # eBPF C code
├── main.go                 # Go loader
├── program_bpfel.go        # Generated (little endian)
├── program_bpfeb.go        # Generated (big endian)
└── go.mod

# Generate Go bindings
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -cc clang program ./bpf/program.c
```

### Kernel Version Requirements

| Feature | Minimum Kernel | Recommended |
|---------|---------------|-------------|
| Basic eBPF | 4.4 | 5.10+ |
| XDP | 4.8 | 5.10+ |
| BTF (CO-RE) | 5.2 | 5.10+ |
| Ring buffer | 5.8 | 5.10+ |
| bpf_loop | 5.17 | 5.17+ |

---

## PART III: PROJECTS BREAKDOWN

### Month 1-2: Foundation

**Week 1-2: Linux Networking** ([15](teaching/15-network-namespaces-and-virtual-devices.md), [01](teaching/01-osi-model-deep-dive.md)-[04](teaching/04-iptables-mastery.md))
- Set up lab environment (VMs with multiple interfaces)
- Create network namespaces, veth pairs, bridges
- Master tcpdump, iptables, routing

**Week 3-4: Go Networking**
- Build TCP/UDP servers
- Implement packet sniffer with gopacket
- Build simple HTTP proxy

**Week 5-6: eBPF Hello World**
- Set up eBPF development environment
- Write first XDP program (packet counter)
- Use bpftool to inspect programs

**Week 7-8: eBPF Maps**
- Implement hash maps, arrays
- Share data between kernel and user space
- Build simple firewall

### Month 3-4: Intermediate

**Week 9-10: XDP Deep Dive** ([11](teaching/11-ebpf-networking-guide.md), then [20](teaching/20-production-load-balancer-datapath.md))
- Parse Ethernet, IP, TCP headers
- Implement packet rewriting
- Build L3/L4 load balancer

**Week 11-12: Connection Tracking**
- Implement connection state machine
- Track 5-tuple connections
- Handle NAT

**Week 13-14: Kubernetes Basics** ([16](teaching/16-kubernetes-networking.md))
- Set up local Kubernetes cluster
- Understand pod networking
- Trace packets through kube-proxy

**Week 15-16: CNI Development** ([17](teaching/17-cni-plugin-development.md), then [18](teaching/18-network-policy-enforcement.md))
- Write simple CNI plugin
- Integrate with Kubernetes
- Test with pods
- Enforce NetworkPolicy on identities, not IP addresses

### Month 5-6: Advanced

**Week 17-20: Main Project** ([19](teaching/19-control-plane-agent-patterns.md) for the agent, [20](teaching/20-production-load-balancer-datapath.md) if you picked Option A)
- Choose from Phase 6 projects
- Design architecture
- Implement core features

**Week 21-24: Polish & Production** ([19 §7-8](teaching/19-control-plane-agent-patterns.md), [20 "Measuring It"](teaching/20-production-load-balancer-datapath.md))
- Add metrics and logging
- Write tests
- Documentation
- Performance optimization

---

## PART IV: RESOURCES

### Books
- "Learning eBPF" - Liz Rice (ESSENTIAL)
- "Linux Observability with BPF" - David Calavera
- "BPF Performance Tools" - Brendan Gregg
- "Kubernetes Networking" - James Strong & Vallery Lancey
- "Network Programming with Go" - Adam Woodbeck

### Online Resources
- https://ebpf.io - Official eBPF site
- https://docs.cilium.io - Cilium documentation
- https://github.com/xdp-project/xdp-tutorial - XDP tutorial
- https://nakryiko.com/posts/ - Andrii Nakryiko's blog (eBPF maintainer)
- https://www.brendangregg.com/ebpf.html - Brendan Gregg's eBPF page

### Videos
- [eBPF - Everything You Need to Know in 5 Minutes](https://www.youtube.com/watch?v=KhPrMW5Rbbc) - Thomas Graf (CTO, Isovalent) KubeCon keynote - Quick intro to eBPF

### Projects to Study
- **Cilium** - eBPF-based Kubernetes networking
- **Katran** - Facebook's L4 load balancer
- **Calico** - Kubernetes networking with eBPF mode
- **Falco** - eBPF-based security
- **Pixie** - eBPF-based observability
- **CoreDNS** - DNS server in Go
- **Traefik** - Cloud-native proxy

### Communities
- eBPF Slack: https://ebpf.io/slack
- Cilium Slack: https://cilium.slack.com
- CNCF Slack: https://slack.cncf.io
- r/kubernetes, r/golang

---

## PART V: CAREER PATH

After completing this roadmap, you'll be qualified for:

- **Cloud Network Engineer** at hyperscalers (Google, Meta, AWS)
- **Kubernetes Networking Specialist**
- **eBPF Engineer** (very high demand)
- **Open Source Contributor** (Cilium, Calico, Falco)
- **SRE/Platform Engineer** with networking focus
- **Security Engineer** (eBPF-based security tools)

**Salary Range:** $150K-$300K+ for senior eBPF engineers

### Further Learning
- DPDK for user-space networking
- P4 programming
- SmartNIC programming
- Kernel development
- Network security (DDoS mitigation)

---

## TIMELINE SUMMARY

**Total: 6-8 months** (10-15 hours/week)

| Month | Focus | Deliverable |
|-------|-------|-------------|
| 1 | Linux networking + Go | Packet sniffer, HTTP proxy |
| 2 | eBPF fundamentals | XDP packet counter, firewall |
| 3 | XDP deep dive | L4 load balancer |
| 4 | Kubernetes networking | Simple CNI plugin |
| 5 | Main project | Core implementation |
| 6 | Production ready | Complete project |

---

## YOUR ADVANTAGES

As a frontend developer with Go knowledge:

1. **Full-stack observability tools** - Build dashboards for your eBPF tools
2. **Developer experience** - Make CLI and APIs that developers love
3. **Documentation** - Clear docs are rare in this space
4. **Visualization** - Network topology, flow diagrams

This combination of **low-level kernel programming + user-friendly interfaces** is extremely rare and valuable!

---

## NEXT STEPS

1. **This week:** Set up Linux VM, create network namespaces — follow [15 Network Namespaces & Virtual Devices](teaching/15-network-namespaces-and-virtual-devices.md) and keep the lab it builds
2. **Install:** bpftool, clang, llvm, Go 1.24+
3. **Read:** First 2 chapters of "Learning eBPF"
4. **Code:** Write your first XDP program
5. **Join:** eBPF Slack community

See the [module index in README.md](README.md) for the full list of teaching modules and exercises.

**Good luck on your journey to becoming a cloud-native network engineer!** 🚀