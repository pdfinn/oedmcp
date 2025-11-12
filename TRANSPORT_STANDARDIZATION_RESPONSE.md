# MCP Transport Standardization - OED MCP Server Response

## Executive Summary

The OED MCP server fully supports the proposed transport standardization. Implementation is straightforward since the underlying mcp-go library already provides the necessary HTTP transport capabilities. The OED server has minimal external dependencies, making degraded mode implementation simpler than other services.

**Status**: ✅ Ready to implement
**Timeline**: 1-2 days for implementation + testing
**Breaking Changes**: None (additive only)

---

## Current State Analysis

### OED MCP Server Profile

| Aspect | Current State |
|--------|---------------|
| **Language** | Go |
| **Current Flags** | None (uses environment variables for data paths) |
| **Transport** | stdio only (via `server.ServeStdio()`) |
| **Supports Dual Transport?** | ❌ No (but library supports it) |
| **External Dependencies** | None (self-contained dictionary files) |
| **Startup Resilience** | ❌ Fails if data files missing/unreadable |

### Current Implementation

```go
// main.go:447
if err := server.ServeStdio(s); err != nil {
    fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
    os.Exit(1)
}
```

**Key Observations**:
- Single-transport (stdio only)
- No command-line flags for transport configuration
- Hard failure on data file errors (not degraded mode compliant)
- Uses mcp-go v0.39.1 which includes SSE and StreamableHTTP support

---

## Proposed Standard Compliance Assessment

### ✅ What Works in Our Favor

1. **Library Support**: mcp-go already provides:
   - `ServeStdio()` for stdio transport
   - `NewSSEServer()` for HTTP/SSE transport
   - Concurrent transport operation examples

2. **No External Runtime Dependencies**:
   - Unlike QGIS MCP (needs QGIS Desktop) or TAK MCP (needs TAK server)
   - Dictionary files are static data, not live connections
   - Simplifies degraded mode implementation

3. **Go's Concurrency Model**:
   - Trivial to run HTTP server in goroutine
   - stdio can remain on main thread (blocking)

### ⚠️ Considerations

1. **Degraded Mode Definition**:
   - For OED MCP, "degraded" means dictionary files unavailable/unreadable
   - Should we start server even if data files missing?
   - Current behavior: fail fast with clear error message
   - Proposed: start in degraded mode, return errors on tool calls

2. **Port Allocation**:
   - Need to assign OED MCP a port in NERVA's 7080-7089 range
   - Suggest: `7087` (currently unallocated in examples)

3. **Health Check Semantics**:
   - What should "healthy" mean for a dictionary server?
     - Option A: Server running + data files readable
     - Option B: Server running only
   - Recommend: Option A (data files are the "dependency")

---

## Answers to Discussion Questions

### 1. Port Allocation

**Question**: Should we enforce the NERVA port standard (7080-7089) in the flag defaults?

**Answer**: ✅ **Yes, absolutely**. This provides:
- Predictable deployment patterns
- Easy firewall/proxy configuration
- Clear service discovery

**Proposal for OED MCP**: Default to `127.0.0.1:7087`

**Rationale**: Checking NERVA docs, ports 7080-7086 appear allocated to other services:
- 7081: TAK MCP
- 7082: OSM MCP
- 7083: Exa MCP
- 7084: AIS MCP
- 7085: QGIS MCP
- 7086: (reserved)
- **7087: OED MCP (proposed)**

### 2. Authentication

**Question**: Do we need HTTP authentication for internal MCP servers behind Cloudflare tunnel?

**Answer**: **Start with `--http-auth-type none` as default, support bearer/basic for future**

**Rationale for OED MCP**:
- Dictionary data already requires user to have legal access
- If behind Cloudflare tunnel, tunnel provides authentication
- For internal NERVA deployments, network isolation is sufficient
- BUT: Framework should support auth for future public API scenarios

**Recommendation**:
- Implement `--http-auth-type` flag with `none|bearer|basic` options
- Default: `none`
- Document how to enable if needed

### 3. Metrics

