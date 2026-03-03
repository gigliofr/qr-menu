# QR Menu System 🍽️

Sistema di gestione menu digitali con codici QR per ristoranti. Semplice, solido, production-ready.

**Versione:** 3.0.0 Simplified  
**Stack:** Go 1.24 + MongoDB Atlas  
**Deploy:** Railway.app  

---

## 🚀 Deploy su Railway (Produzione)

### 1. Setup MongoDB Atlas

1. Crea account gratuito su [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)
2. Crea cluster (free tier M0)
3. Vai a: **Security → Database Access → Add Database User**
4. Scegli: **Certificate** (X.509 authentication)
5. Scarica il certificato PEM (include certificate + private key)

### 2. Setup Railway

1. Vai su [Railway.app](https://railway.app) e crea progetto
2. Collega repository GitHub: `https://github.com/gigliofr/qr-menu`
3. Configura variabili d'ambiente:

```bash
MONGODB_URI=mongodb+srv://qr-menu-dev@cluster0.XXXXX.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509
MONGODB_CERT_CONTENT=<incolla-contenuto-certificato-PEM>
MONGODB_DB_NAME=qr-menu
```

**⚠️ IMPORTANTE:** Per `MONGODB_CERT_CONTENT`, copia l'intero contenuto del file PEM:
- Include sia `-----BEGIN CERTIFICATE-----` che `-----BEGIN PRIVATE KEY-----`
- Mantieni i newlines originali

### 3. Deploy Automatico

Push su GitHub → Railway rileva automaticamente Dockerfile → Build & Deploy! 🎉

**URL Pubblico**: `https://qr-menu-production-XXXX.up.railway.app`

---

## 🏠 Sviluppo Locale

### Prerequisiti
- Go 1.24+
- MongoDB Atlas account
- Certificato X.509 (`.pem`)

### Setup

```bash
# 1. Clone
git clone https://github.com/gigliofr/qr-menu.git
cd qr-menu

# 2. Configura variabili d'ambiente
$env:MONGODB_URI="mongodb+srv://..."
$env:MONGODB_CERT_PATH="C:\path\to\cert.pem"  # Per dev locale
$env:MONGODB_DB_NAME="qr-menu"

# 3. Build & Run
go build -o qr-menu .
./qr-menu

# Oppure direttamente
go run main.go
```

**Server avviato**: `http://localhost:8080`

---

## 📡 API Endpoints

### Autenticazione
- `GET  /login` - Pagina login
- `POST /login` - Effettua login
- `GET  /register` - Pagina registrazione
- `POST /register` - Crea account
- `GET  /logout` - Logout

### Menu Management
- `GET  /admin` - Dashboard amministrativa
- `POST /api/v1/menu` - Crea menu
- `GET  /api/v1/menu` - Lista menu
- `GET  /api/v1/menu/{id}` - Dettagli menu
- `PUT  /api/v1/menu/{id}` - Aggiorna menu
- `DELETE /api/v1/menu/{id}` - Elimina menu

### Public
- `GET  /menu/{id}` - Visualizza menu pubblico (per clienti)
- `GET  /qr/{id}` - Scarica QR code del menu

### Monitoring
- `GET  /api/v1/health` - Health check
- `GET  /api/v1/metrics` - Metriche (autenticato)

---

## 🏗️ Architettura

```
qr-menu/
├── main.go              # Entry point
├── db/                  # MongoDB layer
│   ├── mongo.go         # Connection & CRUD
│   └── mongo_features.go
├── api/                 # REST API handlers
│   ├── auth.go
│   ├── menu.go
│   ├── restaurant.go
│   └── router.go
├── handlers/            # HTTP handlers
├── models/              # Data models
├── pkg/                 # Core packages
│   ├── app/             # App initialization
│   ├── cache/           # Response caching
│   ├── middleware/      # HTTP middlewares
│   └── routing/         # Router setup
├── security/            # Security (rate limit, audit, GDPR)
├── analytics/           # Analytics & metrics
├── logger/              # Structured logging
├── templates/           # HTML templates
├── static/              # CSS, JS, images
└── web/                 # Frontend assets
```

**Database:** MongoDB Atlas (solo database, no file storage)  
**Auth:** X.509 certificate authentication  
**Session:** Cookie-based con Gorilla sessions  
**Caching:** In-memory response cache  

---

## 🔒 Sicurezza

- **Autenticazione X.509**: Certificate-based per MongoDB
- **Password Hashing**: bcrypt
- **Sessions**: Secure HTTP-only cookies
- **Rate Limiting**: Protezione contro brute-force
- **Audit Logging**: Tracking azioni utente
- **GDPR Compliance**: Data export/deletion
- **Security Headers**: HSTS, CSP, X-Frame-Options

---

## 🧪 Testing

```bash
# Run test esistenti
go test ./...

# Run test con coverage
go test -cover ./...

# Test specifici
go test ./pkg/cache/...
go test ./pkg/middleware/...
```

**Test Coverage**: ~85% (core packages)

---

## 📊 Monitoring

### Railway Dashboard
- Build logs: Railway Project → Deployments → Build Logs
- Runtime logs: Railway Project → Deployments → View Logs
- Metrics: CPU, Memory, Network usage

### Health Check
```bash
curl https://your-app.up.railway.app/api/v1/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "mongodb": "connected",
  "uptime": "2h15m30s"
}
```

---

## 🐛 Troubleshooting

### Problema: `TLS internal error` MongoDB
**Causa**: Certificato X.509 incompleto o malformato  
**Fix**: Verifica che `MONGODB_CERT_CONTENT` includa ENTRAMBI:
- `-----BEGIN CERTIFICATE-----` ... `-----END CERTIFICATE-----`
- `-----BEGIN PRIVATE KEY-----` ... `-----END PRIVATE KEY-----`

### Problema: `pattern matches no files: templates/*.html`
**Causa**: Template non copiati in Docker build  
**Fix**: Verificato - Dockerfile ora copia correttamente `templates/`, `static/`, `web/`

### Problema: App crash al startup
**Causa**: MongoDB non configurato (ora obbligatorio)  
**Fix**: Configura tutte le variabili d'ambiente richieste

---

## 📝 License

MIT License - vedi [LICENSE](LICENSE) per dettagli

---

## 🤝 Contributing

Questo è un progetto semplificato per produzione. Feature enterprise (ML, PWA, notifications) sono state rimosse per mantenere il codice pulito e manutenibile.

Per modifiche:
1. Fork del repository
2. Crea feature branch
3. Commit con messaggi chiari
4. Push e crea Pull Request

---

## 📞 Support

**Issues**: [GitHub Issues](https://github.com/gigliofr/qr-menu/issues)  
**Railway Project**: [https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab](https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab)

---

**Built with ❤️ for restaurants**
