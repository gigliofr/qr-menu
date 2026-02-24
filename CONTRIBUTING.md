# Contributing & Development Guide

**QR Menu System v2.0.0**  
**Developer Documentation**

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Setup](#development-setup)
3. [Building & Testing](#building--testing)
4. [Code Structure](#code-structure)
5. [Best Practices](#best-practices)
6. [Common Tasks](#common-tasks)

---

## Getting Started

### Prerequisites
- **Go**: 1.24 or later
- **PostgreSQL**: Optional (most features work without)
- **Git**: For version control

### Clone & Setup

```bash
# Clone repository
git clone <repo-url>
cd qr-menu

# Download dependencies
go mod download

# Verify setup
go test ./...
```

---

## Development Setup

### Development Configuration

Create `.env.development`:
```bash
ENVIRONMENT=dev
SERVER_PORT=8080
SERVER_HOST=localhost

# Enable caching with short TTL for testing
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=10s
CACHE_QUERY_TTL=30s

# Detailed logging
LOGGER_LEVEL=debug
LOGGER_FORMAT=json
LOGGER_OUTPUT_FILE=./logs/dev.log

# Database (optional)
# DATABASE_DSN=postgres://user:pass@localhost:5432/qrmenu
```

### Run Development Server

```bash
# Simple: Load from default env
go run main.go

# With custom env file
source .env.development && go run main.go

# Windows
$env:ENVIRONMENT='dev'; $env:CACHE_ENABLED='true'; go run main.go
```

### Development Workflow

```bash
# 1. Make code changes
# 2. Run tests (catches errors early)
go test ./...

# 3. Build
go build -o qr-menu .

# 4. Run locally
./qr-menu

# 5. Test API
curl http://localhost:8080/health

# 6. Check cache stats
curl http://localhost:8080/api/admin/cache/stats

# 7. Commit changes
git add .
git commit -m "Feature: description"
```

---

## Building & Testing

### Build Commands

```bash
# Development build
go build -o qr-menu .

# Production build (with optimizations)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o qr-menu-prod .

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o qr-menu.exe .

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o qr-menu-macos .
```

### Testing

```bash
# Run all tests
go test ./...

# Run with verbose output
go test ./... -v

# Run specific package
go test ./pkg/cache/... -v

# Run specific test
go test ./pkg/cache/ -run TestResponseCache -v

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run tests with race detector
go test ./... -race

# Run benchmarks
go test ./... -bench=. -benchmem
```

### Code Quality

```bash
# Format code
go fmt ./...

# Lint code
golangci-lint run ./...

# Find unused imports
go mod tidy

# Check for issues
go vet ./...
```

---

## Code Structure

### Package Organization

```
pkg/
├── app/              # Application lifecycle management
│   └── app.go        # App struct, Start(), Stop()
│
├── cache/            # Caching infrastructure
│   ├── cache.go      # InMemoryCache interface & implementation
│   ├── response_cache.go    # HTTP response caching
│   ├── response_cache_test.go
│   └── *_test.go     # Cache tests
│
├── config/           # Configuration management
│   ├── config.go     # Config struct, Load()
│   └── (no tests - external package)
│
├── container/        # Dependency injection container
│   ├── container.go  # ServiceContainer lifecycle
│   └── (no direct tests - integrated in app)
│
├── errors/           # Error handling
│   ├── errors.go     # Error types, helpers
│   └── (no tests - utilities)
│
├── handlers/         # HTTP request handlers
│   ├── handlers.go   # Handler implementations
│   └── (no direct tests - use integration tests)
│
├── http/             # HTTP utilities
│   ├── utils.go      # Response helpers
│   └── (no tests - utilities)
│
├── middleware/       # HTTP middleware
│   ├── middleware.go           # 7 middleware types
│   ├── *_test.go               # Unit tests
│   ├── response_cache.go       # Cache middleware
│   ├── response_cache_test.go  # Cache middleware tests
│   └── (35+ tests)
│
└── routing/          # Route definitions
    ├── router.go     # Router setup, routes configuration
    └── (no direct tests - use integration tests)
```

### Key Files

| File | Purpose | Tests |
|------|---------|-------|
| `main.go` | Application entry point | Integration tests |
| `pkg/app/app.go` | Application lifecycle | 61+ tests (all pools) |
| `pkg/cache/*.go` | Caching layer | 26+ tests |
| `pkg/middleware/*.go` | Middleware stack | 35+ tests |
| `pkg/routing/router.go` | Route configuration | 8 integration tests |
| `*_test.go` | Test files | Various |

---

## Adding New Features

### Example: Add New Middleware

1. **Create middleware** in `pkg/middleware/new_middleware.go`:

```go
package middleware

type NewMiddleware struct {
    // Configuration
}

func NewNewMiddleware() *NewMiddleware {
    return &NewMiddleware{}
}

func (nm *NewMiddleware) Middleware() Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Middleware logic
            next.ServeHTTP(w, r)
        })
    }
}
```

2. **Write tests** in `pkg/middleware/new_middleware_test.go`:

```go
package middleware

import (
    "testing"
    "net/http"
    "net/http/httptest"
)

func TestNewMiddleware(t *testing.T) {
    nm := NewNewMiddleware()
    
    // Create test handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    // Apply middleware
    wrapped := nm.Middleware()(handler)
    
    // Test
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    wrapped.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

3. **Apply in router** in `pkg/routing/router.go`:

```go
func (r *Router) SetupRoutes() {
    r.mux.Use(NewNewMiddleware().Middleware())
    // ... rest of routes
}
```

4. **Run tests**:

```bash
go test ./pkg/middleware/... -v
go test ./...
```

### Example: Add Cache Pattern

1. **Register in router** `pkg/routing/router.go`:

```go
func (r *Router) registerCacheInvalidationPatterns() {
    r.cacheInvalidation.RegisterPattern("/api/v1/myservice", "my_table")
}
```

2. **Add corresponding handler** in `pkg/handlers/handlers.go`:

```go
func (h *Handlers) MyServiceMutation(w http.ResponseWriter, r *http.Request) {
    // Business logic
    // Cache is automatically invalidated by middleware
}
```

3. **Test** via integration tests:

```bash
go test -v phase4d_integration_test.go main.go
```

---

## Code Conventions

### Naming
- **Packages**: lowercase, short (`cache`, `middleware`, `routing`)
- **Structs**: PascalCase (`ResponseCache`, `ServiceContainer`)
- **Functions**: PascalCase for exports, camelCase for private
- **Interfaces**: PascalCase with suffix "er" if appropriate (`Middleware`)
- **Constants**: UPPER_SNAKE_CASE for constants

### Error Handling
```go
// Use custom error type
err := errors.New(
    errors.CodeValidation,
    "descriptive message",
    errors.SeverityWarning,
)

// Chain errors with context
return errors.InitializationError("cache", err)
```

### Testing
```go
// Test function naming
func TestFeatureName(t *testing.T) { }
func BenchmarkFeatureName(b *testing.B) { }

// Table-driven tests
tests := []struct {
    name     string
    input    string
    expected bool
}{
    {"case 1", "input", true},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := fn(tt.input)
        if result != tt.expected {
            t.Error("failed")
        }
    })
}
```

### Comments
- Exported symbols must have comments
- Complex logic needs explanation
- TODO comments for future work
- Reference design decisions

```go
// ResponseCache wraps HTTP response caching
// with TTL support and statistics tracking.
type ResponseCache struct {
    cache Cache  // Underlying storage
    mu    sync.RWMutex
}

// CachedResponse represents a cached HTTP response
// with metadata for cache management.
type CachedResponse struct {
    StatusCode int          // HTTP status
    Body       []byte       // Response body
    ExpiresAt  time.Time    // Cache expiration
}
```

### Imports
```go
// Group imports: standard library, external, local
import (
    // Standard library
    "context"
    "net/http"
    "sync"
    
    // External packages
    "github.com/gorilla/mux"
    
    // Local packages
    "qr-menu/pkg/cache"
    "qr-menu/pkg/config"
)
```

---

## Best Practices

### Performance
- **Use caching**: Enable `CACHE_ENABLED=true` during development
- **Monitor**: Check `/api/admin/cache/stats` regularly
- **Test**: Run benchmarks: `go test -bench=. ./...`
- **Profile**: Use `pprof` for performance analysis

### Security
- **Validate input**: Always validate user input
- **Error messages**: Don't expose internal details in errors
- **Logging**: Don't log sensitive data
- **Dependencies**: Keep Go modules updated

### Reliability
- **Error handling**: Always handle errors explicitly
- **Graceful shutdown**: Use context for cancelation
- **Concurrency**: Use mutexes for shared state
- **Testing**: Write tests for critical paths

### Maintainability
- **Small functions**: Keep functions focused and small
- **Documentation**: Comment exported symbols
- **Consistency**: Follow project conventions
- **Testing**: Aim for > 80% coverage on critical code

---

## Common Tasks

### Add a New Package

```bash
# Create directory
mkdir -p pkg/mypackage

# Create main file
touch pkg/mypackage/mypackage.go

# Create test file
touch pkg/mypackage/mypackage_test.go

# Add to imports and initialization
```

### Update Dependencies

```bash
# Add dependency
go get github.com/user/package

# Update dependency
go get -u github.com/user/package

# Cleanup
go mod tidy

# Verify
go test ./...
```

### Debug Application

```bash
# Enable debug logging
LOGGER_LEVEL=debug go run main.go

# Listen on different port
PORT=3000 go run main.go

# Run with race detector (finds concurrency issues)
go run -race main.go

# Print environment
env | grep -E "(CACHE|SERVER|DATABASE)"
```

### Benchmark Performance

```bash
# Run cache benchmarks
go test -bench=. -benchmem ./pkg/cache/...

# Run middleware benchmarks
go test -bench=. -benchmem ./pkg/middleware/...

# Compare benchmarks
go test -bench=. -benchmem ./... -benchtime=10s
```

### Check Code Quality

```bash
# Format all code
go fmt ./...

# Check for obvious errors
go vet ./...

# Check for unused dependencies
go mod tidy

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Publishing Changes

### Git Workflow

```bash
# Create branch for feature
git checkout -b feature/description

# Make changes and test
go test ./...

# Stage changes
git add .

# Commit with clear message
git commit -m "Feature: Add new middleware"

# Push
git push origin feature/description

# Create pull request
# ... review and merge ...

# Sync local
git checkout main
git pull origin main
```

### Commit Messages

```
# Good
commit 1d3b5c7
Author: User <email>

Feature: Add response caching middleware

- Implements ResponseCachingMiddleware
- Supports TTL-based expiration
- Adds X-Cache header (HIT/MISS)
- Includes 10 unit tests
- Performance: 100x improvement on cache hits

# Bad
commit 2f8a3e9
Author: User <email>

Fixed stuff

Changes: Lots of stuff
```

---

## Troubleshooting Development

### Build Fails
```bash
# Clean build cache
go clean -cache

# Download missing dependencies
go mod download

# Verify modules
go mod verify

# Rebuild
go build -v ./...
```

### Tests Fail
```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestName -v ./pkg/...

# Check for race conditions
go test -race ./...

# Update test files
# ... fix code or test assertions ...
```

### Memory Issues
```bash
# Clear go cache
go clean -cache

# Reduce cache sizes during development
CACHE_MAX_RESPONSE_SIZE=100 go run main.go

# Monitor memory usage
top && go run main.go
```

---

## Resources

### Documentation
- [Go Documentation](https://golang.org/doc/)
- [Gorilla Mux](https://github.com/gorilla/mux)
- [PostgreSQL Driver](https://github.com/lib/pq)

### Tools
- [Visual Studio Code](https://code.visualstudio.com/) + Go extension
- [GoLand IDE](https://www.jetbrains.com/go/)
- [Go Tools](https://golang.org/doc/cmd)

### Learning
- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Concurrency in Go](https://go.dev/blog/pipelines)

---

## Questions & Support

- **Code Issues**: Check test files for similar patterns
- **Architecture**: Read [ARCHITECTURE.md](ARCHITECTURE.md)
- **Deployment**: Refer to [DEPLOYMENT.md](DEPLOYMENT.md)
- **Tests**: Run `go test ./... -v`

---

**Developer Guide v2.0**  
**Last Updated**: February 24, 2026  
**Status**: ✅ Complete