**Question**: Should Prometheus metrics be required or optional?

**Answer**: **Optional but strongly recommended**

**OED MCP Metrics Would Include**:
```
# HELP oedmcp_lookups_total Total number of word lookups
# TYPE oedmcp_lookups_total counter
oedmcp_lookups_total{result="found"} 1234
oedmcp_lookups_total{result="not_found"} 56

# HELP oedmcp_lookup_duration_seconds Time to look up words
# TYPE oedmcp_lookup_duration_seconds histogram
oedmcp_lookup_duration_seconds_bucket{le="0.01"} 1200
oedmcp_lookup_duration_seconds_bucket{le="0.05"} 1289
oedmcp_lookup_duration_seconds_bucket{le="0.1"} 1290
oedmcp_lookup_duration_seconds_sum 12.5
oedmcp_lookup_duration_seconds_count 1290

# HELP oedmcp_data_file_readable Dictionary data file accessibility
# TYPE oedmcp_data_file_readable gauge
oedmcp_data_file_readable 1

# HELP oedmcp_index_size_bytes Size of loaded dictionary index
# TYPE oedmcp_index_size_bytes gauge
oedmcp_index_size_bytes 5872640
```

**Implementation**: Use prometheus/client_golang library (standard in Go ecosystem)

### 4. WebSocket vs SSE

**Question**: Should we standardize on SSE for HTTP transport, or support both?

**Answer**: **SSE only** (matches Anthropic's MCP-over-HTTP spec)

**Rationale**:
- MCP specification uses SSE for HTTP transport
- mcp-go's `NewSSEServer()` implements this
- Anthropic API expects SSE when MCP servers behind Cloudflare tunnel
- WebSocket adds complexity without clear benefit for request/response pattern

**OED MCP Position**: Use `NewSSEServer()` from mcp-go library

### 5. Health Check Depth

**Question**: Should health checks probe dependencies, or just report server process health?

**Answer**: **Probe dependencies** (dictionary files in OED MCP's case)

**Proposed Health Response**:
```json
{
  "status": "healthy",
  "service": "oedmcp",
  "version": "1.0.0",
  "uptime_seconds": 86400,
  "connections": {
    "data_file": {
      "status": "connected",
      "path": "/path/to/oed2",
      "size_mb": 520,
      "readable": true,
      "last_error": null
    },
    "index_file": {
      "status": "connected",
      "path": "/path/to/oed2index",
      "size_mb": 5.6,
      "entries": 291500,
      "readable": true,
      "last_error": null
    }
  }
}
```

**Degraded Example**:
```json
{
  "status": "degraded",
  "service": "oedmcp",
  "version": "1.0.0",
  "uptime_seconds": 120,
  "connections": {
    "data_file": {
      "status": "disconnected",
      "path": "/path/to/oed2",
      "readable": false,
      "last_error": "open /path/to/oed2: no such file or directory"
    },
    "index_file": {
      "status": "disconnected",
      "path": "/path/to/oed2index",
      "readable": false,
      "last_error": "open /path/to/oed2index: no such file or directory"
    }
  }
}
```

**Why This Approach**:
- NERVA monitoring needs to know if dictionary is accessible, not just if process is alive
- Helps diagnose issues: server running but data files moved/deleted
- Aligns with "degraded mode" philosophy

### 6. Backward Compatibility

**Question**: Timeline for removing deprecated --transport flag from Python servers?

**Answer**: **Not applicable to OED MCP** (no existing flags to deprecate)

**General Recommendation**:
- Phase 1-2: Support both old and new flags (6 months)
- Phase 3: Deprecation warnings (3 months)
- Phase 4: Remove old flags (after all deployments migrated)

For OED MCP: Clean slate - implement standard flags only

---

## Implementation Plan for OED MCP

### Phase 1: Add Flag Interface (1 day)

**Changes to main.go**:

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
    // ... existing imports
)

var (
    enableHTTP   = flag.Bool("enable-http", false, "Enable HTTP transport in addition to stdio")
    httpAddr     = flag.String("http-addr", "127.0.0.1:7087", "HTTP server bind address")
    httpAuthType = flag.String("http-auth-type", "none", "HTTP authentication: none, bearer, basic")
    logLevel     = flag.String("log-level", "info", "Log level: debug, info, warning, error")
)

func main() {
    flag.Parse()

    // ... existing initialization code ...

    // Create MCP server
    s := server.NewMCPServer(/* ... */)

    // ... add tools ...

    // Start HTTP server if enabled
    if *enableHTTP {
        go startHTTPServer(s, *httpAddr, *httpAuthType, oedDict)
    }

    // Start stdio transport (blocks)
    if err := server.ServeStdio(s); err != nil {
        fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
        os.Exit(1)
    }
}
```

**Backward Compatibility**: ✅ Perfect
- No existing flags to conflict with
- Stdio remains default behavior
- HTTP is opt-in via `--enable-http`

### Phase 2: Implement Dual Transport (1 day)

**HTTP Server Implementation**:

```go
func startHTTPServer(mcpServer *server.MCPServer, addr string, authType string, dict *dict.OEDDict) {
    // Create SSE server for MCP protocol
    sseServer := server.NewSSEServer(mcpServer,
        server.WithStaticBasePath("/"),
        server.WithSSEEndpoint("/sse"),
        server.WithMessageEndpoint("/message"),
    )

    mux := http.NewServeMux()

    // Mount SSE handler for MCP protocol
    mux.Handle("/sse", sseServer)
    mux.Handle("/message", sseServer)

    // Add health endpoint
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        handleHealth(w, r, dict)
    })

    // Add metrics endpoint (optional)
    mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
        handleMetrics(w, r)
    })

    // Apply authentication middleware if configured
    handler := applyAuth(mux, authType)

    log.Printf("Starting HTTP server on %s", addr)
    httpServer := &http.Server{
        Addr:    addr,
        Handler: handler,
    }

    if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Printf("HTTP server error: %v", err)
    }
}
```

**Testing**:
```bash
# Test stdio only (default)
./oedmcp
# stdio MCP protocol works ✓

