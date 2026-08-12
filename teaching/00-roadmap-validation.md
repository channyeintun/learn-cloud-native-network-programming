# Roadmap Validation Report

## ✅ Cloud-Native Network Programming Roadmap

Your new roadmap focuses on building cloud networking tools using **eBPF, XDP, and Kubernetes networking**.

## 📊 Complete Concept Overview

![Network Concepts Overview](./images/network_concepts_overview.png)

*This visual shows how the core networking concepts connect across teaching modules 01-07. Modules 08-14 extend it with advanced eBPF topics, and modules 15-20 take it into Kubernetes and production systems.*

---

## Current Knowledge Assessment

Based on what you shared previously. **"Your Starting Level" describes where *you* began — not
gaps in this course.** Every topic below now has a module that closes it:

| Topic | Your Starting Level | Roadmap Requires | Covered By |
|-------|--------------------|------------------|------------|
| Number Systems | ✅ Strong | Basic | — (assumed) |
| IP Addressing | ✅ Strong | Intermediate | [01 OSI Model](./01-osi-model-deep-dive.md) |
| Subnetting/CIDR | ✅ Strong | Advanced | [01](./01-osi-model-deep-dive.md), [02 NAT & Routing](./02-nat-and-routing.md) |
| OSI Model | ⚠️ Mentioned | Deep L2-L4 | [01 OSI Model](./01-osi-model-deep-dive.md) |
| TCP/UDP | ⚠️ Basic | State machines | [01](./01-osi-model-deep-dive.md), [03 Connection Tracking](./03-connection-tracking.md) |
| NAT | ⚠️ Conceptual | SNAT/DNAT | [02 NAT & Routing](./02-nat-and-routing.md), [04 iptables Mastery](./04-iptables-mastery.md) |
| Linux Namespaces | ❌ New to you | Essential | ✅ [15 Network Namespaces & Virtual Devices](./15-network-namespaces-and-virtual-devices.md) |
| eBPF | ❌ New to you | Core skill | ✅ [06](./06-ebpf-fundamentals.md), [08](./08-ebpf-vm-deep-dive.md)-[10](./10-core-btf-portability.md), [12](./12-ebpf-security.md), [13](./13-socket-programming.md) |
| XDP | ❌ New to you | Essential | ✅ [11 eBPF Networking Guide](./11-ebpf-networking-guide.md), [20 Production LB Datapath](./20-production-load-balancer-datapath.md) |
| Kubernetes | ❓ Unknown | Deep understanding | ✅ [16 K8s Networking](./16-kubernetes-networking.md), [17 CNI Plugins](./17-cni-plugin-development.md), [18 Network Policy](./18-network-policy-enforcement.md), [19 Control Plane Agents](./19-control-plane-agent-patterns.md) |
| Go networking | ❓ Unknown | Core skill | ✅ [05 Go Networking](./05-go-networking.md) + [exercises](../exercises/), [14 Go Development](./14-go-development.md) |

---

## What You'll Build

### Milestone Projects

1. **Month 1-2: Foundation**
   - Packet sniffer with gopacket
   - HTTP proxy
   - XDP packet counter

2. **Month 3-4: eBPF Mastery**
   - XDP firewall with configurable rules
   - L4 load balancer
   - Connection tracker

3. **Month 5-6: Kubernetes Integration** — [16](./16-kubernetes-networking.md), [17](./17-cni-plugin-development.md), [18](./18-network-policy-enforcement.md)
   - Simple CNI plugin
   - eBPF-based network policy enforcer

4. **Month 6-8: Major Project** — [19 Control Plane Agent Patterns](./19-control-plane-agent-patterns.md) is the agent skeleton every option needs
   Choose one:
   - L4 Load Balancer (like Katran) — [20 Production Load Balancer Datapath](./20-production-load-balancer-datapath.md)
   - Network Observability Tool (like Hubble) — [12](./12-ebpf-security.md), [18 drop observability](./18-network-policy-enforcement.md)
   - DNS Proxy (like CoreDNS) — [13 Socket Programming](./13-socket-programming.md), [16 CoreDNS section](./16-kubernetes-networking.md)
   - Service Mesh Data Plane — [13 SOCKMAP/sk_msg](./13-socket-programming.md), [18 L7 proxy redirect](./18-network-policy-enforcement.md)

