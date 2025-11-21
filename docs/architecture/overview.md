# Architecture Overview

masque-vpn is a VPN implementation based on the MASQUE (CONNECT-IP) protocol using QUIC transport.

## Components

- **VPN Server**: Handles client connections and traffic routing
- **VPN Client**: Connects to server and routes local traffic
- **Web UI**: Browser-based client management and configuration
- **Certificate System**: PKI-based authentication using mutual TLS

## Protocol Stack

- **Application Layer**: MASQUE CONNECT-IP (RFC 9484)
- **Transport Layer**: QUIC (RFC 9000)
- **Security Layer**: TLS 1.3
- **Network Layer**: IP tunneling via TUN device

## Data Flow

```
Client Application
    ↓
TUN Device (Client)
    ↓
MASQUE CONNECT-IP over QUIC
    ↓
VPN Server
    ↓
TUN Device (Server)
    ↓
Target Network
```