# Test dual transport
./oedmcp --enable-http --http-addr 0.0.0.0:7087
# stdio works ✓
# curl http://localhost:7087/health works ✓
# SSE endpoint works ✓
```

### Phase 3: Implement Degraded Mode (1 day)

**Changes to dict/oed.go**:

```go
type OEDDict struct {
    dataPath  string
    indexPath string
    index     map[string]int64
    dataFile  *os.File
    healthy   bool        // NEW: health status
    lastError error       // NEW: last error encountered
    mu        sync.RWMutex
}

func (d *OEDDict) HealthStatus() (healthy bool, connections map[string]interface{}) {
    d.mu.RLock()
    defer d.mu.RUnlock()

    connections = map[string]interface{}{
        "data_file": map[string]interface{}{
            "status":     statusString(d.dataFile != nil),
            "path":       d.dataPath,
            "readable":   d.healthy,
            "last_error": errorString(d.lastError),
        },
        "index_file": map[string]interface{}{
            "status":   statusString(d.index != nil),
            "path":     d.indexPath,
            "entries":  len(d.index),
            "readable": d.healthy,
        },
    }

    return d.healthy, connections
}
```

**Startup Change**:
```go
// OLD: Hard failure
oedDict, err := dict.NewOEDDict(cfg.DataPath, cfg.IndexPath)
if err != nil {
    log.Fatalf("Failed to initialize OED dictionary: %v", err)
}

