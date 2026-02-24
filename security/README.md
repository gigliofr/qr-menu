# Security & Compliance Module

## Overview

This module provides comprehensive security and GDPR compliance features for the QR Menu System, including rate limiting, audit logging, GDPR data management, encryption utilities, and security headers.

## Components

### 1. Rate Limiting (`ratelimit.go`)

Advanced token bucket rate limiting with per-user and per-endpoint configuration.

**Features:**
- Token bucket algorithm for smooth rate limiting
- Per-user and per-endpoint limits
- Configurable request rates and burst sizes
- Automatic cleanup of old buckets
- HTTP headers for rate limit status

**Default Configuration:**
- General endpoints: 10 req/s, burst of 20
- Login endpoint: 3 req/s, burst of 5
- Registration: 2 req/s, burst of 3
- Webhooks: 100 req/s, burst of 200

**Usage:**
```go
rateLimiter := security.NewRateLimiter()
defer rateLimiter.Stop()

// Apply as middleware
r.Use(rateLimiter.RateLimitMiddleware)
```

**Response Headers:**
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `Retry-After`: Seconds to wait when rate limited (429 response)

### 2. Audit Logging (`audit.go`)

Comprehensive audit trail for security and compliance requirements.

**Features:**
- Circular buffer for efficient storage
- Event filtering by user, action, time range
- JSON export for compliance reporting
- HTTP request logging middleware
- Specialized logging for auth, data access, modifications, deletions

**Event Types:**
- Authentication events (login, logout, password changes)
- Data access events (GDPR compliance)
- Data modification events (before/after tracking)
- Data deletion events (GDPR right to be forgotten)
- HTTP request logging

**Usage:**
```go
auditLogger := security.NewAuditLogger(10000) // Max 10k events

// Log authentication
auditLogger.LogAuth(userID, "login", true, request, details)

// Log data access (GDPR)
auditLogger.LogDataAccess(userID, "user_profile", "read", details)

// Log data modification
auditLogger.LogDataModification(userID, "restaurant", "update", before, after)

// Apply as middleware
auditMiddleware := security.NewAuditMiddleware(auditLogger)
r.Use(auditMiddleware.Middleware)
```

**Querying Events:**
```go
// By user
events := auditLogger.GetEventsByUser("user123", 100)

// By action
events := auditLogger.GetEventsByAction("delete", 50)

// By time range
events := auditLogger.GetEventsInTimeRange(startTime, endTime, 200)

// Export as JSON
jsonData, _ := auditLogger.ExportJSON()
```

### 3. GDPR Compliance (`gdpr.go`)

Complete GDPR data protection and privacy compliance toolkit.

**Features:**
- Consent management (marketing, analytics, cookies, data sharing)
- Data export (right to data portability)
- Data deletion with grace period (right to be forgotten)
- Data anonymization for analytics
- Audit trail integration

**Consent Management:**
```go
gdprManager := security.NewGDPRManager(auditLogger)

// Record consent
record := security.ConsentRecord{
    UserID:      "user123",
    ConsentType: security.ConsentMarketing,
    Granted:     true,
    IPAddress:   "192.168.1.1",
    UserAgent:   "Mozilla/5.0...",
}
gdprManager.RecordConsent(record)

// Check consent
hasConsent := gdprManager.HasConsent("user123", security.ConsentMarketing)

// Get all consents
consents := gdprManager.GetConsents("user123")
```

**Data Export (Article 15 - Right of Access):**
```go
// Export all user data as JSON
exportData, err := gdprManager.ExportUserDataJSON(
    userID, 
    userData, 
    restaurants, 
    menus, 
    orders, 
    analytics,
)
// Returns complete JSON with all user data, audit logs, and consents
```

**Data Deletion (Article 17 - Right to be Forgotten):**
```go
// Request deletion (30-day grace period)
request, err := gdprManager.RequestDataDeletion(userID, "User requested account closure")

// Cancel deletion (during grace period)
err := gdprManager.CancelDataDeletion(userID)

// Process scheduled deletions (run periodically)
deletedUsers := gdprManager.ProcessScheduledDeletions()
```

**Data Anonymization:**
```go
// Anonymize sensitive fields for analytics
anonymizedData := gdprManager.AnonymizeData(userData)
// Redacts: email, phone, address, name, ip_address
```

