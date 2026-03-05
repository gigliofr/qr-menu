# 🚂 Railway Deployment Guide - QR Menu

**Versione:** 2.0  
**Data:** 2026-03-05  
**Status:** Production-Ready con Security Hardening

---

## 🎯 DEPLOYMENT RAPIDO (5 minuti)

### 1. Railway Project Setup

```bash
# Option A: Railway CLI
railway login
railway init
railway link

# Option B: Railway Dashboard
# 1. Vai su https://railway.app
# 2. New Project → Deploy from GitHub
# 3. Seleziona repository: qr-menu
# 4. Branch: main (prod) o develop (staging)
```

---

## 🔐 ENVIRONMENT VARIABLES RICHIESTE

### **CRITICAL (OBBLIGATORIE)**

#### MongoDB Atlas
```bash
MONGODB_URI=mongodb+srv://qr-menu-cluster.xxxxx.mongodb.net/qr-menu?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&appName=qr-menu

MONGODB_CERT_CONTENT=-----BEGIN CERTIFICATE-----
MIIDjTCCAnWgAwIBAgIIU...
[full certificate content]
...
-----END CERTIFICATE-----

MONGODB_DB_NAME=qr-menu
```

**Come ottenere:**
1. MongoDB Atlas → Clusters → Connect → Connect your application
2. Driver: Go, Version: 1.12 o superiore
3. Connection String: copia e incolla
4. Certificate: Database Access → Certificates → Download PEM

---

#### Session Secret (NUOVO - OBBLIGATORIO)
```bash
SESSION_SECRET=<64-char-hex-string>
```

**Generazione (PowerShell):**
```powershell
-join ((1..32) | ForEach-Object {'{0:x2}' -f (Get-Random -Maximum 256)})
```

**Generazione (Bash/Git Bash):**
```bash
openssl rand -hex 32
```

**Esempio output:**
```
7f3a9b2c8d1e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a
```

⚠️ **IMPORTANTE:** 
- Usa SECRET DIVERSI per staging e production
- NON committare nel repository
- Rotazione consigliata ogni 90 giorni

---

### **ENVIRONMENT SETTINGS**

#### Ambiente
```bash
ENVIRONMENT=production
# Valori possibili: development | staging | production
```

**Effetti:**
- `production` o `staging` → Abilita HTTPS redirect
- `production` o `staging` → Secure cookies attivati
- `development` → Session key da file (non consigliato in prod)

---

#### Logging
```bash
LOG_LEVEL=INFO
# Valori: DEBUG | INFO | WARN | ERROR
```

**Consigliato:**
- Production: `INFO` o `WARN`
- Staging: `DEBUG` (per troubleshooting)
- Development: `DEBUG`

---

### **OPTIONAL (Consigliati)**

#### CORS Origins
```bash
ALLOWED_ORIGINS=https://qr-menu.yourdomain.com
# Se multiple: separati da virgola
# ALLOWED_ORIGINS=https://domain1.com,https://domain2.com
```

**Default se non specificato:**
- Production: Permette solo railway.app domain
- Staging: Permette railway.app + localhost:3000
- Development: Permette localhost:3000 e localhost:8080

---

#### Seed Data (Staging Only)
```bash
ENABLE_SEED_DATA=false
# true = Carica dati di test all'avvio (SOLO STAGING!)
# false = Nessun seed (PRODUCTION default)
```

---

## 📋 TEMPLATE COMPLETO ENV VARS

### **Production Environment**

Copia e incolla su Railway Dashboard → Variables:

```bash
# === DATABASE ===
MONGODB_URI=mongodb+srv://qr-menu-cluster.xxxxx.mongodb.net/qr-menu?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&appName=qr-menu
MONGODB_CERT_CONTENT=-----BEGIN CERTIFICATE-----
[paste your certificate]
-----END CERTIFICATE-----
MONGODB_DB_NAME=qr-menu

# === SECURITY ===
SESSION_SECRET=[generate with: openssl rand -hex 32]
ENVIRONMENT=production

# === LOGGING ===
LOG_LEVEL=INFO

# === CORS (Optional) ===
ALLOWED_ORIGINS=https://qr-menu-production.up.railway.app

# === DATA ===
ENABLE_SEED_DATA=false
```

---

### **Staging Environment**

Per un secondo Railway project (staging):

```bash
# === DATABASE ===
MONGODB_URI=mongodb+srv://qr-menu-cluster.xxxxx.mongodb.net/qr-menu-staging?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&appName=qr-menu-staging
MONGODB_CERT_CONTENT=-----BEGIN CERTIFICATE-----
[paste your certificate]
-----END CERTIFICATE-----
MONGODB_DB_NAME=qr-menu-staging

# === SECURITY ===
SESSION_SECRET=[DIFFERENT from production!]
ENVIRONMENT=staging

# === LOGGING ===
LOG_LEVEL=DEBUG

# === CORS (Optional) ===
ALLOWED_ORIGINS=https://qr-menu-staging.up.railway.app,http://localhost:3000

# === DATA ===
ENABLE_SEED_DATA=true
```

