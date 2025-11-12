# OED MCP Server - Transport Deployment Guide

Comprehensive deployment guide for the OED MCP server with standardized transport configuration supporting stdio and HTTP/SSE modes.

## Table of Contents

- [Quick Start](#quick-start)
- [Transport Modes](#transport-modes)
- [Deployment Scenarios](#deployment-scenarios)
- [Health Monitoring](#health-monitoring)
- [Security](#security)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Stdio Only (Default)

For local Claude Desktop usage:

```bash
./oedmcp
```

### Dual Transport (Stdio + HTTP)

For monitored deployments:

```bash
./oedmcp --enable-http --http-addr 127.0.0.1:7087
```

## Transport Modes

The OED MCP server supports two transport modes that can run simultaneously:

### Stdio Transport

- **Purpose**: Standard MCP protocol communication
- **Use case**: Claude Desktop, direct process communication
- **Always enabled**: Yes (cannot be disabled)
- **Configuration**: None needed

### HTTP Transport

- **Purpose**: Server-Sent Events (SSE) for MCP protocol + health/metrics endpoints
- **Use case**: Cloudflare tunnels, monitoring, containerized deployments
- **Enabled by**: `--enable-http` flag
- **Default bind**: `127.0.0.1:7087` (localhost only)

**HTTP Endpoints**:
- `/sse` - MCP protocol over SSE
- `/message` - MCP protocol message endpoint
- `/health` - Health check (JSON)
- `/metrics` - Prometheus metrics

## Deployment Scenarios

### Claude Desktop (Local)

**Configuration**: `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)

#### Stdio Only (Recommended)

```json
{
  "mcpServers": {
    "oed": {
      "command": "/usr/local/bin/oedmcp",
      "env": {
        "OED_DATA_PATH": "/Users/you/data/oed2",
        "OED_INDEX_PATH": "/Users/you/data/oed2index"
      }
    }
  }
}
```

#### With HTTP Monitoring

```json
{
  "mcpServers": {
    "oed": {
      "command": "/usr/local/bin/oedmcp",
      "args": [
        "--enable-http",
        "--http-addr", "127.0.0.1:7087",
        "--log-level", "info"
      ],
      "env": {
        "OED_DATA_PATH": "/Users/you/data/oed2",
        "OED_INDEX_PATH": "/Users/you/data/oed2index"
      }
    }
  }
}
```

Then monitor with:
```bash
curl http://127.0.0.1:7087/health | jq .
```

### Docker Container

#### Dockerfile

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY . .
RUN go mod download
RUN go build -o oedmcp

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app
COPY --from=builder /build/oedmcp /app/

# Create data volume mount point
VOLUME ["/data"]

# Expose HTTP port
EXPOSE 7087

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:7087/health || exit 1

# Run with HTTP enabled
ENTRYPOINT ["/app/oedmcp"]
CMD ["--enable-http", "--http-addr", "0.0.0.0:7087"]
```

#### docker-compose.yml

```yaml
version: '3.8'

services:
  oedmcp:
    build: .
    ports:
      - "127.0.0.1:7087:7087"
    environment:
      - OED_DATA_PATH=/data/oed2
      - OED_INDEX_PATH=/data/oed2index
    volumes:
      - /path/to/oed/data:/data:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:7087/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

**Run**:
```bash
docker-compose up -d
docker-compose ps  # Check health status
curl http://localhost:7087/health | jq .
```

### Kubernetes

#### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: oedmcp
  labels:
    app: oedmcp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: oedmcp
  template:
    metadata:
      labels:
        app: oedmcp
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "7087"
        prometheus.io/path: "/metrics"
    spec:
      containers:
      - name: oedmcp
        image: oedmcp:latest
        args:
          - --enable-http
          - --http-addr=0.0.0.0:7087
          - --log-level=info
        ports:
        - containerPort: 7087
          name: http
          protocol: TCP
        env:
        - name: OED_DATA_PATH
          value: /data/oed2
        - name: OED_INDEX_PATH
          value: /data/oed2index
        volumeMounts:
        - name: oed-data
          mountPath: /data
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: http
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: oed-data
        persistentVolumeClaim:
          claimName: oed-data-pvc
```

### Systemd Service

Create `/etc/systemd/system/oedmcp.service`:

```ini
[Unit]
Description=OED MCP Server
After=network.target

[Service]
Type=simple
User=oedmcp
Environment="OED_DATA_PATH=/var/lib/oedmcp/oed2"
Environment="OED_INDEX_PATH=/var/lib/oedmcp/oed2index"
ExecStart=/opt/oedmcp/oedmcp --enable-http --http-addr 127.0.0.1:7087
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

## Health Monitoring

### Health Check Endpoint

```bash
curl http://localhost:7087/health | jq .
```

**Healthy Response**:
```json
{
  "status": "healthy",
  "service": "oedmcp",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "connections": {
    "data_file": {
      "status": "connected",
      "readable": true
    }
  }
}
```

**Degraded Response**:
```json
{
  "status": "degraded",
  "connections": {
    "data_file": {
      "status": "disconnected",
      "last_error": "file not found"
    }
  }
}
```

## Security

### Bearer Token Authentication

```bash
export OED_HTTP_AUTH_TOKEN="your-secret-token"
./oedmcp --enable-http --http-auth-type bearer
```

### Basic Authentication

```bash
export OED_HTTP_AUTH_USER="admin"
export OED_HTTP_AUTH_PASS="password"
./oedmcp --enable-http --http-auth-type basic
```

## Troubleshooting

### Health Check Returns Degraded

Check data file paths and permissions:
```bash
ls -l $OED_DATA_PATH $OED_INDEX_PATH
```

### Port Already in Use

Use a different port:
```bash
./oedmcp --enable-http --http-addr 127.0.0.1:7088
```

---

For more details, see [README.md](README.md)