// NEW: Degraded mode support
oedDict, err := dict.NewOEDDict(cfg.DataPath, cfg.IndexPath)
if err != nil {
    log.Printf("WARNING: Failed to initialize OED dictionary: %v", err)
    log.Printf("Server starting in degraded mode - dictionary operations will fail")
    oedDict = dict.NewDegradedDict(cfg.DataPath, cfg.IndexPath, err)
}
```

**Tool Handler Changes**:
```go
// In tool handlers, check health
s.AddTool(lookupTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    if !oedDict.IsHealthy() {
        return mcp.NewToolResultError("Dictionary unavailable: " + oedDict.LastError().Error()), nil
    }

    // ... existing lookup logic ...
})
```

### Phase 4: Health Endpoint Implementation

**Health Handler**:

```go
func handleHealth(w http.ResponseWriter, r *http.Request, dict *dict.OEDDict) {
    healthy, connections := dict.HealthStatus()

    status := "healthy"
    if !healthy {
        status = "degraded"
    }

    response := map[string]interface{}{
        "status":         status,
        "service":        "oedmcp",
        "version":        "1.0.0",
        "uptime_seconds": int(time.Since(startTime).Seconds()),
        "connections":    connections,
    }

    w.Header().Set("Content-Type", "application/json")

    // Return 200 even if degraded (server is running)
    // Monitoring systems check status field for health
    w.WriteHeader(http.StatusOK)

    json.NewEncoder(w).Encode(response)
}
```

**Standard Test Cases**:

```bash
# Test 1: Stdio-only mode (default)
./oedmcp
# ✓ Server starts
# ✓ stdio responds to MCP protocol
# ✓ No HTTP port listening

# Test 2: Dual transport mode
./oedmcp --enable-http --http-addr 0.0.0.0:7087
# ✓ Server starts
# ✓ stdio responds to MCP protocol
# ✓ HTTP health check responds

# Test 3: Health check format
curl http://localhost:7087/health
# ✓ Returns JSON with required fields
# ✓ Status is "healthy"

# Test 4: Degraded mode
mv /path/to/oed2 /path/to/oed2.backup
./oedmcp --enable-http
# ✓ Server starts (doesn't exit)
# ✓ Health returns "degraded" status
# ✓ MCP tool calls return informative errors

# Test 5: Concurrent requests
# Terminal 1: Start server
./oedmcp --enable-http

# Terminal 2: MCP protocol test
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"oed_lookup","arguments":{"word":"test"}}}' | nc localhost 7087

# Terminal 3: Health checks
while true; do curl http://localhost:7087/health; sleep 1; done