---

## 🚀 DEPLOYMENT STEPS (Dettaglio)

### Step 1: MongoDB Atlas Setup

1. **Crea Cluster** (se non esiste)
   - Vai su MongoDB Atlas → Clusters → Create
   - Tier: M0 (Free) o M2+ (Prod)
   - Region: Scegli vicino a Railway (es. europe-west per Railway Europe)

2. **Crea Database**
   - Production: `qr-menu`
   - Staging: `qr-menu-staging`

3. **Network Access**
   - IP Whitelist: `0.0.0.0/0` (permetti tutte le IP - necessario per Railway)
   - ⚠️ Railway usa IP dinamici, quindi serve questo wildcard

4. **Database Access**
   - User: X509 Certificate Authentication
   - Download certificate (PEM format)
   - Copy full content (inclusi `-----BEGIN CERTIFICATE-----` e `-----END CERTIFICATE-----`)

---

### Step 2: Railway Project Creation

#### **Opzione A: Railway Dashboard (Consigliata)**

1. Vai su https://railway.app → New Project
2. **Deploy from GitHub Repo**
3. Seleziona repository: `qr-menu`
4. **Settings:**
   - Name: `qr-menu-production` (o `qr-menu-staging`)
   - Branch: `main` (prod) o `develop` (staging)
   - Root Directory: `/` (default)
   - Build Command: (automatico, rileva Go)
   - Start Command: (automatico, esegue binary compilato)

5. **Environment Variables:**
   - Click "Variables" tab
   - Add each variable from template above
   - ⚠️ Per `MONGODB_CERT_CONTENT`: usa "Multiline" mode

6. **Generate Domain:**
   - Settings → Networking → Generate Domain
   - O configura custom domain

---

#### **Opzione B: Railway CLI**

```bash
# Install CLI
npm i -g @railway/cli

# Login
railway login

# Link existing project (se già creato)
railway link

# O crea nuovo project
railway init

# Add variables (uno alla volta)
railway variables set MONGODB_URI="mongodb+srv://..."
railway variables set SESSION_SECRET="$(openssl rand -hex 32)"
railway variables set ENVIRONMENT="production"
railway variables set LOG_LEVEL="INFO"

# Deploy
git push origin main
# Railway auto-deploya su push
```

---

### Step 3: Verify Deployment

#### **Health Check**

```powershell
# Check che il servizio sia up
Invoke-RestMethod -Uri "https://qr-menu-production.up.railway.app/api/v1/health"

# Expected output:
# {
#   "status": "ok",
#   "timestamp": "2026-03-05T...",
#   "database": "connected",
#   "environment": "production"
# }
```

#### **Test Login**

```powershell
# Apri browser
Start-Process "https://qr-menu-production.up.railway.app/login"
```

**Credenziali seed (se ENABLE_SEED_DATA=true):**
- Username: `admin`
- Password: `admin123`

⚠️ **CAMBIA PASSWORD SUBITO IN PRODUZIONE!**

---

### Step 4: Database Migration (Prima Deploy)

Se hai dati esistenti da migrare:

```bash
# Local: Run migration script
cd scripts
node migrate_restaurant_usernames.js

# Verify migration
node verify_usernames.js
```

**Output atteso:**
```
✅ 10 restaurants migrated
✅ 0 duplicates
✅ Index created: idx_restaurants_username_unique
```

---

## 🔧 TROUBLESHOOTING

### ❌ Build Fails: "connect ECONNREFUSED"

**Problema:** MongoDB non raggiungibile

**Fix:**
1. Verifica `MONGODB_URI` sia corretto
2. Verifica `MONGODB_CERT_CONTENT` sia completo (inclusi header/footer)
3. Check MongoDB Atlas Network Access: IP `0.0.0.0/0` whitelisted
4. Railway logs: `railway logs` per vedere errore esatto

---

### ❌ "SESSION_SECRET not set, using file storage"

**Problema:** Session secret non configurato

**Fix:**
```bash
# Railway Dashboard → Variables → Add
SESSION_SECRET=$(openssl rand -hex 32)

# O Railway CLI
railway variables set SESSION_SECRET="$(openssl rand -hex 32)"
```

**Verificare:**
```bash
railway logs | grep -i "session"
# Dovrebbe mostrare: "✅ Using SESSION_SECRET from environment variable"
```

---

### ❌ Redirect Loop / HTTPS Issues

**Problema:** Railway non rileva HTTPS correttamente

**Fix:**
Verifica header `X-Forwarded-Proto` nel middleware (già implementato):

```go
// main.go - httpsRedirectMiddleware
if r.Header.Get("X-Forwarded-Proto") != "https" {
    // redirect...
}
```

**Verifica Railway Settings:**
- Settings → Networking → HTTPS: Enabled
- Domain deve essere generato (non solo IP)

---

### ❌ 429 Too Many Requests