---

## Key Technology Stack

```
┌─────────────────────────────────────────────┐
│            Your Skill Stack                 │
├─────────────────────────────────────────────┤
│  Language     │  Go                         │
│  Core Tech    │  eBPF, XDP                  │
│  Libraries    │  cilium/ebpf, netlink       │
│  Platform     │  Linux Kernel, Kubernetes   │
│  Tools        │  bpftool, bpftrace, tcpdump │
└─────────────────────────────────────────────┘
```

---

## Learning Path

### Phase 1: Linux Networking (2-3 weeks)
> Modules [01](./01-osi-model-deep-dive.md)-[04](./04-iptables-mastery.md) and [15](./15-network-namespaces-and-virtual-devices.md)
- Network namespaces, veth pairs, bridges
- Packet flow through kernel
- tcpdump mastery

### Phase 2: Go Network Programming (2-3 weeks)
> Module [05](./05-go-networking.md) + the [exercises](../exercises/)
- TCP/UDP sockets
- gopacket for packet parsing
- netlink for kernel communication

### Phase 3: eBPF Fundamentals (4-5 weeks) ⭐ MOST IMPORTANT
> Modules [06](./06-ebpf-fundamentals.md), [08](./08-ebpf-vm-deep-dive.md), [09](./09-ebpf-maps-mastery.md), [10](./10-core-btf-portability.md)
- eBPF architecture & verifier
- XDP programs
- Maps for state management
- cilium/ebpf library

### Phase 4: XDP Deep Dive (3-4 weeks)
> Modules [11](./11-ebpf-networking-guide.md), [20](./20-production-load-balancer-datapath.md)
- Header parsing & rewriting
- Load balancing algorithms
- Connection tracking

### Phase 5: Kubernetes Networking (4-5 weeks)
> Modules [16](./16-kubernetes-networking.md), [17](./17-cni-plugin-development.md), [18](./18-network-policy-enforcement.md)
- CNI plugin development
- Service networking
- Network policies with eBPF

### Phase 6: Production Tool (5-6 weeks)
> Modules [12](./12-ebpf-security.md), [13](./13-socket-programming.md), [14](./14-go-development.md), [19](./19-control-plane-agent-patterns.md), [20](./20-production-load-balancer-datapath.md)
- Complete one major project
- Metrics, logging, testing
- Documentation

> **Module 15 is numbered late but belongs to Phase 1.** Its namespace/veth/bridge lab is what you
> attach the Phase 3 and Phase 4 XDP and TC programs to — read it before Module 6.

---

## Teaching Modules

21 modules (00-20), all written and ready, plus 5 runnable Go [exercises](../exercises/).

