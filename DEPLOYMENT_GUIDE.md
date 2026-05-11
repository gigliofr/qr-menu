# QR Menu System - Deployment Guide

## Overview

The QR Menu System supports two deployment scenarios:
1. **Development Mode** (file-storage fallback): Runs without MongoDB, using local JSON files from `storage/` directory
2. **Production Mode** (MongoDB): Requires MongoDB connection string for data persistence

---

## Development Mode (File-Storage Fallback)

### When to Use
- Local development and testing
- Quick prototyping without infrastructure
- Offline/isolated environments

### Setup

```bash
# 1. No environment configuration needed
# MONGODB_URI is NOT required

# 2. Run the server
go run .

# 3. Server will bind to default port 8080 (or PORT env var if set)
# Output: "QR Menu System ready"
```

### How It Works
- `db/mongo.go` has fallback logic: if `MONGODB_URI` is not set, it reads from local `storage/` JSON files
- Supported file patterns:
  - `storage/menu_*.json` (menu data)
  - `storage/restaurant_*.json` (restaurant data)
  - `storage/session_*.json` (session data - if needed)

### Example Data Files
```
storage/
  menu_ba07044b-cd7b-4dca-8890-890b2eaa1aa3.json    # Menu by ID
  restaurant_6b23f20c-d3f9-42f6-8185-0c987bb2261d.json  # Restaurant by ID
```

### Public Menu Route (File-Storage)
```
GET /menu/{menuID}

Example: http://localhost:8080/menu/ba07044b-cd7b-4dca-8890-890b2eaa1aa3

Loads menu_ba07044b-cd7b-4dca-8890-890b2eaa1aa3.json from storage/
```

### Testing
```bash
# Test file-storage fallback is working
curl http://localhost:8080/menu/ba07044b-cd7b-4dca-8890-890b2eaa1aa3

# Run accessibility checks
npx pa11y http://localhost:8080/menu/ba07044b-cd7b-4dca-8890-890b2eaa1aa3
```

### Limitations
- No persistent writes (all data is read-only from files)
- Single-process only (no clustering)
- No real-time data sync
- Good for demo/prototype, **NOT suitable for production**

---

## Production Mode (MongoDB)

### Prerequisites
- MongoDB Atlas cluster or self-hosted MongoDB instance
- MongoDB connection string (URI format)
- Network access to MongoDB from deployment environment

### Environment Variables

```bash
# Required
export MONGODB_URI="mongodb+srv://user:password@cluster.mongodb.net/qr-menu?retryWrites=true&w=majority"

# Optional
export PORT=8080                    # Default: 8080
export NODE_ENV=production          # For app-level optimizations
export LOG_LEVEL=info               # Default: info
```

### Setup

#### Local Testing with MongoDB
```bash
# 1. Set MongoDB connection string
export MONGODB_URI="mongodb+srv://user:password@cluster.mongodb.net/qr-menu"

# 2. Run the server
go run .

# 3. Server will connect to MongoDB and bind to port 8080
```

#### Docker Deployment (with Docker Compose)
```bash
# 1. Build image
docker build -t qr-menu:latest .

# 2. Run with MongoDB
docker run -d \
  -e MONGODB_URI="mongodb+srv://user:password@cluster.mongodb.net/qr-menu" \
  -e PORT=8080 \
  -p 8080:8080 \
  qr-menu:latest
```

#### Railway Deployment
1. Connect your GitHub repository to Railway
2. Set environment variables in Railway dashboard:
   - `MONGODB_URI` → Your MongoDB Atlas URI
   - `NODE_ENV` → `production`
   - `PORT` → `8080` (Railway sets this automatically)
3. Deploy via Git push or Railway CLI

### Public Menu Route (MongoDB)
```
GET /menu/{menuID}

Example: http://your-domain.com/menu/ba07044b-cd7b-4dca-8890-890b2eaa1aa3

Fetches menu document from MongoDB collection "menus"
```

### Monitoring & Verification
```bash
# Health check
curl http://your-domain.com/health

# Public menu test
curl http://your-domain.com/menu/{menuID}

# Accessibility check (post-deployment)
npx pa11y http://your-domain.com/menu/{menuID}
```

---

## Pre-Deployment Checklist

### Code Quality
- [ ] All tests pass: `go test ./...`
- [ ] pa11y checks pass (WCAG AA compliance)
  ```bash
  npx pa11y http://localhost:8080/templates/login.html
  npx pa11y http://localhost:8080/templates/guest_access.html
  npx pa11y http://localhost:8080/menu/{sampleMenuID}
  ```