# ✓ Both work simultaneously
```

---

## Metrics Implementation (Optional but Recommended)

**Dependencies**:
```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
```

**Metrics Definition**:
```go
var (
    lookupsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "oedmcp_lookups_total",
            Help: "Total number of word lookups",
        },
        []string{"result"}, // "found" or "not_found"
    )

    lookupDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name: "oedmcp_lookup_duration_seconds",
            Help: "Time to look up words",
            Buckets: prometheus.DefBuckets,
        },
    )

    dataFileReadable = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "oedmcp_data_file_readable",
            Help: "Dictionary data file accessibility",
        },
    )
)
```

**Metrics Endpoint**:
```go
mux.Handle("/metrics", promhttp.Handler())
```

---

## Migration Timeline

### Week 1: Implementation
- **Day 1**: Add flag interface, implement basic HTTP server
- **Day 2**: Implement health endpoint with proper response format
- **Day 3**: Implement degraded mode support
- **Day 4**: Add metrics (optional)
- **Day 5**: Testing and documentation

### Week 2: Integration Testing
- Test with NERVA deployment stack
- Verify Claude Desktop integration still works
- Test HTTP transport via Cloudflare tunnel (if applicable)
- Performance testing (concurrent stdio + HTTP)

### Week 3: Documentation & Rollout
- Update README.md with new flags
- Add DEPLOYMENT.md with health check examples
- Update Claude Desktop config examples
- Deploy to staging environment

### Week 4: Production Rollout
- Deploy to production NERVA stack
- Monitor health endpoints
- Verify metrics collection
- Performance validation

---

## Breaking Changes Assessment

✅ **NONE** - This is purely additive:

1. **Existing Deployments**: Continue working unchanged
   - No flags required for stdio-only operation
   - Environment variables still work for config
   - MCP protocol unchanged

2. **New Deployments**: Can opt-in to HTTP
   - Add `--enable-http` flag
   - Configure health monitoring
   - Enable metrics collection

3. **Future Enhancements**: Supported
   - Authentication can be added later
   - Metrics are optional
   - Can evolve health check format

---

## Concerns & Risks

### 1. Startup Performance
**Concern**: Will dual transport slow server startup?

**Mitigation**:
- HTTP server starts in goroutine (non-blocking)
- Dictionary loading is the bottleneck (~2 seconds for 5.6MB index)
- HTTP startup adds <10ms

### 2. Security
**Concern**: Exposing HTTP endpoint without authentication?

**Mitigation**:
- Default bind to 127.0.0.1 (localhost only)
- Implement `--http-auth-type` flag
- Document network isolation requirements
- Behind Cloudflare tunnel in NERVA deployment

### 3. Error Handling in Degraded Mode
**Concern**: Should server start if dictionary files missing?

**Options**:
- **A**: Fail fast (current behavior) - clear errors but blocks NERVA startup
- **B**: Degraded mode - NERVA starts but OED tools return errors

**Recommendation**: **Option B** (degraded mode)
- Aligns with NERVA resilience goals
- Health endpoint clearly indicates status
- Allows system to start even if OED data misconfigured
- BUT: Document this as behavior change (was fail-fast before)

### 4. mcp-go Library Limitations
**Concern**: Does mcp-go v0.39.1 properly support concurrent transports?

**Validation Needed**:
- Check mcp-go examples for dual transport patterns
- Test concurrent stdio + HTTP requests
- May need to upgrade mcp-go version

**Action**: Review mcp-go documentation and examples before implementation

---

## Language-Specific Constraints (Go)

### ✅ Advantages

1. **Goroutines**: Trivial concurrency for HTTP server
2. **Standard Library**: Excellent HTTP server support
3. **Type Safety**: Compile-time checks for health response format
4. **Static Binary**: Easy deployment (single executable)
5. **Cross-Platform**: Same code runs on Linux, macOS, Windows

### ⚠️ Considerations

1. **Context Cancellation**: Ensure clean shutdown
   ```go
   ctx, cancel := context.WithCancel(context.Background())
   defer cancel()

   // Pass ctx to servers for coordinated shutdown
   ```

2. **Error Handling**: Go's explicit errors make degraded mode natural
   ```go
   if err != nil {
       log.Printf("Non-fatal error: %v", err)
       // Continue in degraded mode
   }
   ```

3. **Testing**: Table-driven tests for all scenarios
   ```go
   tests := []struct{
       name string
       flags []string
       wantHealthy bool
   }{
       {"stdio only", []string{}, true},
       {"dual transport", []string{"--enable-http"}, true},
       {"degraded mode", []string{"--enable-http", "--data-path=/nonexistent"}, false},
   }
   ```

---

## Realistic Timeline

### Optimistic (1 week)
- Experienced Go developer
- mcp-go examples available
- No unexpected library issues
- 3-4 hours/day dedicated time

### Realistic (2 weeks)
- Learning mcp-go HTTP transport patterns
- Testing edge cases (concurrent requests, degraded mode)
- Documentation and integration testing
- 2-3 hours/day with other priorities

### Conservative (3 weeks)
- Full testing with NERVA stack
- Coordination with other MCP server teams
- Security review for HTTP endpoint
- Production deployment validation

**Recommendation**: **Plan for 2 weeks**, aim for 1 week

---

## Recommendation

✅ **PROCEED** with standardization for OED MCP Server

**Justification**:
1. **Low Risk**: Additive changes only, no breaking changes
2. **High Value**: Enables NERVA resilience and monitoring
3. **Library Support**: mcp-go already has necessary features
4. **Simple Dependencies**: No external runtime dependencies to manage
5. **Consistent Operations**: Aligns with other Go MCP servers (TAK, OSM)

**Priority**: **High** - OED MCP should be among the first servers to implement this standard as a reference implementation for other Go-based MCP servers.

---

## Next Steps

1. **Immediate** (This Week):
   - Review mcp-go SSE server examples
   - Create feature branch: `feature/transport-standardization`
   - Implement Phase 1 (flag interface)

2. **Short Term** (Next Week):
   - Implement Phases 2-3 (dual transport + degraded mode)
   - Unit tests for all scenarios
   - Integration tests with stdio + HTTP

3. **Medium Term** (Within Month):
   - Deploy to NERVA staging environment
   - Coordinate with TAK/OSM/QGIS teams on common patterns
   - Update documentation for production deployment

4. **Long Term** (Within Quarter):
   - Add Prometheus metrics
   - Performance optimization based on production metrics
   - Consider additional health check details

---

## Questions for NERVA Team

1. **Port Assignment**: Confirm 7087 is appropriate for OED MCP?
2. **Authentication**: NERVA internal deployment needs `--http-auth-type` or network isolation sufficient?
3. **Degraded Mode Philosophy**: Should OED MCP start even if dictionary files missing, or is fail-fast preferred?
4. **Metrics Collection**: Is Prometheus the standard for NERVA monitoring stack?
5. **Deployment Coordination**: When should OED MCP standardization align with other servers (TAK, OSM, QGIS)?

---

## Appendix: Example Configurations

### Claude Desktop Config (Stdio Only - Current)
```json
{
  "mcpServers": {
    "oed": {
      "command": "/usr/local/bin/oedmcp",
      "env": {
        "OED_DATA_PATH": "/data/oed/oed2",
        "OED_INDEX_PATH": "/data/oed/oed2index"
      }
    }
  }
}
```

### Claude Desktop Config (Stdio Only - With New Flags)
```json
{
  "mcpServers": {
    "oed": {
      "command": "/usr/local/bin/oedmcp",
      "args": ["--log-level", "info"],
      "env": {
        "OED_DATA_PATH": "/data/oed/oed2",
        "OED_INDEX_PATH": "/data/oed/oed2index"
      }
    }
  }
}
```

### NERVA Deployment (Dual Transport)
```bash
#!/bin/bash
# Start OED MCP with HTTP health checks enabled

