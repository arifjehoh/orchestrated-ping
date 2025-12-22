# Orchestrated Ping API

A lightweight, production-ready REST API built with Go that implements health checking and ping-pong functionality. Designed for cloud-native deployments with structured logging following the Elastic Common Schema (ECS) specification.

## Overview

This API serves as a demonstration service for learning cloud-native practices including containerization, Kubernetes deployment, monitoring, and observability. It showcases modern Go development practices with structured logging, middleware patterns, and production-ready configurations.

## Features

- **RESTful API** with JSON responses
- **ECS-compliant structured logging** for seamless integration with Elasticsearch, Kibana, and other ECS-aware tools
- **Kubernetes-ready health probes** (liveness and readiness)
- **Request tracing** with unique request IDs
- **Production middleware** including timeout, recovery, and request logging
- **Zero external dependencies** for logging (uses Go 1.21+ stdlib `log/slog`)
- **Minimal footprint** - compiles to a single static binary

## Architecture

### Tech Stack
- **Go 1.21+** - Application runtime
- **go-chi/chi v5** - Lightweight HTTP router
- **log/slog** - Structured logging (Go stdlib)

### Middleware Chain
1. **RequestID** - Generates unique ID for request tracing
2. **RealIP** - Extracts real client IP from headers
3. **StructuredLogger** - Custom ECS-formatted request logging
4. **Recoverer** - Panic recovery middleware
5. **Timeout** - 60-second request timeout

## API Endpoints

### `GET /ping`
Returns a pong response to verify the service is responding.

**Response:**
```json
{
  "status": "success",
  "message": "pong",
  "time": "2025-12-22T10:30:00.123Z"
}
```

**Use Case:** Simple connectivity test, application functionality verification

---

### `GET /health`
Liveness probe for Kubernetes. Indicates whether the application is running.

**Response:**
```json
{
  "status": "healthy",
  "uptime": "2h15m30s"
}
```

**Use Case:** Kubernetes liveness probe - determines if the pod should be restarted

**Kubernetes Configuration:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
```

---

### `GET /ready`
Readiness probe for Kubernetes. Indicates whether the application is ready to serve traffic.

**Response:**
```json
{
  "status": "ready",
  "message": "application is ready to serve traffic",
  "time": "2025-12-22T10:30:00.123Z"
}
```

**Use Case:** Kubernetes readiness probe - determines if the pod should receive traffic

**Kubernetes Configuration:**
```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

## ECS Logging

