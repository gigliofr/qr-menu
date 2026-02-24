# Deployment & Configuration Guide

**QR Menu System v2.0.0**  
**Production Deployment Instructions**

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Environment Configuration](#environment-configuration)
3. [Deployment Scenarios](#deployment-scenarios)
4. [Monitoring & Health](#monitoring--health)
5. [Troubleshooting](#troubleshooting)

---

## Quick Start

### Development Environment

```bash
# Build
go build -o qr-menu .

# Run (default settings)
./qr-menu

# Access points
# Admin: http://localhost:8080/admin
# API: http://localhost:8080/api/v1
# Health: http://localhost:8080/health
```

### Production Environment

```bash
# Configure for production
export ENVIRONMENT=prod
export CACHE_ENABLED=true
export CACHE_RESPONSE_TTL=30m
export CACHE_QUERY_TTL=60m
export SECURITY_ENABLE_HTTPS=true
export SECURITY_CERT_FILE=/path/to/cert.pem
export SECURITY_KEY_FILE=/path/to/key.pem

# Build
go build -o qr-menu-prod .

# Run
./qr-menu-prod
```

### Docker Deployment

```dockerfile
FROM golang:1.24-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o qr-menu .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/qr-menu .
COPY --from=builder /app/static ./static
EXPOSE 8080
CMD ["./qr-menu"]
```

```bash
# Build Docker image
docker build -t qr-menu:2.0.0 .

# Run container
docker run -p 8080:8080 \
  -e ENVIRONMENT=prod \
  -e CACHE_ENABLED=true \
  -e DATABASE_DSN='postgres://user:pass@db:5432/qrmenu' \
  qr-menu:2.0.0
```

---

## Environment Configuration

### Core Settings

#### Server Configuration
```bash
# Port (default: 8080)
SERVER_PORT=8080

# Host (default: localhost)
SERVER_HOST=0.0.0.0    # For container/remote access
SERVER_HOST=localhost  # For local development

# Environment (default: dev)
ENVIRONMENT=prod       # prod|staging|dev

# Max request body size (default: 10MB)
SERVER_MAX_BODY_SIZE=10485760

# Timeouts
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_IDLE_TIMEOUT=120s
```

#### Database Configuration
```bash
# Database connection
# PostgreSQL example:
DATABASE_DSN=postgres://user:password@localhost:5432/qrmenu?sslmode=disable
DATABASE_ENGINE=postgres

# Connection pool
DATABASE_MAX_OPEN_CONNS=25
DATABASE_MAX_IDLE_CONNS=5
DATABASE_CONN_MAX_LIFETIME=5m
DATABASE_CONN_MAX_IDLE_TIME=10m

# Migrations
DATABASE_MIGRATION_PATH=./migrations
DATABASE_AUTO_MIGRATE=true
```

#### Cache Configuration (CRITICAL FOR PERFORMANCE)
```bash
# Enable/disable (default: true)
CACHE_ENABLED=true

# Response cache TTL
CACHE_RESPONSE_TTL=5m      # Development
CACHE_RESPONSE_TTL=30m     # Production recommended

# Query cache TTL
CACHE_QUERY_TTL=10m        # Development
CACHE_QUERY_TTL=60m        # Production recommended

# Cache size limits
CACHE_MAX_RESPONSE_SIZE=1000    # Max responses to cache
CACHE_MAX_QUERY_SIZE=500        # Max query results to cache

# Auto-invalidation (default: true)
CACHE_INVALIDATE_ON_MUTATION=true
```

#### Security Configuration
```bash
# HTTPS
SECURITY_ENABLE_HTTPS=false
SECURITY_CERT_FILE=/etc/letsencrypt/live/domain/cert.pem
SECURITY_KEY_FILE=/etc/letsencrypt/live/domain/privkey.pem

# CORS
SECURITY_CORS_ENABLED=true
SECURITY_CORS_ALLOWED_ORIGINS=https://example.com,https://app.example.com

# Session & Auth
SECURITY_SESSION_TIMEOUT=24h
SECURITY_PASSWORD_MIN_LEN=8
SECURITY_PASSWORD_REQUIRE_SPECIAL=true
SECURITY_PASSWORD_REQUIRE_NUMBERS=true

# Rate Limiting
SECURITY_RATE_LIMIT_PER_SEC=10
SECURITY_RATE_LIMIT_BURST=100
```

#### Logging Configuration
```bash
# Log level (debug|info|warn|error|fatal)
LOGGER_LEVEL=info          # Development: debug
LOGGER_LEVEL=warn          # Production: warn

# Log format (json|text)
LOGGER_FORMAT=json
LOGGER_OUTPUT_FILE=./logs/qr-menu.log

# Log rotation
LOGGER_MAX_SIZE=100        # MB
LOGGER_MAX_BACKUPS=10      # number of files
LOGGER_MAX_AGE=30          # days
LOGGER_COMPRESS=true
```

#### Analytics Configuration
```bash
# Enable/disable
ANALYTICS_ENABLED=true
ANALYTICS_TRACKING_ENABLED=true

# Storage
ANALYTICS_STORAGE_PATH=./analytics

# Data retention
ANALYTICS_RETENTION_DAYS=90
ANALYTICS_CLEANUP_INTERVAL=24h
```

#### Backup Configuration
```bash
# Enable/disable
BACKUP_ENABLED=true

# Storage
BACKUP_STORAGE_PATH=./backups
BACKUP_MAX_BACKUPS=30

# Scheduling
BACKUP_SCHEDULE_TIME=02:00  # 2 AM daily

# Compression
BACKUP_COMPRESSION_LEVEL=6  # 1-9

# Retention
BACKUP_RETENTION_DAYS=90
BACKUP_ROTATION_INTERVAL=24h
```

#### Notifications Configuration
```bash
# Enable/disable
NOTIFICATIONS_ENABLED=true

# Queue
NOTIFICATIONS_QUEUE_SIZE=100
NOTIFICATIONS_WORKERS=3
NOTIFICATIONS_BATCH_SIZE=10
NOTIFICATIONS_BATCH_TIMEOUT=5s

# Retry
NOTIFICATIONS_MAX_RETRIES=3
NOTIFICATIONS_RETRY_DELAY=10s

# Firebase Cloud Messaging
NOTIFICATIONS_FCM_CREDENTIALS_URL=/path/to/firebase.json
```

#### Localization Configuration
```bash
# Default language
LOCALIZATION_DEFAULT_LANG=it

# Timezone
LOCALIZATION_TIMEZONE_OFFSET=1  # Hours from UTC

# Date/Time formats
LOCALIZATION_DATE_FORMAT=2006-01-02      # Go format
LOCALIZATION_TIME_FORMAT=15:04:05        # Go format
```

---

## Deployment Scenarios

### Scenario 1: Local Development

```bash
# .env file
ENVIRONMENT=dev
SERVER_PORT=8080
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=1m
CACHE_QUERY_TTL=5m
LOGGER_LEVEL=debug
DATABASE_ENGINE=postgres
# DATABASE_DSN optional - most features work without DB
```

### Scenario 2: Docker Development

```bash
# docker-compose.yml
version: '3.8'
services:
  qr-menu:
    build: .
    ports:
      - "8080:8080"
    environment:
      ENVIRONMENT: dev
      SERVER_HOST: 0.0.0.0
      SERVER_PORT: 8080
      CACHE_ENABLED: "true"
      DATABASE_DSN: postgres://dev:dev@postgres:5432/qrmenu
      LOGGER_LEVEL: debug
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
      POSTGRES_DB: qrmenu
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
# Run
docker-compose up
```

### Scenario 3: Production Deployment

```bash
# .env.production
ENVIRONMENT=prod
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Cache (for performance)
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=30m
CACHE_QUERY_TTL=60m
CACHE_MAX_RESPONSE_SIZE=5000
CACHE_MAX_QUERY_SIZE=2000

# Security
SECURITY_ENABLE_HTTPS=true
SECURITY_CERT_FILE=/etc/letsencrypt/live/menu.example.com/fullchain.pem
SECURITY_KEY_FILE=/etc/letsencrypt/live/menu.example.com/privkey.pem
SECURITY_CORS_ENABLED=true
SECURITY_CORS_ALLOWED_ORIGINS=https://menu.example.com

# Database (PostgreSQL required)
DATABASE_DSN=postgres://dbuser:${DB_PASSWORD}@db.example.com:5432/qrmenu
DATABASE_ENGINE=postgres
DATABASE_MAX_OPEN_CONNS=50
DATABASE_MAX_IDLE_CONNS=10

# Logging
LOGGER_LEVEL=warn
LOGGER_OUTPUT_FILE=/var/log/qr-menu/qr-menu.log
LOGGER_MAX_SIZE=200
LOGGER_MAX_BACKUPS=20

# Backup
BACKUP_ENABLED=true
BACKUP_STORAGE_PATH=/var/backups/qr-menu
BACKUP_MAX_BACKUPS=30

# Notifications
NOTIFICATIONS_ENABLED=true
NOTIFICATIONS_FCM_CREDENTIALS_URL=/etc/qr-menu/firebase.json
```

```bash
# Build and deploy
go build -o qr-menu-prod .

# Run with systemd
sudo cp qr-menu-prod /usr/local/bin/

# Create systemd service
sudo tee /etc/systemd/system/qr-menu.service << EOF
[Unit]
Description=QR Menu System
After=network.target

[Service]
Type=simple
User=qrmenu
WorkingDirectory=/opt/qr-menu
EnvironmentFile=/etc/qr-menu/.env.production
ExecStart=/usr/local/bin/qr-menu-prod
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable qr-menu
sudo systemctl start qr-menu
```

### Scenario 4: Kubernetes Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: qr-menu
spec:
  replicas: 3
  selector:
    matchLabels:
      app: qr-menu
  template:
    metadata:
      labels:
        app: qr-menu
    spec:
      containers:
      - name: qr-menu
        image: qr-menu:2.0.0
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "prod"
        - name: CACHE_ENABLED
          value: "true"
        - name: CACHE_RESPONSE_TTL
          value: "30m"
        - name: DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: qr-menu-secrets
              key: database-dsn
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: qr-menu-service
spec:
  selector:
    app: qr-menu
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

```bash
# Deploy
kubectl apply -f deployment.yaml

# Monitor
kubectl logs -f deployment/qr-menu
kubectl describe pod <pod-name>
```

---

## Monitoring & Health

### Health Check Endpoints

```bash
# Application health
curl http://localhost:8080/health
# Returns: {"status":"ok", "timestamp":"..."}

# Readiness check
curl http://localhost:8080/ready
# Returns: {"status":"ok", "services":{...}}

# Application status
curl http://localhost:8080/status
# Returns: Full status with all services
```

### Cache Monitoring

```bash
# Cache statistics
curl http://localhost:8080/api/admin/cache/stats
# Returns:
# {
#   "response_cache": {
#     "hits": 150,
#     "misses": 50,
#     "hit_rate": "75.00%",
#     "size": 145
#   },
#   "query_cache": {...}
# }

# Cache status
curl http://localhost:8080/api/admin/cache/status
# Returns:
# {
#   "enabled": true,
#   "response_cache_size": 145,
#   "response_cache_hits": 150,
#   ...
# }

# Clear cache
curl -X POST http://localhost:8080/api/admin/cache/clear
```

### Performance Monitoring

```bash
# Check response time (with caching)
# First request: ~10-15ms (miss)
# Subsequent: ~1-10µs (hit) = 100-1,000x faster

# Monitor database queries
# Uncached: 50-100ms
# Cached: 1-10µs = 5,000-100,000x faster

# Target metrics
# Cache hit rate: > 70%
# Response time: < 50ms (with cache enabled)
# Database load: 99%+ reduction
```

---

## Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8080          # Linux/Mac
netstat -ano | grep 8080  # Windows

# Use different port
PORT=3000 go run main.go

# Or kill existing process
kill -9 <PID>
pkill -f qr-menu
```

### Database Connection Issues

```bash
# Check CONNECTION string
echo $DATABASE_DSN

# Test PostgreSQL connection
psql "postgresql://user:pass@localhost:5432/dbname"

# Common DSN formats:
# PostgreSQL: postgresql://user:password@localhost:5432/dbname
# PostgreSQL with SSL: postgresql://user:password@localhost:5432/dbname?sslmode=require

# Run without database (cache will work)
unset DATABASE_DSN
go run main.go
```

### Cache Issues

```bash
# Clear all caches
curl -X POST http://localhost:8080/api/admin/cache/clear

# Check cache status
curl http://localhost:8080/api/admin/cache/status

# Disable caching (if problematic)
CACHE_ENABLED=false go run main.go

# Check cache statistics
curl http://localhost:8080/api/admin/cache/stats
```

### High Memory Usage

```bash
# Reduce cache sizes
CACHE_MAX_RESPONSE_SIZE=500 ./qr-menu
CACHE_MAX_QUERY_SIZE=250 ./qr-menu

# Reduce TTLs
CACHE_RESPONSE_TTL=1m ./qr-menu
CACHE_QUERY_TTL=5m ./qr-menu

# Or disable caching entirely
CACHE_ENABLED=false ./qr-menu
```

### Slow Performance

```bash
# Enable caching (huge improvement)
CACHE_ENABLED=true go run main.go

# Increase cache TTL for stable data
CACHE_RESPONSE_TTL=30m ./qr-menu

# Increase cache size
CACHE_MAX_RESPONSE_SIZE=5000 ./qr-menu

# Check cache statistics
curl http://localhost:8080/api/admin/cache/stats
# Look for hit_rate > 70%
```

### CORS Issues

```bash
# Enable CORS
SECURITY_CORS_ENABLED=true ./qr-menu

# Configure allowed origins
SECURITY_CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080 ./qr-menu
```

### Rate Limiting

```bash
# Increase limits (if needed)
SECURITY_RATE_LIMIT_PER_SEC=100 ./qr-menu
SECURITY_RATE_LIMIT_BURST=1000 ./qr-menu

# Or disable (not recommended for production)
SECURITY_RATE_LIMIT_PER_SEC=999999 ./qr-menu
```

---

## Performance Tuning

### For Development
```bash
# Fast startup, aggressive caching (short TTL for testing)
ENVIRONMENT=dev
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=10s
CACHE_QUERY_TTL=30s
LOGGER_LEVEL=debug
```

### For Small Deployments
```bash
# Balanced cache settings
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=5m
CACHE_QUERY_TTL=15m
CACHE_MAX_RESPONSE_SIZE=500
CACHE_MAX_QUERY_SIZE=250
```

### For Large Deployments
```bash
# Maximize performance with larger caches
CACHE_ENABLED=true
CACHE_RESPONSE_TTL=30m
CACHE_QUERY_TTL=60m
CACHE_MAX_RESPONSE_SIZE=10000
CACHE_MAX_QUERY_SIZE=5000
LOGGER_LEVEL=warn  # Less I/O overhead
DATABASE_MAX_OPEN_CONNS=100
```

---

## Backup & Recovery

### Automatic Backups
```bash
# Enabled by default, daily at 2 AM
BACKUP_ENABLED=true
BACKUP_SCHEDULE_TIME=02:00
BACKUP_STORAGE_PATH=./backups
BACKUP_MAX_BACKUPS=30  # Keep last 30 backups
```

### Manual Backup
```bash
# Create backup via API
curl -X POST http://localhost:8080/api/backup/create

# List backups
curl http://localhost:8080/api/backup/list

# Restore backup
curl -X PUT http://localhost:8080/api/backup/{id}
```

---

## Environment Variable Reference

| Variable | Default | Example | Purpose |
|----------|---------|---------|---------|
| `SERVER_PORT` | 8080 | 3000 | HTTP port |
| `ENVIRONMENT` | dev | prod | Environment type |
| `CACHE_ENABLED` | true | false | Enable/disable cache |
| `CACHE_RESPONSE_TTL` | 5m | 30m | Response cache TTL |
| `DATABASE_DSN` | - | postgres://... | DB connection |
| `LOGGER_LEVEL` | info | debug | Log level |
| `SECURITY_ENABLE_HTTPS` | false | true | Enable HTTPS |

See **Environment Configuration** section above for complete reference.

---

## Checklist for Production Deployment

- [ ] Build application: `go build -o qr-menu-prod .`
- [ ] Configure all environment variables
- [ ] Set `ENVIRONMENT=prod`
- [ ] Enable caching: `CACHE_ENABLED=true`
- [ ] Configure database with PostgreSQL
- [ ] Set up SSL/TLS certificates
- [ ] Enable CORS for your domain
- [ ] Configure backup settings
- [ ] Test health endpoints
- [ ] Set up monitoring
- [ ] Configure log rotation
- [ ] Test cache statistics
- [ ] Load test application
- [ ] Check performance metrics
- [ ] Deploy and monitor

---

**Deployment Guide v2.0**  
**Last Updated**: February 24, 2026  
**Status**: ✅ Production Ready
