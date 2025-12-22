# Project Structure

This document describes the improved project structure following Go best practices.

## Directory Layout

```
api/
├── main.go                      # Application entry point
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── Dockerfile                   # Container image definition
├── .dockerignore               # Docker build exclusions
├── README.md                    # API documentation
└── internal/                    # Private application code
    ├── config/                  # Configuration management
    │   └── config.go           # Config loading and validation
    ├── models/                  # Data models
    │   └── responses.go        # API response types
    ├── logger/                  # Logging infrastructure
    │   └── ecs.go             # ECS-compliant logger
    ├── middleware/              # HTTP middleware
    │   └── logger.go           # Request logging middleware
    ├── handlers/                # HTTP request handlers
    │   └── handlers.go         # Ping, health, ready endpoints
    └── server/                  # HTTP server setup
        └── server.go           # Server initialization and lifecycle
```

## Package Descriptions

### `main`
- **Purpose**: Application bootstrap and dependency wiring
- **Responsibilities**:
  - Load configuration
  - Initialize logger
  - Wire dependencies via dependency injection
  - Start HTTP server
  - Handle graceful shutdown

### `internal/config`
- **Purpose**: Configuration management
- **Key Features**:
  - Environment variable loading with defaults
  - Configuration validation
  - Type-safe configuration access
  - Centralized constants (service name, version, ECS version)

### `internal/models`
- **Purpose**: Shared data structures
- **Contains**:
  - API response types (`Response`, `HealthResponse`, `ErrorResponse`)
  - Ensures consistent API contracts
  - Easy JSON marshaling with struct tags

### `internal/logger`
- **Purpose**: Structured logging with ECS compliance
- **Key Features**:
  - Custom `slog.Handler` implementation
  - Automatic field mapping to ECS specification
  - Service metadata injection
  - JSON output for log aggregation

### `internal/middleware`
- **Purpose**: HTTP middleware components
- **Contains**:
  - Request logger middleware with structured logging
  - Captures request/response metrics
  - Integrates with ECS logging

### `internal/handlers`
- **Purpose**: HTTP request handlers
- **Pattern**: Dependency injection via struct
- **Benefits**:
  - Testable (dependencies can be mocked)
  - Stateful (maintains start time for uptime)
  - Centralized error handling via `writeJSON`

### `internal/server`
- **Purpose**: HTTP server lifecycle management
- **Responsibilities**:
  - Router setup with middleware chain
  - Server configuration (timeouts, address)
  - Graceful shutdown handling

## Design Patterns

### 1. Dependency Injection
All packages receive dependencies through constructors:
```go
handler := handlers.New(logger, startTime)
server := server.New(config, logger, handler)
```

**Benefits:**
- Loose coupling
- Easy testing
- Clear dependencies
- No global state

### 2. Internal Package
The `internal/` directory prevents external imports:
- Enforces encapsulation
- Allows refactoring without breaking external code
- Standard Go project layout

### 3. Configuration Object
Single config struct passed through application:
- Type-safe access
- Validation at startup
- Single source of truth
- Easy to test

### 4. Handler Struct Pattern
Handlers are methods on a struct:
```go
type Handler struct {
    logger *slog.Logger
    startTime time.Time
}
```

**Benefits:**
- Share state across handlers
- Inject dependencies once
- Easy to test with mock dependencies

### 5. Middleware Functions
Middleware as higher-order functions:
```go
func Logger(logger *slog.Logger) func(next http.Handler) http.Handler
```

**Benefits:**
- Composable
- Reusable
- Takes dependencies as parameters

## Best Practices Implemented

### Code Organization
- ✅ Standard project layout (`internal/`)
- ✅ Package per responsibility
- ✅ Clear separation of concerns
- ✅ No cyclic dependencies

### Configuration
- ✅ Environment-based config
- ✅ Sensible defaults
- ✅ Validation at startup
- ✅ Type-safe access

### Error Handling
- ✅ Errors propagated to main
- ✅ Graceful shutdown on errors
- ✅ Structured error logging
- ✅ Proper exit codes

### Logging
- ✅ Structured logging (JSON)
- ✅ ECS compliance
- ✅ Context propagation (request IDs)
- ✅ Appropriate log levels

### HTTP Server
- ✅ Graceful shutdown
- ✅ Configurable timeouts
- ✅ Middleware composition
- ✅ Request/response logging

### Testing-Ready
- ✅ Dependency injection
- ✅ No global state
- ✅ Interfaces where beneficial
- ✅ Small, focused packages

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `ENVIRONMENT` | `development` | Environment name for logging |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `SHUTDOWN_TIMEOUT` | `30s` | Graceful shutdown timeout |

## Building and Running

### Development
```bash
# Install dependencies
go mod download

# Run the application
go run main.go

# Build binary
go build -o app main.go
```

### Production Build
```bash
# Optimized build
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o app main.go
```

### Docker
```bash
# Build image
docker build -t orchestrated-ping:latest .

# Run container
docker run -p 8080:8080 -e ENVIRONMENT=production orchestrated-ping:latest
```

## Migration from Old Structure

The refactoring involved:

1. **Extracted configuration** → `internal/config`
2. **Separated models** → `internal/models`
3. **Modularized logging** → `internal/logger`
4. **Extracted middleware** → `internal/middleware`
5. **Separated handlers** → `internal/handlers`
6. **Created server package** → `internal/server`
7. **Simplified main.go** → Bootstrap only
8. **Removed global variables** → Dependency injection

This structure scales better as the application grows and makes testing significantly easier.
