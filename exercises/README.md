# Go Network Programming Exercises

Hands-on exercises for Module 5: Go Network Programming.

## Prerequisites

- Go 1.21+ installed
- For Exercise 04 (ICMP Ping): root/sudo access

## Exercises

| # | Exercise | Description | Run Command |
|---|----------|-------------|-------------|
| 01 | [TCP Echo Server](./01-tcp-echo) | Concurrent TCP server with graceful shutdown | `go run ./01-tcp-echo` |
| 02 | [UDP Server](./02-udp-server) | UDP echo server with stats tracking | `go run ./02-udp-server` |
| 03 | [Port Scanner](./03-port-scanner) | Concurrent port scanner with worker pool | `go run ./03-port-scanner -host scanme.nmap.org` |
| 04 | [ICMP Ping](./04-icmp-ping) | ICMP ping with RTT statistics | `sudo go run ./04-icmp-ping -host 8.8.8.8` |
| 05 | [Health Checker](./05-health-checker) | HTTP health monitor for multiple endpoints | `go run ./05-health-checker` |

## Quick Start

```bash
# Install dependencies
go mod tidy

# Run TCP Echo Server
go run ./01-tcp-echo
# In another terminal: nc localhost 8080

# Run Port Scanner
go run ./03-port-scanner -host localhost -start 1 -end 100

# Run Health Checker
go run ./05-health-checker
```

## Learning Objectives

Each exercise teaches specific networking concepts:

- **01-tcp-echo**: TCP listeners, connection handling, goroutines, graceful shutdown
- **02-udp-server**: UDP protocol, connectionless communication, stats tracking
- **03-port-scanner**: Dial timeouts, worker pool pattern, concurrent I/O
- **04-icmp-ping**: Raw sockets, ICMP protocol, privileged operations
- **05-health-checker**: HTTP clients, interface binding, concurrent monitoring

## Project Structure

```
exercises/
├── go.mod
├── README.md
├── 01-tcp-echo/
│   └── main.go
├── 02-udp-server/
│   └── main.go
├── 03-port-scanner/
│   └── main.go
├── 04-icmp-ping/
│   └── main.go
└── 05-health-checker/
    └── main.go
```