### 4. Security Headers (`headers.go`)

Comprehensive HTTP security headers and CORS configuration.

**Security Headers Included:**
- **Content-Security-Policy (CSP)**: Prevents XSS attacks
- **Strict-Transport-Security (HSTS)**: Forces HTTPS
- **X-Frame-Options**: Prevents clickjacking
- **X-Content-Type-Options**: Prevents MIME sniffing
- **X-XSS-Protection**: Legacy XSS protection
- **Referrer-Policy**: Controls referrer information
- **Permissions-Policy**: Controls browser features

**Default CSP Policy:**
```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval' https://js.stripe.com;
style-src 'self' 'unsafe-inline';
img-src 'self' data: https:;
connect-src 'self' https://api.stripe.com;
frame-src https://js.stripe.com;
object-src 'none';
base-uri 'self';
form-action 'self';
frame-ancestors 'none';
upgrade-insecure-requests
```

**Usage:**
```go
// Use default configuration
config := security.DefaultSecurityHeadersConfig()
securityHeaders := security.NewSecurityHeadersMiddleware(config)
r.Use(securityHeaders.Middleware)

// Custom configuration
config := security.SecurityHeadersConfig{
    CSP: "default-src 'self'",
    HSTS: "max-age=31536000; includeSubDomains",
    FrameOptions: "DENY",
    // ... other headers
}
```

**CORS Configuration:**
```go
corsConfig := security.DefaultCORSConfig()
// Allows: localhost:3000, localhost:8080
// Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
// Credentials: true

corsMiddleware := security.NewCORSMiddleware(corsConfig)
r.Use(corsMiddleware.Middleware)
```

### 5. Encryption Utilities (`encryption.go`)

Field-level encryption, password hashing, and token management.

**Features:**
- AES-256-GCM encryption for data at rest
- bcrypt password hashing
- Cryptographically secure token generation
- Field-level encryption helpers
- PBKDF2 key derivation

**Encryption:**
```go
encryption := security.NewEncryption("your-secret-key")

// Encrypt data
encrypted, err := encryption.Encrypt("sensitive data")

// Decrypt data
decrypted, err := encryption.Decrypt(encrypted)
```

**Password Hashing:**
```go
// Hash password (bcrypt with default cost)
hash, err := security.HashPassword("user-password")

// Verify password
isValid := security.CheckPasswordHash("user-password", hash)
```

**Token Generation:**
```go
// Generate cryptographic random token
token, err := security.GenerateRandomToken(32) // 32 bytes = 64 hex chars

// Generate API key
apiKey, err := security.GenerateAPIKey()

// Hash data (SHA-256)
hash := security.HashData([]byte("data to hash"))
```

**Field-Level Encryption:**
```go
fieldEnc := security.NewFieldEncryption("encryption-key")

// Encrypt sensitive fields in a map
data := map[string]interface{}{
    "name": "John",
    "email": "john@example.com",
    "phone": "+1234567890",
}
fieldEnc.EncryptSensitiveFields(data)
// Automatically encrypts: email, phone, address, ssn, credit_card

// Decrypt fields
fieldEnc.DecryptSensitiveFields(data)
```

**Token Manager:**
```go
tokenMgr := security.NewTokenManager()

// Generate token with expiration
token, err := tokenMgr.GenerateToken(userID, "email_verification", expiresAt)

// Validate token
info, valid := tokenMgr.ValidateToken(token)
if valid {
    // Use info.UserID, info.Type
}

// Revoke token
tokenMgr.RevokeToken(token)
```

## API Endpoints

### GDPR Endpoints

All require user authentication.

#### `GET /api/v1/gdpr/my-data`
Export all user data (GDPR Article 15).

**Response:** JSON file download with all user data, audit logs, consents.

#### `POST /api/v1/gdpr/request-deletion`
Request account deletion (GDPR Article 17).

**Body:**
```json
{
  "reason": "User requested account closure"
}
```

**Response:** Deletion request with scheduled date (30-day grace period).

#### `POST /api/v1/gdpr/cancel-deletion`
Cancel pending deletion request.

**Response:** 204 No Content

#### `GET /api/v1/gdpr/deletion-request`
Get deletion request status.

