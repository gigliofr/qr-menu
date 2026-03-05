# 🚀 RAILWAY - CONFIGURAZIONE IMMEDIATA

**Status:** Pronto per deploy immediato  
**Data:** 2026-03-05  
**Commit:** b373e30

---

## ⚡ AZIONI IMMEDIATE RICHIESTE

### 1. Genera SESSION_SECRET (1 minuto)

**PowerShell:**
```powershell
-join ((1..32) | ForEach-Object {'{0:x2}' -f (Get-Random -Maximum 256)})
```

**Output esempio:**
```
a7b3c9d2e8f1a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9
```

**⚠️ IMPORTANTE: Salva questo valore, ti servirà per Railway!**

---

### 2. Railway Dashboard - Add Variables (2 minuti)

Vai su Railway Dashboard → Project → Variables → Add:

#### **Variable 1: SESSION_SECRET**
```
Key: SESSION_SECRET
Value: [il valore generato sopra]
```

#### **Variable 2: ENVIRONMENT**
```
Key: ENVIRONMENT
Value: production
```

#### **Variable 3: LOG_LEVEL**
```
Key: LOG_LEVEL
Value: INFO
```

#### **Variable 4: ENABLE_SEED_DATA**
```
Key: ENABLE_SEED_DATA
Value: false
```

---

### 3. Verifica MongoDB Vars (già configurate?)

Controlla che esistano già (se NO, vedi [RAILWAY_SETUP_GUIDE.md](RAILWAY_SETUP_GUIDE.md)):

- ✅ `MONGODB_URI`
- ✅ `MONGODB_CERT_CONTENT`
- ✅ `MONGODB_DB_NAME`

---

### 4. Trigger Redeploy (30 secondi)

Railway Dashboard → Deployments → Latest → **Redeploy**

O push vuoto:
```bash
git commit --allow-empty -m "trigger: Railway redeploy with new env vars"
git push origin main
```

---

## ✅ VERIFICATION CHECKLIST

Dopo il deploy, verifica:

```powershell
# 1. Health check
Invoke-RestMethod "https://qr-menu-production.up.railway.app/api/v1/health"

# 2. Check logs per SESSION_SECRET
# Railway logs dovrebbero mostrare:
# "✅ Using SESSION_SECRET from environment variable"

# 3. Test HTTPS redirect
# Prova ad accedere con http:// → dovrebbe fare redirect a https://

# 4. Test login
Start-Process "https://qr-menu-production.up.railway.app/login"
```

---

## 📋 COSA È STATO FIXATO (Commit: b373e30)

### 🐛 **Bug Fix Critici**
- ✅ **ShareMenuHandler syntax error** - Fix variabili mancanti che bloccavano build
- ✅ **Compilazione OK** - Verificato `go build` senza errori

### 🔒 **Security Improvements**
- ✅ **Password minima 8 char** (era 6)
- ✅ **Session secret da env var** - `SESSION_SECRET` invece di file
- ✅ **HTTPS redirect middleware** - Force HTTPS in prod/staging
- ✅ **Environment detection** - ENVIRONMENT var supportata

### ⚡ **Performance & Monitoring**
- ✅ **Rate Limiter già attivo** - Verificato in routes.go
- ✅ **Logging migliorato** - SESSION_SECRET usage tracking

---

## 📚 DOCUMENTAZIONE CREATA

1. **[RAILWAY_SETUP_GUIDE.md](RAILWAY_SETUP_GUIDE.md)** - Guida completa deployment
2. **[MAINTENANCE_PLAN.md](MAINTENANCE_PLAN.md)** - Piano manutenzione + multi-ambiente
3. **[CODE_REVIEW_FINDINGS.md](CODE_REVIEW_FINDINGS.md)** - Code review e optimization

---

## 🎯 PROSSIMI STEP (Opzionali)

Dopo verifica deploy OK:

### Settimana 1-2: Quick Wins Rimanenti
- [ ] Setup Railway staging project (branch: develop)
- [ ] MongoDB Atlas staging database
- [ ] CORS environment-specific (se necessario)
- [ ] Custom domain setup (opzionale)

### Settimana 3-4: Code Quality
- [ ] Split handlers.go (2000 lines → 8 file)
- [ ] Fix N+1 query in SetActiveMenuHandler
- [ ] CSRF token verification
- [ ] Test coverage > 70%

**Full roadmap:** Vedi [MAINTENANCE_PLAN.md](MAINTENANCE_PLAN.md)

---

## 🆘 TROUBLESHOOTING RAPIDO

### ❌ Build fails dopo push

**Check Railway logs:**
```
railway logs
```

**Common fixes:**
- Syntax error? → Vedi commit history, rollback se necessario
- Missing env var? → Aggiungi SESSION_SECRET su Railway

### ❌ "SESSION_SECRET not set" in logs

**Fix:**
1. Railway Dashboard → Variables
2. Add: `SESSION_SECRET` = [generated value]
3. Redeploy

### ❌ Redirect loop / HTTPS issues

**Check:**
- Railway Settings → Networking → HTTPS: Enabled
- Domain generated (non solo IP)
- Logs: verificare header `X-Forwarded-Proto`

**Full troubleshooting:** [RAILWAY_SETUP_GUIDE.md](RAILWAY_SETUP_GUIDE.md#-troubleshooting)

---

## 📞 SUMMARY PER IL TEAM

**Cosa è stato fatto:**
- Fix build error bloccante
- Implementati 4 security improvement critici in < 2 ore
- Documentazione completa (566+ linee guida Railway)
- Code review e maintenance plan (11500+ parole)

**Deploy status:**
- ✅ Codice pronto
- ⏳ Richiede setup env var Railway (5 min)
- ✅ Backward compatible (nessun breaking change)

**Rischi:**
- 🟢 ZERO - Tutte modifiche backward compatible
- Session fallback a file storage se SESSION_SECRET manca (dev mode)
- Password policy più stretta (ma solo per NUOVI utenti)

---

**READY TO DEPLOY!** 🚀

Deploy time stimato: **5 minuti** (setup env vars + redeploy)

---

**Q&A:**

**Q: Devo fare migration database?**  
A: NO, già fatta (migrate_restaurant_usernames.js eseguito)

**Q: Utenti esistenti devono cambiare password?**  
A: NO, policy 8 char si applica solo a NUOVE registrazioni

**Q: SESSION_SECRET obbligatorio?**  
A: NO in development (usa file fallback), SÌ in prod/staging (security best practice)

**Q: Cosa succede se non lo metto?**  
A: Funziona ma usa file storage (non consigliato, log warning)

**Q: Posso usare stesso SESSION_SECRET per staging e prod?**  
A: Tecnicamente sì, ma NON consigliato (security best practice: secret diversi)

