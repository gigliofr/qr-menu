# QR Menu System - Guida Completa
*Versione 2.0.0 | Ultima modifica: 24 Febbraio 2026*

---

## 📑 Indice

1. [Panoramica](#panoramica)
2. [Quick Start](#quick-start)
3. [Architettura](#architettura)
4. [Funzionalità](#funzionalità)
5. [API Reference](#api-reference)
6. [Deployment](#deployment)
7. [Security & Compliance](#security--compliance)
8. [Testing](#testing)
9. [Contributing](#contributing)

---

## Panoramica

**QR Menu System** è una piattaforma enterprise per la gestione digitale di menu tramite QR code, con supporto multi-ristorante, analytics avanzati, ML/AI e compliance GDPR.

### Tecnologie

- **Backend**: Go 1.24, Gorilla Mux
- **Frontend**: React 18, Next.js 14, TypeScript, Tailwind CSS
- **Mobile**: Flutter 3.x (iOS/Android)
- **Infra**: Docker, Kubernetes
- **Database**: PostgreSQL (opzionale, in-memory per demo)
- **Security**: AES-256-GCM, JWT, RBAC, Rate Limiting

### Features Principali

✅ Menu digitali con QR code
✅ Multi-ristorante & multi-utente
✅ Analytics real-time
✅ Notifiche push (FCM)
✅ Backup automatici
✅ Localizzazione (i18n)
✅ PWA (offline-first)
✅ ML: Recommendations, Forecasting, A/B Testing
✅ RBAC (5 ruoli, 11 permessi)
✅ Stripe payments
✅ GDPR compliance

---

## Quick Start

### Prerequisiti

- Go 1.24+
- Node.js 18+ (per frontend)
- Flutter 3.x (per mobile)
- Docker (opzionale)

### Installazione Locale

```bash
# 1. Clone repository
git clone https://github.com/yourusername/qr-menu.git
cd qr-menu

# 2. Build backend
go build -o qr-menu.exe .

# 3. Avvia server
./qr-menu.exe
# Server: http://localhost:8080
# Admin: http://localhost:8080/admin
```

### Variabili d'Ambiente

```bash
# Server
PORT=8080

# Database (opzionale)
DATABASE_URL=postgres://user:pass@localhost:5432/qrmenu

# JWT
JWT_SECRET=your_secret_key_here

# Stripe (opzionale)
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### Docker

```bash
# Build
docker build -t qr-menu:latest .

# Run
docker run -p 8080:8080 qr-menu:latest
```

### Kubernetes

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

---

## Architettura

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Mobile    │────▶│   Backend    │────▶│  Database   │
│  (Flutter)  │     │    (Go)      │     │ (optional)  │
└─────────────┘     └──────────────┘     └─────────────┘
                           │
                           ▼
┌─────────────┐     ┌──────────────┐
│   Frontend  │────▶│   Storage    │
│   (React)   │     │ (in-memory)  │
└─────────────┘     └──────────────┘
```

### Struttura Progetto

```
qr-menu/
├── main.go                    # Entry point (50 righe - refactored!)
├── pkg/
│   └── app/
│       ├── initializer.go     # Service initialization
│       └── routes.go          # Route setup
├── api/                       # REST API handlers
├── handlers/                  # HTTP handlers
├── middleware/                # HTTP middleware
├── models/                    # Data models
├── analytics/                 # Analytics engine
├── backup/                    # Backup system
├── notifications/             # Push notifications
├── localization/              # i18n support
├── pwa/                       # PWA manager
├── security/                  # Security & GDPR
├── ml/                        # ML/Analytics
│   ├── recommendations.go     # Collaborative filtering
│   ├── predictions.go         # Forecasting
│   └── abtesting.go          # A/B testing
├── frontend/                  # React app
├── mobile/                    # Flutter app
└── k8s/                       # Kubernetes manifests
```

---

## Funzionalità

### 1. Gestione Menu

- Creazione menu multi-categoria
- Upload immagini piatti
- Prezzi, allergeni, disponibilità
- Menu multipli per ristorante
- Duplicazione menu/item
- QR code dinamici

### 2. Analytics

**Metriche disponibili:**
- Visualizzazioni totali/uniche
- Analisi oraria/giornaliera
- Device, OS, browser breakdown
- Geolocalizzazione
- Item popolari
- Share tracking (WhatsApp, Telegram, Facebook)
- Scan QR code

**API:** `GET /api/analytics?restaurant_id={id}`

### 3. Notifiche Push

**Canali supportati:**
- Email
- SMS
- Push notifications (FCM)
- In-app

**Features:**
- Preferenze utente
- Quiet hours
- Retry con exponential backoff
- Storia notifiche

**API:**
```bash
POST /api/notifications/send
GET /api/notifications/history
PUT /api/notifications/preferences
```

### 4. Backup & Restore

**Caratteristiche:**
- Backup automatici schedulati
- Compressione ZIP
- Retention policy (30 giorni default)
- Restore da backup
- Metadata tracking

**API:**
```bash
POST /api/backup/create
GET /api/backup/list
POST /api/backup/restore
```

### 5. Localizzazione (i18n)

**Lingue supportate:**
- Italiano (it)
- English (en)
- Español (es)
- Français (fr)
- Deutsch (de)

**Features:**
- Traduzioni dinamiche
- Formattazione valuta/date
- Preferenze utente
- Fallback automatico

**API:**
```bash
GET /api/localization/translations?locale=it
POST /api/localization/set-locale
```

### 6. PWA (Progressive Web App)

**Funzionalità:**
- Offline-first
- Installabile
- Service Worker
- Cache intelligente
- Notifiche push web

**File generati:**
- `/manifest.json`
- `/service-worker.js`
- `/offline.html`

### 7. Machine Learning

#### Recommendation Engine
- Collaborative filtering
- Item similarity (cosine, Pearson, Jaccard)
- Cold start handling
- Trending items

**API:** `GET /api/v1/ml/recommendations?limit=10`

#### Predictive Analytics
- Demand forecasting (Holt-Winters)
- Seasonality detection
- Trend analysis
- Inventory optimization

**API:** `GET /api/v1/ml/forecast?metric=orders&periods=7`

#### A/B Testing
- Multi-variant experiments
- Statistical significance (Z-test)
- Conversion tracking

**API:** `POST /api/v1/ml/experiments`

---

## API Reference

### Authentication

Tutte le API protette richiedono JWT token:
```
Authorization: Bearer <token>
```

**Login:**
```bash
POST /login
{
  "username": "admin",
  "password": "password"
}
```

### Core Endpoints

#### Menu Management
```bash
GET    /api/menus                    # List menus
POST   /api/menu                     # Create menu
GET    /api/menu/{id}                # Get menu
POST   /api/menu/{id}/generate-qr   # Generate QR code
DELETE /admin/menu/{id}/delete       # Delete menu
```

#### Analytics
```bash
GET /api/analytics?restaurant_id={id}
POST /api/track/share               # Track share event
```

#### Backup
```bash
POST   /api/backup/create
GET    /api/backup/list
DELETE /api/backup/delete?backup_id={id}
POST   /api/backup/restore
```

#### Notifications
```bash
POST /api/notifications/send
GET  /api/notifications/history
PUT  /api/notifications/preferences
```

#### ML & Analytics
```bash
GET  /api/v1/ml/recommendations
GET  /api/v1/ml/forecast
POST /api/v1/ml/experiments
GET  /api/v1/ml/experiments/{id}/results
```

### RBAC Endpoints

```bash
GET  /api/v1/rbac/roles
POST /api/v1/rbac/assign-role
GET  /api/v1/rbac/check-permission
```

**5 Ruoli disponibili:**
1. Super Admin
2. Restaurant Owner
3. Restaurant Manager
4. Staff
5. Customer

### Billing & Payments (Stripe)

```bash
POST /api/v1/checkout/create         # Create checkout session
GET  /api/v1/checkout/success        # Payment success callback
POST /api/v1/checkout/customer-portal
```

### Security & GDPR

```bash
GET    /api/v1/gdpr/export           # Export user data (JSON)
DELETE /api/v1/gdpr/delete           # Delete user data
POST   /api/v1/gdpr/consent          # Update consent
GET    /api/v1/security/audit-logs   # View audit logs
```

---

## Deployment

### Production Checklist

- [ ] Configurare `DATABASE_URL` (PostgreSQL)
- [ ] Impostare `JWT_SECRET` sicuro (32+ caratteri)
- [ ] Configurare Stripe keys (se billing attivo)
- [ ] Abilitare HTTPS/TLS
- [ ] Configurare backup automatici
- [ ] Setup monitoring (Prometheus/Grafana)
- [ ] Configurare log aggregation (ELK)
- [ ] Test di carico
- [ ] Security audit
- [ ] Backup strategy

### Environment Variables (Production)

```bash
# Server
PORT=8080
ENV=production

# Database
DATABASE_URL=postgres://user:pass@db-host:5432/qrmenu

# JWT
JWT_SECRET=your_production_secret_32chars_min

# Stripe
STRIPE_SECRET_KEY=sk_live_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Security
RATE_LIMIT_GLOBAL=1000
RATE_LIMIT_PER_IP=100
ENCRYPTION_KEY=32_byte_key_for_aes256

# Notifications
FCM_SERVER_KEY=your_fcm_server_key
```

### Docker Compose (Production)

```yaml
version: '3.8'
services:
  app:
    image: qr-menu:latest
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://user:pass@db:5432/qrmenu
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - db
  
  db:
    image: postgres:15
    environment:
      - POSTGRES_USER=qrmenu
      - POSTGRES_PASSWORD=secure_password
      - POSTGRES_DB=qrmenu
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

### Kubernetes (Production)

```bash
# Secrets
kubectl create secret generic qr-menu-secrets \
  --from-literal=jwt-secret=$JWT_SECRET \
  --from-literal=db-password=$DB_PASSWORD

# Deploy
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml

# Scale
kubectl scale deployment qr-menu --replicas=5
```

---

## Security & Compliance

### Security Features

1. **Rate Limiting**: Token bucket (100 req/min per IP)
2. **Audit Logging**: Tutti gli eventi critici registrati
3. **Encryption**: AES-256-GCM per dati sensibili
4. **Security Headers**: CSP, HSTS, X-Frame-Options
5. **CORS**: Configurabile per origin
6. **Input Validation**: Sanitizzazione automatica
7. **JWT Authentication**: Token sicuri con expiry

### GDPR Compliance

- ✅ Right to access (export data)
- ✅ Right to be forgotten (delete data)
- ✅ Consent management
- ✅ Data minimization
- ✅ Purpose limitation
- ✅ Audit trail completo

### Best Practices

1. **Password**: Bcrypt hashing (cost 12)
2. **JWT**: Short-lived tokens (1h) + refresh tokens
3. **API Keys**: Rotazione periodica
4. **Logs**: No dati sensibili nei log
5. **HTTPS**: TLS 1.3 in produzione
6. **Database**: Connessioni cifrate
7. **Backups**: Encrypted at rest

---

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./api/...
go test ./ml/...
```

### Integration Tests

```bash
# Test containers
go test -tags=integration ./...

# Test API endpoints
./test_api.ps1
```

### Load Testing

```bash
# Apache Bench
ab -n 1000 -c 10 http://localhost:8080/api/menus

# Hey
hey -n 1000 -c 50 http://localhost:8080/api/v1/health
```

### Test Utente (manuale)

Vedi [TESTING_GUIDE.md](TESTING_GUIDE.md) per il piano completo.

---

## Contributing

1. Fork il repository
2. Crea feature branch (`git checkout -b feature/amazing-feature`)
3. Commit (`git commit -m 'Add amazing feature'`)
4. Push (`git push origin feature/amazing-feature`)
5. Apri Pull Request

### Coding Standards

- Go: `gofmt`, `golint`, `go vet`
- React: ESLint, Prettier
- Commit messages: Conventional Commits
- Tests: Coverage > 80%

---

## Support & Documentation

- **API Docs**: http://localhost:8080/api/v1/docs
- **Health Check**: http://localhost:8080/ping
- **Admin Panel**: http://localhost:8080/admin

## License

Copyright © 2026 QR Menu System. All rights reserved.

---

**Built with ❤️ in Italy**