**Problema:** Rate limiting troppo aggressivo per staging

**Temporaneo Fix:**
Aggiungi env var (staging only):
```bash
DISABLE_RATE_LIMIT=true
```

**Permanent Fix:**
Modifica `security/ratelimit.go` config per staging:
```go
var endpointConfigs = map[string]RateLimitConfig{
    "/api/auth/login": {
        RequestsPerSecond: 10, // Aumentato per staging
        BurstSize:         20,
    },
}
```

---

## 📊 MONITORING

### Railway Built-in Metrics

1. **Deploy Logs:**
   ```bash
   railway logs
   # O railway logs --follow
   ```

2. **Resource Usage:**
   - Railway Dashboard → Metrics
   - CPU, Memory, Network

3. **Application Logs:**
   ```bash
   railway logs | grep -i "error"
   railway logs | grep -i "security"
   railway logs | grep -i "failed login"
   ```

### Custom Alerts (Futuro)

Considera integrare:
- Sentry (error tracking)
- LogDNA/Logtail (log aggregation)
- Better Uptime (uptime monitoring)

---

## 🔄 CONTINUOUS DEPLOYMENT

### Auto-Deploy Setup

Railway auto-deploya su git push se configurato:

**Staging:**
```bash
# Branch: develop
git checkout develop
git merge feature/my-feature
git push origin develop
# → Railway staging auto-deploy
```

**Production:**
```bash
# Branch: main
git checkout main
git merge develop --no-ff
git tag v1.2.0
git push origin main --tags
# → Railway production auto-deploy
```

### Rollback

```bash
# Railway Dashboard → Deployments
# Click su deployment precedente → "Redeploy"

# O Railway CLI
railway rollback
```

---

## 📝 POST-DEPLOY CHECKLIST

Dopo ogni deploy production:

- [ ] ✅ Health check OK: `/api/v1/health`
- [ ] ✅ Login page accessible: `/login`
- [ ] ✅ Admin panel accessible: `/admin`
- [ ] ✅ Public menu works: `/r/{username}`
- [ ] ✅ QR code generation works
- [ ] ✅ Image upload works
- [ ] ✅ Analytics tracking works
- [ ] ✅ Session persistence (login, logout, re-login)
- [ ] ✅ Rate limiting attivo (test con 10+ rapidi request)
- [ ] ✅ HTTPS redirect funzionante (test http:// → https://)
- [ ] ✅ Logs puliti (no error critici)
- [ ] ✅ Database queries veloci (< 200ms P95)

---

## 🆘 EMERGENCY CONTACTS / LINKS

- **Railway Status:** https://status.railway.app
- **MongoDB Atlas Status:** https://status.cloud.mongodb.com
- **Railway Docs:** https://docs.railway.app
- **MongoDB Docs:** https://docs.mongodb.com/atlas
- **Project Repo:** [your GitHub repo URL]
- **Maintenance Plan:** `/MAINTENANCE_PLAN.md`
- **Code Review:** `/CODE_REVIEW_FINDINGS.md`

---

## 🎓 BEST PRACTICES

### Secrets Management
- ✅ SESSION_SECRET diverso per staging/prod
- ✅ Session secret rotation ogni 90 giorni
- ✅ Database user separato per staging/prod
- ✅ Never commit secrets to repo

### Monitoring
- ✅ Check Railway logs giornalmente in prod
- ✅ Setup alerts per high error rate
- ✅ Monitor database slow queries (MongoDB profiler)
- ✅ Track failed login attempts

### Performance
- ✅ Database indexes verificati (migrate_restaurant_usernames.js)
- ✅ Rate limiting attivo (anti-DDoS)
- ✅ HTTPS forzato (performance + security)
- ✅ Session cookie security (HttpOnly, Secure, SameSite)

### Security
- ✅ HTTPS redirect attivo
- ✅ Secure cookies in prod/staging
- ✅ Rate limiting per endpoint critici
- ✅ Password minima 8 caratteri
- ✅ Session timeout 7 giorni
- ✅ CORS configurato per domain specifici

---

**Documento creato:** 2026-03-05  
**Ultima modifica:** 2026-03-05  
**Versione:** 2.0  
**Author:** QR Menu DevTeam

**Prossimi Step:** Vedi [MAINTENANCE_PLAN.md](MAINTENANCE_PLAN.md) per roadmap completa

---

## 🚀 QUICK COMMAND REFERENCE

```bash
# Deploy
git push origin main                  # Auto-deploy production

# Logs
railway logs                           # View logs
railway logs --follow                  # Stream logs
railway logs | grep ERROR              # Filter errors

# Variables
railway variables                      # List all
railway variables set KEY=value        # Set variable
railway variables delete KEY           # Delete variable

# Status
railway status                         # Project status
railway open                           # Open dashboard

# Rollback
railway rollback                       # Rollback to previous

# Database
railway run node scripts/migrate_restaurant_usernames.js
```

---

**PRONTO PER DEPLOY!** 🎉
