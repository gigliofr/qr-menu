# QR-Menu Monitoring with Prometheus & Grafana

## Overview

This directory contains configuration for monitoring qr-menu using Prometheus (metrics collection) and Grafana (visualization).

## Components

- **Prometheus**: Time-series database for metrics collection from `/metrics` endpoint
- **Grafana**: Visualization and alerting dashboard
- Dashboard templates in JSON format
- Configuration files for auto-provisioning

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- qr-menu app running on `http://localhost:8080`

### Start Monitoring Stack

```bash
# Start Prometheus and Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Check services
docker ps

# Wait for services to be healthy (30-60s)
docker-compose -f docker-compose.monitoring.yml ps
```

### Access Interfaces

- **Grafana Dashboard**: http://localhost:3000 (admin/admin)
- **Prometheus UI**: http://localhost:9090
- **Metrics Endpoint**: http://localhost:8080/metrics

### Import Dashboard

1. Open Grafana (http://localhost:3000)
2. Login with admin/admin
3. Go to **Dashboards** → **Import**
4. Upload `config/grafana-dashboard.json`
5. Select Prometheus as data source

Or use API:
```bash
curl -X POST http://localhost:3000/api/dashboards/db \
  -H "Authorization: Bearer $(curl -s -X POST http://localhost:3000/api/auth/login \
    -d '{"user":"admin","password":"admin"}' | jq -r '.token')" \
  -d @config/grafana-dashboard.json
```

## Architecture

```
qr-menu App
    ↓ (exports metrics)
  /metrics endpoint
    ↓
  Prometheus (scrapes every 10s)
    ↓
  Time-series DB
    ↓
  Grafana (queries & visualizes)
    ↓
  Browser Dashboard
```

## Metrics Collected

### HTTP Metrics

- `qr_menu_http_requests_total` - Total requests by method, path, status
- `qr_menu_http_request_duration_seconds` - Request latency histogram
  - Buckets: [.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10] seconds

### Derived Metrics (PromQL)

- **Request Rate**: `rate(qr_menu_http_requests_total[5m])`
- **P95 Latency**: `histogram_quantile(0.95, rate(qr_menu_http_request_duration_seconds_bucket[5m]))`
- **Error Rate**: `rate(qr_menu_http_requests_total{status=~"5.."}[5m])`
- **Availability**: `100 - (5xx_errors / total_requests) * 100`

## Dashboard Panels

### 1. HTTP Request Rate (5m)
- Line chart showing requests per second
- Grouped by method and path

### 2. Request Latency (P95/P99)
- Latency percentiles over time
- Helps identify slowdowns

### 3. Error Rate
- 4xx and 5xx error trends
- Identifies problematic endpoints

### 4. Service Availability
- Gauge showing uptime percentage
- Threshold-based coloring (red < 50%, yellow < 80%, green ≥ 80%)

### 5. Top Endpoints by Traffic
- Stacked bar chart of request distribution
- Shows which endpoints get most traffic

## Configuration Files

### prometheus.yml
- Scrape targets and intervals
- Job configurations
- Alert rule files (optional)

### Grafana Provisioning
Place datasources and dashboards in:
```
config/grafana-provisioning/
├── dashboards/
│   └── qr-menu.json
└── datasources/
    └── prometheus.yml
```

## Alert Setup (Optional)

Create `config/alerts.yml`:
```yaml
groups:
  - name: qrmenu_alerts
    interval: 5m
    rules:
      - alert: HighErrorRate
        expr: rate(qr_menu_http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        annotations:
          summary: "High error rate detected"
```

Add to `prometheus.yml`:
```yaml
rule_files:
  - "alerts.yml"
```

## Troubleshooting

### Prometheus not scraping metrics

```bash
# Check Prometheus logs
docker-compose -f docker-compose.monitoring.yml logs prometheus

# Verify metrics endpoint
curl http://localhost:8080/metrics

# Check targets in Prometheus UI
http://localhost:9090/targets
```

### Grafana not connecting to Prometheus

```bash
# Check Grafana logs
docker-compose -f docker-compose.monitoring.yml logs grafana

# Verify Prometheus is accessible from Grafana container
docker exec qr-menu-grafana curl http://prometheus:9090
```

### Reset Grafana password

```bash
docker-compose -f docker-compose.monitoring.yml exec grafana \
  grafana-cli admin reset-admin-password newpassword
```

### High memory usage

Adjust Prometheus retention:
```bash
# Edit docker-compose.monitoring.yml, add to Prometheus command:
- '--storage.tsdb.retention.time=7d'  # Keep only 7 days
```

## Production Deployment

### Managed Services

For Railway/AWS/GCP, use managed monitoring:
- **AWS**: CloudWatch (built-in)
- **Google Cloud**: Cloud Monitoring
- **Datadog**: External SaaS
- **New Relic**: External SaaS

### Self-Hosted Setup

1. Deploy Prometheus on separate VM
2. Use external volume storage for persistence
3. Configure authentication (reverse proxy)
4. Set up alert routing
5. Configure backup strategy

### Configuration Updates

```bash
# Reload Prometheus config without restart
curl -X POST http://localhost:9090/-/reload
```

## References

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Dashboards](https://grafana.com/grafana/dashboards/)
- [PromQL Query Syntax](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [Go Prometheus Client](https://pkg.go.dev/github.com/prometheus/client_golang)

## Support

For issues with qr-menu metrics, check:
1. `/metrics` endpoint returns valid Prometheus format
2. Middleware is properly registered in router
3. Port 8080 is accessible from Prometheus container
4. Check app logs for metric recording errors