| Module | File | Track | Status |
|--------|------|-------|--------|
| Roadmap Validation | [00-roadmap-validation.md](./00-roadmap-validation.md) | Orientation | ✅ Ready |
| OSI Model Deep Dive | [01-osi-model-deep-dive.md](./01-osi-model-deep-dive.md) | Phase 1 | ✅ Ready |
| NAT and Routing | [02-nat-and-routing.md](./02-nat-and-routing.md) | Phase 1 | ✅ Ready |
| Connection Tracking | [03-connection-tracking.md](./03-connection-tracking.md) | Phase 1 | ✅ Ready |
| iptables Mastery | [04-iptables-mastery.md](./04-iptables-mastery.md) | Phase 1 | ✅ Ready |
| Go Networking | [05-go-networking.md](./05-go-networking.md) | Phase 2 | ✅ Ready |
| eBPF Fundamentals | [06-ebpf-fundamentals.md](./06-ebpf-fundamentals.md) | Phase 3 | ✅ Ready |
| Quick Reference | [07-quick-reference.md](./07-quick-reference.md) | Reference | ✅ Ready |
| eBPF VM Deep Dive | [08-ebpf-vm-deep-dive.md](./08-ebpf-vm-deep-dive.md) | Phase 3 | ✅ Ready |
| eBPF Maps Mastery | [09-ebpf-maps-mastery.md](./09-ebpf-maps-mastery.md) | Phase 3 | ✅ Ready |
| CO-RE & BTF Portability | [10-core-btf-portability.md](./10-core-btf-portability.md) | Phase 3 | ✅ Ready |
| eBPF Networking Guide | [11-ebpf-networking-guide.md](./11-ebpf-networking-guide.md) | Phase 4 | ✅ Ready |
| eBPF Security | [12-ebpf-security.md](./12-ebpf-security.md) | Phase 6 | ✅ Ready |
| Socket Programming | [13-socket-programming.md](./13-socket-programming.md) | Phase 6 | ✅ Ready |
| Go Development | [14-go-development.md](./14-go-development.md) | Phase 6 | ✅ Ready |
| Network Namespaces & Virtual Devices | [15-network-namespaces-and-virtual-devices.md](./15-network-namespaces-and-virtual-devices.md) | Phase 1 (read before 06) | ✅ Ready |
| Kubernetes Networking | [16-kubernetes-networking.md](./16-kubernetes-networking.md) | Phase 5 | ✅ Ready |
| CNI Plugin Development | [17-cni-plugin-development.md](./17-cni-plugin-development.md) | Phase 5 | ✅ Ready |
| Network Policy Enforcement | [18-network-policy-enforcement.md](./18-network-policy-enforcement.md) | Phase 5 | ✅ Ready |
| Control Plane Agent Patterns | [19-control-plane-agent-patterns.md](./19-control-plane-agent-patterns.md) | Phase 6 | ✅ Ready |
| Production Load Balancer Datapath | [20-production-load-balancer-datapath.md](./20-production-load-balancer-datapath.md) | Phase 4 + 6 | ✅ Ready |

---

## Why This Path?

### Career Value

eBPF engineers are in **extremely high demand**:

- Meta, Google, CloudFlare, Netflix use eBPF
- Kubernetes networking is moving to eBPF (Cilium)
- Security tools use eBPF (Falco, Tetragon)
- Observability uses eBPF (Pixie, Grafana Beyla)

**Salary range**: $150K-$300K+ for senior eBPF engineers

### Your Advantage

As a frontend developer who masters eBPF:
- Build observability dashboards others can't
- Create developer-friendly CLIs and APIs
- Full-stack from kernel to UI is rare

---

## Next Steps

1. Foundations: modules [01](./01-osi-model-deep-dive.md)-[05](./05-go-networking.md)
2. Build the lab: [15 Network Namespaces & Virtual Devices](./15-network-namespaces-and-virtual-devices.md) — out of numeric order, before any eBPF
3. eBPF core: [06](./06-ebpf-fundamentals.md), then [08](./08-ebpf-vm-deep-dive.md)-[10](./10-core-btf-portability.md)
4. Applied eBPF: [11](./11-ebpf-networking-guide.md)-[14](./14-go-development.md)
5. Kubernetes: [16](./16-kubernetes-networking.md), [17](./17-cni-plugin-development.md), [18](./18-network-policy-enforcement.md)
6. Production: [19](./19-control-plane-agent-patterns.md) (the agent) and [20](./20-production-load-balancer-datapath.md) (the datapath)
7. Keep [07-quick-reference.md](./07-quick-reference.md) open while practicing

---

## Recommended Starting Point

If you're comfortable with the networking basics from previous modules:

**Jump to Module 5 (Go Networking)** → **Module 15 (Network Namespaces — build the lab first)** → **Module 6 (eBPF Fundamentals)** → **Modules 8-14 (Advanced eBPF)** → **Modules 16-18 (Kubernetes, CNI, policy)** → **Modules 19-20 (agent + production datapath)**

If you need to reinforce fundamentals:

**Start with Module 1 (OSI Model)** → progress through 1-5 → **Module 15** → 6 → continue with 8-14, then 16-20, using Module 7 as a command reference throughout.

Either way, Module 15 comes before Module 6: the eBPF modules assume a namespace/veth lab to attach programs to, and Module 15 is where you build one.
