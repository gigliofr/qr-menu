# QR-Menu Load Balancing with HAProxy

## Overview

This directory contains configuration for horizontal scaling qr-menu using HAProxy as a reverse proxy and load balancer.

## Files

- `haproxy.cfg` - HAProxy configuration file with roundrobin load balancing
- `docker-compose.haproxy.yml` - Docker Compose setup with 3 app instances, MongoDB, Redis, and HAProxy

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- `.env` file with `MONGODB_URI` configured

### Run with Docker Compose

```bash
# Start all services (3 app instances + HAProxy + MongoDB + Redis)
docker-compose -f docker-compose.haproxy.yml up -d

# Check service health
docker ps

# View HAProxy stats dashboard
open http://localhost:8404/stats

# Access application through load balancer
open http://localhost/

# Check app health through load balancer
curl http://localhost/api/v1/health
```

### Stop Services

```bash
docker-compose -f docker-compose.haproxy.yml down
```

## Architecture

```
User Traffic
    ↓
  HAProxy (Port 80/443)
    ↓
    ├─→ App1 (8081)
    ├─→ App2 (8082)
    └─→ App3 (8083)
         ↓
    ┌───┴───┐
    ↓       ↓
 MongoDB  Redis
```

## HAProxy Features

### Load Balancing

- **Algorithm**: Round-robin distribution across 3 instances
- **Health Checks**: Each backend checked every 10 seconds
- **Failover**: Automatically removes unhealthy instances

### Backend Routing

- `/api/v1/health` → Direct to all instances
- `/metrics` → Single instance (for consistency)
- `/cache/*` → All instances with cache check
- `/static/*` → All instances
- `/admin/*` → All instances
- Default → All instances (roundrobin)

### Security

- HTTP → HTTPS redirect
- Rate limiting: 100 requests/second per IP
- TLS 1.2+ enforcement
- Request ID tracking for tracing

### Stats Dashboard

View real-time metrics at `http://localhost:8404/stats`:
- Request counts per backend
- Session statistics
- Error rates
- Server health status

## Configuration

### Scaling

To add more instances:

1. Update `haproxy.cfg` with new server entries:
   ```
   server app4 localhost:8084 check
   server app5 localhost:8085 check
   ```

2. Add services in `docker-compose.haproxy.yml`:
   ```yaml
   app4:
     ...ports: ["8084:8080"]
   ```

### Health Checks

HAProxy performs periodic health checks on `/api/v1/health`. Ensure this endpoint:
- Returns HTTP 200 for healthy instances
- Includes database and cache status
- Is fast (< 1 second)

### SSL Certificates

For production:

1. Place SSL certificate at `./certs/qr-menu.pem`
2. Update HAProxy config:
   ```
   bind 0.0.0.0:443 ssl crt /etc/ssl/certs/qr-menu.pem
   ```

## Monitoring

### Real-time Metrics

- HAProxy Stats: `http://localhost:8404/stats`
- Prometheus Metrics: `http://localhost/metrics`
- Cache Stats: `http://localhost/cache/stats`
- Health Status: `http://localhost/api/v1/health`

### Logs

View HAProxy logs:
```bash
docker-compose -f docker-compose.haproxy.yml logs haproxy -f
```

View app logs:
```bash
docker-compose -f docker-compose.haproxy.yml logs app1 -f
```

## Production Deployment

For Railway/Heroku deployment:

1. Use managed load balancer (built-in)
2. Set `REDIS_URL` to managed Redis instance
3. Set `MONGODB_URI` to Atlas cluster
4. Remove HAProxy layer (not needed)
5. Deploy multiple app instances using platform's scaling

## Troubleshooting

### Instances show as DOWN

```bash
# Check individual instance
curl -i http://localhost:8081/api/v1/health

# Check HAProxy logs
docker-compose -f docker-compose.haproxy.yml logs haproxy
```

### Rate limiting issues

Adjust in `haproxy.cfg`:
```
http-request deny if { sc_http_req_rate(0) gt 100 }
```

### Backend unavailable

Verify MongoDB and Redis are running:
```bash
docker-compose -f docker-compose.haproxy.yml ps
```

## References

- [HAProxy Documentation](http://www.haproxy.org/)
- [Docker Compose Networking](https://docs.docker.com/compose/networking/)
- [QR-Menu Deployment Guide](../DEPLOYMENT_GUIDE.md)