All logs are formatted according to the [Elastic Common Schema (ECS) v8.11.0](https://www.elastic.co/guide/en/ecs/current/index.html) specification for standardized observability.

### Log Fields

| ECS Field | Description | Example |
|-----------|-------------|---------|
| `@timestamp` | Event timestamp (UTC, RFC3339 nano) | `2025-12-22T10:30:00.123456Z` |
| `ecs.version` | ECS specification version | `8.11.0` |
| `message` | Human-readable log message | `request completed` |
| `log.level` | Severity level | `INFO`, `DEBUG`, `ERROR` |
| `service.name` | Service identifier | `orchestrated-ping` |
| `service.version` | Service version | `1.0.0` |
| `service.environment` | Deployment environment | `production`, `development` |
| `http.request.method` | HTTP method | `GET`, `POST` |
| `http.response.status_code` | HTTP status code | `200`, `404` |
| `http.response.body.bytes` | Response size in bytes | `58` |
| `url.path` | Request path | `/ping` |
| `client.address` | Client IP address | `192.168.1.100` |
| `event.duration` | Request duration (nanoseconds) | `125000000` |
| `trace.id` | Unique request identifier | `abc123xyz` |
| `server.port` | Server listening port | `8080` |
| `error.message` | Error details | `connection timeout` |

### Example Log Output

**Startup:**
```json
{
  "@timestamp": "2025-12-22T10:30:00.123456Z",
  "ecs.version": "8.11.0",
  "message": "starting server",
  "log.level": "INFO",
  "service.name": "orchestrated-ping",
  "service.version": "1.0.0",
  "server.port": "8080",
  "service.environment": "production"
}
```

**Request:**
```json
{
  "@timestamp": "2025-12-22T10:30:05.789012Z",
  "ecs.version": "8.11.0",
  "message": "request completed",
  "log.level": "INFO",
  "service.name": "orchestrated-ping",
  "service.version": "1.0.0",
  "http.request.method": "GET",
  "url.path": "/ping",
  "http.response.status_code": 200,
  "http.response.body.bytes": 58,
  "event.duration": 125000000,
  "client.address": "127.0.0.1:54321",
  "trace.id": "abc123xyz"
}
```

## Configuration

The application is configured via environment variables:

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `PORT` | HTTP server port | `8080` | No |
| `ENVIRONMENT` | Deployment environment (for logging) | `development` | No |

## Development

### Prerequisites
- Go 1.21 or higher
- Git

### Setup

1. **Install dependencies:**
```bash
go mod download
```

2. **Run the application:**
```bash
go run main.go
```

3. **Test endpoints:**
```bash
# Ping endpoint
curl http://localhost:8080/ping

# Health check
curl http://localhost:8080/health

# Readiness check
curl http://localhost:8080/ready
```

### Build

**Standard build:**
```bash
go build -o app main.go
```

**Optimized build (smaller binary):**
```bash
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o app main.go
```

**Cross-compilation for Linux:**
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o app main.go
```

## Production Deployment

### Docker
See the root `Dockerfile` for containerization. The image uses a multi-stage build with `scratch` as the final base for minimal size (~5-10MB).

### Environment Variables
```bash
export PORT=8080
export ENVIRONMENT=production
```

### Resource Requirements
- **Memory**: ~10-20MB
- **CPU**: Minimal (suitable for fractional CPU allocations)

## Project Structure

```
api/
├── main.go                      # Application entry point
├── internal/                    # Private application packages
│   ├── config/                  # Configuration management
│   ├── models/                  # Data models and types
│   ├── logger/                  # ECS-compliant structured logging
│   ├── middleware/              # HTTP middleware
│   ├── handlers/                # HTTP request handlers
│   └── server/                  # Server initialization and lifecycle
├── go.mod                       # Go module definition
├── Dockerfile                   # Container image definition
├── README.md                    # This file
└── ARCHITECTURE.md              # Detailed architecture documentation

```

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed information about the project structure, design patterns, and best practices implemented.

## Code Components

### `main` package
Application entry point that:
- Loads and validates configuration
- Initializes ECS-formatted structured logger
- Wires dependencies via dependency injection
- Starts HTTP server
- Handles graceful shutdown with signal handling

### `config` package
Configuration management with:
- Environment variable loading
- Sensible defaults
- Type-safe configuration access
- Startup validation

### `logger` package
Custom `slog.Handler` implementation:
- ECS specification compliance
- Automatic field mapping
- Service metadata injection
- JSON output for log aggregation

### `middleware` package
HTTP middleware components:
- Request logger with structured logging
- Request/response metrics capture
- Integration with ECS logging

### `handlers` package
HTTP request handlers with dependency injection:
- `Ping()` - Simple pong response
- `Health()` - Liveness check with uptime
- `Ready()` - Readiness check for load balancers
- Centralized JSON response handling

### `server` package
HTTP server lifecycle management:
- Router setup with middleware chain
- Server configuration (timeouts, address)
- Graceful shutdown handling
- Clean separation from business logic

## Observability Integration

### Prometheus Metrics
While not currently implemented, the application structure supports adding Prometheus metrics:
- Request count by endpoint and status code
- Request duration histograms
- Active connections gauge

### Log Aggregation
ECS-formatted logs can be consumed by:
- **Elasticsearch** - Direct indexing
- **Logstash** - ECS pipeline processing
- **Filebeat** - Log shipping with ECS fields
- **Fluentd/Fluent Bit** - Log forwarding
- **Cloud providers** - GCP Cloud Logging, AWS CloudWatch, Azure Monitor

### Distributed Tracing
Request IDs (`trace.id` in logs) enable correlation across services. Can be extended to integrate with OpenTelemetry, Jaeger, or Zipkin.

## Future Enhancements

- [ ] Prometheus metrics endpoint (`/metrics`)
- [ ] OpenTelemetry instrumentation
- [ ] Graceful shutdown handling
- [ ] Configuration validation on startup
- [ ] Rate limiting middleware
- [ ] CORS support
- [ ] API versioning

## License

See the root `LICENSE` file for details.

## Resources

- [go-chi router](https://github.com/go-chi/chi)
- [Elastic Common Schema (ECS)](https://www.elastic.co/guide/en/ecs/current/index.html)
- [Go slog package](https://pkg.go.dev/log/slog)
- [Kubernetes probes](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/)
