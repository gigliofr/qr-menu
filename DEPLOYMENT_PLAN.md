# 🚀 Piano Deployment QR-Menu - Google Cloud Run

## 📊 Riepilogo Piano

**Piattaforma**: Google Cloud Run (FREE TIER)  
**Database**: MongoDB Atlas (FREE TIER)  
**Costo Totale**: $0 (per test e piccolo carico)  
**Tempo Totale**: ~1-2 ore

---

## 📋 TASK BREAKDOWN

### ✅ COMPLETATI (2/8)

#### Task 1: Dockerfile ✅
- **Status**: ✅ COMPLETATO
- **File**: `Dockerfile`
- **Dettagli**: Multi-stage build, distroless image, pronto per Cloud Run
- **Validazione**: Build ottimizzato, sicuro, leggero (~50MB)

#### Task 2: .gcloudignore ✅
- **Status**: ✅ COMPLETATO
- **File**: `.gcloudignore`
- **Dettagli**: Esclude file inutili, riduce upload size
- **Benefici**: Deploy più veloce (~10MB instead of ~500MB)

---

### 🔄 IN PROGRESS (6/8)

#### Task 3: Setup Google Cloud Project
**Tempo Stimato**: 15 minuti
**Passaggi**:
1. Installa Google Cloud SDK
2. `gcloud auth login`
3. Crea project: `gcloud projects create qr-menu-prod`
4. Abilita API: `gcloud services enable run.googleapis.com`
5. Crea Secret Manager per certificato X.509
6. Configura IAM permissions

**Comando da eseguire**:
```bash
gcloud auth login
gcloud projects create qr-menu-prod --name="QR Menu"
gcloud config set project qr-menu-prod
gcloud services enable run.googleapis.com cloudbuil.googleapis.com
```

---

#### Task 4: Deploy su Cloud Run
**Tempo Stimato**: 10 minuti
**Passaggi**:
1. Carica certificato nel Secret Manager
2. Deploy con `gcloud run deploy`
3. Configura environment variables
4. Imposta resource limits (512Mi RAM, 1 CPU)

**Comando principale**:
```bash
cd c:\Users\gigli\GoWs\qr-menu

gcloud run deploy qr-menu \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1
```

---

#### Task 5: Verificare connessione MongoDB
**Tempo Stimato**: 5 minuti
**Checklist**:
- [ ] Certificato X.509 caricato correttamente
- [ ] Connection string nel Secret Manager
- [ ] MongoDB Atlas whitelist include IP Cloud Run
- [ ] Log della connessione visibili in Cloud Run console

**Log check**:
```bash
gcloud run logs read qr-menu --region us-central1
```

---

#### Task 6: Test endpoint pubblico
**Tempo Stimato**: 10 minuti
**Test da eseguire**:

1. **Health Check**
```bash
curl https://qr-menu-[ID].a.run.app/health
```
Aspettato:
```json
{
  "success": true,
  "status": "healthy",
  "data": {
    "database": "mongodb",
    "services": {
      "authentication": "running",
      "logging": "running"
    }
  }
}
```

2. **API Restaurants**
```bash
curl https://qr-menu-[ID].a.run.app/api/restaurants
```

3. **Creazione Ristorante Test**
```bash
curl -X POST https://qr-menu-[ID].a.run.app/api/restaurants \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Restaurant",
    "email": "test@example.com",
    "phone": "+39 06 123456",
    "address": "Via Test 1, Roma"
  }'
```

---

#### Task 7: Setup monitoring e logs
**Tempo Stimato**: 5 minuti
**Configurazioni**:

1. **Cloud Run Dashboard**
   - URL: `https://console.cloud.google.com/run`
   - Visualizza: Requests, Errors, Latency, CPU/Memory

2. **Logging Configuration**
```bash
# Real-time logs
gcloud run logs read qr-menu --follow --region us-central1

# Filtro per errori
gcloud run logs read qr-menu --region us-central1 \
  --limit 50 | grep ERROR
```

3. **Alerts Setup** (opzionale)
   - Configure alert se error rate > 5%
   - Configure alert se latency > 1s

---

#### Task 8: Documentazione Deployment
**File**: `CLOUD_RUN_DEPLOYMENT.md` ✅ CREATO
**Contiene**:
- [ ] Step-by-step setup
- [ ] Secret Manager configuration
- [ ] Troubleshooting guide
- [ ] Cost breakdown
- [ ] Monitoring setup
- [ ] Performance tuning

---

## 📈 Ordine di Esecuzione Consigliato

```
1. ✅ Task 1-2: Dockerfile + .gcloudignore (COMPLETATI)
   ↓
2. 🔄 Task 3: Setup Google Cloud (15 min)
   ↓
3. 🔄 Task 4: Deploy su Cloud Run (10 min)
   ↓
4. 🔄 Task 5: Verifica MongoDB (5 min)
   ↓
5. 🔄 Task 6: Test Endpoint (10 min)
   ↓
6. 🔄 Task 7: Monitoring (5 min)
   ↓
7. ✅ Task 8: Documentazione (DONE)

⏱️ TEMPO TOTALE: ~45-60 minuti
```

---

## 💰 Costi Dettagliati

### Google Cloud Run (FREE TIER)
| Risorsa | Limite Gratis | Uso Stimato | Costo |
|---------|---------------|-----------|-------|
| Requests | 2M/mese | ~100k/mese (test) | FREE |
| Compute | 180k cpu-seconds/mese | ~10k/mese | FREE |
| Memory | 360k memory-seconds/mese | ~5k/mese | FREE |
| Outbound | 1GB/mese | ~100MB/mese | FREE |

### MongoDB Atlas (FREE TIER)
| Risorsa | Limite | Costo |
|---------|--------|-------|
| Storage | 512MB | FREE |
| Throughput | Unlimited | FREE |
| Connections | Unlimited | FREE |

### **TOTALE: $0/mese**

---

## 🎯 Success Criteria

✅ Deployment considerato SUCCESSO quando:

1. [ ] Cloud Run health endpoint risponde
2. [ ] Certificato MongoDB verifi
cato correttamente
3. [ ] API risponde alle richieste
4. [ ] Log visibili in console
5. [ ] Response time < 1 secondo
6. [ ] Zero errors per 10 minuti di testing

---

## 🚨 Contingency Plan

Se qualcosa fallisce:

| Problema | Soluzione |
|----------|-----------|
| Deploy fallisce | Controlla `gcloud run logs`; verifica Dockerfile |
| MongoDB connection error | Whitelist IP Cloud Run in MongoDB Atlas |
| Certificate error | Ricarica secret in Secret Manager |
| Out of memory | Aumenta RAM: `--memory 1Gi` |
| Timeout | Aumenta timeout: `--timeout 3600` |

---

## 📞 Support Resources

- Cloud Run Docs: https://cloud.google.com/run/docs
- Troubleshooting: https://cloud.google.com/run/docs/troubleshooting
- MongoDB Atlas Docs: https://docs.mongodb.com/atlas/
- gcloud Reference: https://cloud.google.com/sdk/gcloud/reference

---

**🎬 PRONTO A INIZIARE?**

Prossimo step: `gcloud auth login` per autenticarti con Google Cloud!