- [ ] No uncommitted changes: `git status`
- [ ] Latest changes pushed to main: `git log -1`

### Environment-Specific

#### For Production (MongoDB)
- [ ] MongoDB Atlas cluster configured and accessible
- [ ] Database connection string tested (can connect from deployment environment)
- [ ] Database backups enabled
- [ ] Admin user credentials secured
- [ ] Network security groups allow MongoDB access only from app IPs

#### For Development (File-Storage)
- [ ] Sample JSON files exist in `storage/` directory
- [ ] JSON files are valid (can parse with `jq` or similar)
- [ ] File permissions allow read access

### Deployment Infrastructure
- [ ] Docker image builds successfully: `docker build -t qr-menu:latest .`
- [ ] Port binding configured (default 8080)
- [ ] TLS/HTTPS configured (production)
- [ ] CORS headers configured for frontend domain
- [ ] Rate limiting enabled (security middleware)
- [ ] Logging configured (stdout/stderr or file rotation)

### Performance & Security
- [ ] Security headers verified (`X-Content-Type-Options`, `X-Frame-Options`, etc.)
- [ ] GDPR compliance checked (`security/gdpr.go`)
- [ ] Audit logging enabled (`handlers/audit_helper.go`)
- [ ] Rate limiting active (`security/ratelimit.go`)
- [ ] Encryption enabled where needed (`security/encryption.go`)

### Post-Deployment
- [ ] Server starts without errors
- [ ] Public menu route returns valid HTML
- [ ] JSON endpoints return proper MIME types
- [ ] Error pages are user-friendly (no stack traces in production)
- [ ] Monitoring/alerts configured
- [ ] Rollback procedure documented and tested

---

## Environment Decision Matrix

| Aspect | File-Storage (Dev) | MongoDB (Prod) |
|--------|-------------------|----------------|
| **Setup Time** | < 1 min | 5-10 min |
| **Data Persistence** | None (read-only) | Full CRUD |
| **Scaling** | Single process | Clustered |
| **Backup Strategy** | Manual JSON export | MongoDB backups |
| **Cost** | Free (local) | ~$57/month (MongoDB Atlas) |
| **Use Case** | Demo, prototyping | Production, staging |
| **Failover** | N/A | Automatic (Atlas) |
| **Security** | Filesystem ACLs | Network security + auth |

---

## Switching Between Modes

### Dev → Production
```bash
# 1. Remove old env vars
unset MONGODB_URI  # Clear if needed

# 2. Set production MongoDB URI
export MONGODB_URI="mongodb+srv://user:pass@prod-cluster.mongodb.net/qr-menu"

# 3. Run server (will detect MONGODB_URI and use MongoDB)
go run .

# 4. Verify connection in logs
```

### Production → Dev (Rollback)
```bash
# 1. Unset MongoDB URI
unset MONGODB_URI

# 2. Run server (will use file-storage fallback)
go run .

# 3. Server will load from local storage/ directory
```

---

## Troubleshooting

### Server won't start
```bash
# Check logs
go run . 2>&1 | head -20

# Verify port is not in use
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows
```

### MongoDB connection fails
```bash
# Test connection string
mongosh "mongodb+srv://user:pass@cluster.mongodb.net/qr-menu"

# Check network access from deployment environment
# Verify IP whitelist in MongoDB Atlas
```

### Menu page returns 404
```bash
# Check menu ID exists
ls -la storage/menu_*.json  # Dev mode
db.menus.findOne({_id: ObjectId("...")})  # MongoDB mode

# Verify handlers.go PublicMenuHandler is mounted
grep -n "PublicMenuHandler" handlers/handlers.go
```

### Accessibility test fails
```bash
# Re-run pa11y with verbose output
npx pa11y http://localhost:8080/menu/{menuID} --runner axe

# Check design-system.css color variables
grep --color=always "color-text\|color-bg" static/css/design-system.css
```

---

## Recommended Deployment Flow

1. **Local Development**
   - File-storage mode, iterate quickly
   - Run pa11y before each commit

2. **Testing/Staging**
   - MongoDB staging instance
   - Full test suite, load testing

3. **Production**
   - MongoDB Atlas production cluster
   - Managed backups, monitoring
   - CDN for static assets
   - Rate limiting + security headers

---

## References

- [MongoDB Connection String](https://docs.mongodb.com/manual/reference/connection-string/)
- [Railway Deployment](./RAILWAY_SETUP_GUIDE.md)
- [Docker Setup](./Dockerfile)
- [Accessibility Testing](https://www.pa11y.org/)
- [Security Headers](./security/headers.go)