export OED_DATA_PATH=/data/oed/oed2
export OED_INDEX_PATH=/data/oed/oed2index

/usr/local/bin/oedmcp \
  --enable-http \
  --http-addr 127.0.0.1:7087 \
  --http-auth-type none \
  --log-level info
```

### Cloudflare Tunnel Config
```yaml
# cloudflared config for MCP servers
tunnel: <tunnel-id>
credentials-file: /path/to/credentials.json

ingress:
  - hostname: oedmcp.example.com
    service: http://localhost:7087
    originRequest:
      noTLSVerify: true

  - service: http_status:404
```

### Kubernetes Health Probe
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: oedmcp
spec:
  containers:
  - name: oedmcp
    image: oedmcp:latest
    args:
      - --enable-http
      - --http-addr=0.0.0.0:7087
    ports:
    - containerPort: 7087
      name: http
    livenessProbe:
      httpGet:
        path: /health
        port: http
      initialDelaySeconds: 5
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /health
        port: http
      initialDelaySeconds: 3
      periodSeconds: 5
```

### Docker Compose
```yaml
version: '3.8'

services:
  oedmcp:
    image: oedmcp:latest
    command:
      - --enable-http
      - --http-addr=0.0.0.0:7087
    environment:
      - OED_DATA_PATH=/data/oed2
      - OED_INDEX_PATH=/data/oed2index
    volumes:
      - /host/path/to/oed:/data:ro
    ports:
      - "127.0.0.1:7087:7087"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7087/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s
```

---

## Summary

The OED MCP Server is **well-positioned** to implement the proposed transport standardization:

✅ **Technical Feasibility**: mcp-go library supports all required features
✅ **Implementation Complexity**: Low (1-2 weeks)
✅ **Breaking Changes**: None (purely additive)
✅ **Operational Benefits**: Enables NERVA monitoring and resilience
✅ **Timeline**: Can be reference implementation for other Go servers

**Recommendation**: Proceed with implementation, coordinate with TAK/OSM MCP teams for consistent Go patterns.