**Response:**
```json
{
  "user_id": "user123",
  "requested_at": "2026-02-24T10:00:00Z",
  "scheduled_at": "2026-03-26T10:00:00Z",
  "status": "scheduled"
}
```

#### `POST /api/v1/gdpr/consent`
Record user consent.

**Body:**
```json
{
  "consent_type": "marketing",
  "granted": true
}
```

**Consent Types:** `marketing`, `analytics`, `data_sharing`, `cookies`

#### `GET /api/v1/gdpr/consents`
Get all user consents.

**Response:** Array of consent records with timestamps.

### Audit Log Endpoints

#### `GET /api/v1/audit/my-logs`
Get current user's audit logs.

**Query Params:**
- `limit`: Max results (default: 100)

**Response:** Array of audit events.

#### `GET /api/v1/audit/logs` (Admin Only)
Get all audit logs.

**Query Params:**
- `limit`: Max results
- `action`: Filter by action type
- `user_id`: Filter by user

#### `GET /api/v1/audit/export` (Admin Only)
Export all audit logs as JSON file.

**Response:** JSON file download.

## Middleware Order

For proper security, apply middleware in this order:

```go
r.Use(corsMiddleware.Middleware)           // 1. CORS headers
r.Use(securityHeaders.Middleware)          // 2. Security headers
r.Use(rateLimiter.RateLimitMiddleware)     // 3. Rate limiting
r.Use(auditMiddleware.Middleware)          // 4. Audit logging
r.Use(middleware.LoggingMiddleware)        // 5. Application logging
r.Use(middleware.SecurityMiddleware)       // 6. Custom security
r.Use(middleware.AuthMiddleware)           // 7. Authentication
```

## Compliance Checklist

### GDPR Compliance

- ✅ **Article 15** - Right of Access: Data export API
- ✅ **Article 16** - Right to Rectification: Update APIs
- ✅ **Article 17** - Right to Erasure: Deletion with grace period
- ✅ **Article 18** - Right to Restriction: Deletion cancellation
- ✅ **Article 20** - Right to Data Portability: JSON export
- ✅ **Article 21** - Right to Object: Consent management
- ✅ **Article 30** - Records of Processing: Audit logs
- ✅ **Article 32** - Security: Encryption, access controls
- ✅ **Article 33** - Breach Notification: Audit trail for forensics

### Security Best Practices

- ✅ Rate limiting to prevent abuse
- ✅ Comprehensive audit logging
- ✅ Strong encryption (AES-256-GCM)
- ✅ Secure password hashing (bcrypt)
- ✅ Security headers (CSP, HSTS, etc.)
- ✅ CORS configuration
- ✅ Input validation
- ✅ Token-based authentication
- ✅ Automatic cleanup mechanisms

## Configuration

### Environment Variables

```bash
# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_DEFAULT_RPS=10

# Audit Logging
AUDIT_LOG_ENABLED=true
AUDIT_LOG_MAX_EVENTS=10000

# GDPR
GDPR_DELETION_GRACE_PERIOD_DAYS=30

# Encryption
ENCRYPTION_KEY=your-secret-encryption-key-here
```

## Performance Considerations

- **Rate Limiter**: Uses concurrent maps with RWMutex, automatic cleanup every 5 minutes
- **Audit Logger**: Circular buffer (10k events default) to prevent unbounded memory growth
- **GDPR Manager**: In-memory storage for demo; use database in production
- **Encryption**: AES-GCM is hardware-accelerated on modern CPUs

## Testing

```bash
# Test rate limiting
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/menus

# Check rate limit headers
curl -I http://localhost:8080/api/v1/auth/login

# Export GDPR data
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/gdpr/my-data -o my-data.json

# Get audit logs
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/audit/my-logs
```

## Production Deployment

1. **Use Strong Encryption Keys**: Generate with `openssl rand -hex 32`
2. **Enable HTTPS**: Required for HSTS and secure cookies
3. **Database Storage**: Replace in-memory stores with persistent database
4. **Log Rotation**: Implement log rotation for audit logs
5. **Monitoring**: Set up alerts for rate limit violations and security events
6. **Backup**: Regular backups of audit logs for compliance
7. **Review**: Periodic security audits and penetration testing

## License

Part of the QR Menu System - Enterprise Edition
