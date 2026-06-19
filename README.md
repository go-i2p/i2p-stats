# I2P Router Statistics Web Application

A lightweight, JavaScript-free I2P router statistics dashboard written in Go, exposed directly as a SAMv3-managed I2P hidden service.

## Features

- **JS-Free & Modern Interface**: Curated premium dark HSL aesthetic with zero client-side JavaScript. Uses Grid/Flexbox layouts, glassmorphism design, and micro-hover transitions.
- **Embedded Static Assets**: Fully standalone single binary deployment using Go `embed.FS` for CSS and HTML templates.
- **In-Memory Cache**: Protects your I2P router's HTTP/JSON-RPC API against high volume requests using `singleflight` coalescing and an RWMutex-backed TTL cache.
- **No TCP Binding**: Runs strictly over SAMv3 hidden service stream listeners. Does not listen on any local TCP ports, rendering external network attacks structurally impossible.
- **Graceful Shutdown**: Wires cleanly into system signals (`SIGINT`/`SIGTERM`) to release sessions and shutdown listeners cleanly.

## Prerequisites

1. **SAMv3 Bridge Enabled** in your I2P router (typically `127.0.0.1:7656`).
2. **I2PControl Enabled** in your I2P router (typically `127.0.0.1:7650`) with a configured password.

## Configuration

All configuration is driven via environment variables:

| Environment Variable | Default | Description |
|---|---|---|
| `I2PCONTROL_ENDPOINT` | `http://127.0.0.1:7650` | HTTP Endpoint for the I2PControl API |
| `I2PCONTROL_PASSWORD` | `itoopie` | Password for I2PControl authentication |
| `SAM_ADDR` | `127.0.0.1:7656` | TCP Address of the SAMv3 bridge |
| `TUNNEL_NAME` | `i2pstats` | Name of the I2P hidden service tunnel |
| `KEY_DIR` | `./keys` | Directory where persistent destination keys are stored |
| `REFRESH_INTERVAL` | `30s` | Frequency at which statistics are refreshed |

## Running the Application

### 1. Build the Binary
```bash
go build -o i2pstats .
```

### 2. Run the Service
```bash
export I2PCONTROL_PASSWORD="your-secure-password"
./i2pstats
```

On first startup, the service will generate private keys in `./keys` and log the stable `.b32.i2p` destination address:
```
2026/06/19 17:15:00 INFO starting i2pstats sam=127.0.0.1:7656 i2pcontrol=http://127.0.0.1:7650 tunnel=i2pstats refresh=30s
2026/06/19 17:15:03 INFO hidden service ready b32=exampledestinationhash.b32.i2p
```

You can then view the dashboard by accessing the `.b32.i2p` address inside an I2P-configured browser (e.g. I2P Browser, Tor Browser configured for I2P, or via a HTTP proxy configured to route `.i2p` requests).