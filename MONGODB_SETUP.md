# MongoDB Atlas Setup Guide

## 1. Una tantum Setup su MongoDB Atlas

### 1.1 Creare un Cluster MongoDB Atlas

1. Vai su [mongodb.com/cloud/atlas](https://mongodb.com/cloud/atlas)
2. Crea account e login
3. Crea un nuovo **Shared Cluster** (gratuito):
   - **Cloud**: AWS, GCP, o Azure
   - **Region**: `eu-west-1` (Europa) per bassa latenza
   - **Cluster Tier**: M0 Sandbox (gratuito, 512MB)
   - **Name**: `qr-menu-prod` o simile

### 1.2 Configurare X.509 Certificate Authentication

#### Opzione A: Certificato fornito
Se hai il certificato fornito dal team (`X509-cert-4084673564018728353.pem`):

1. In **Atlas Dashboard** → **Security** → **Database Access**
2. Clicca **"X.509 Certificate"** tab
3. Carica il file PEM (il certificato abbiamo già)
4. Accetta i termini
5. Copia la **Connection String** (vedi sotto)

#### Opzione B: Generare nuovo certificato
Se devi generare uno nuovo:

```bash
# MongoDB Atlas auto-genera certificati X.509
# Nel dashboard: Security → Database Access → X.509 Certificate
# Clicca "Generate certificate" → scarica il file .pem

# Il certificato è valido per 365 giorni
```

### 1.3 IP Whitelist
1. **Security** → **Network Access**
2. Clicca **"Add IP Address"**
3. Per development: `0.0.0.0/0` (INSECURO!) oppure
4. Per production: IP specifico della macchina/server

---

## 2. Configurazione Applicazione

### 2.1 Environment Variables

Aggiungi a `.env` o `.bashrc` (Windows PowerShell):

```powershell
# Windows PowerShell
$env:MONGODB_URI="mongodb+srv://cert-issuer@cluster0.a1b2c3d.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509"
$env:MONGODB_CERT_PATH="C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem"
$env:MONGODB_DB_NAME="qr-menu"
$env:MIGRATE_FROM_FILES="true"  # Solo primo avvio!
```

### 2.2 File Placement

```
Project Root/
├── C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem  ← Certificato
└── Codice Go
```

---

## 3. Connection String Formato

### Con X.509 Certificate

```
mongodb+srv://qr-menu-dev@cluster0.a1b2c3d3.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509&tlsCAFile=<path-to-cert>
```

**Parametri:**
- `cluster0.a1b2c3d3.mongodb.net`: Endpoint Atlas (da copiare da UI)
- `authSource=$external`: Usa X.509 esterno
- `authMechanism=MONGODB-X509`: Autenticazione X.509
- `tlsCAFile`: Path al certificato PEM

### Alternative Connection String

Se la URI di sopra non funziona, prova:

```
mongodb+srv://cluster0.a1b2c3d3.mongodb.net:27017/?authSource=$external&authMechanism=MONGODB-X509&retryWrites=true&w=majority
```

---

## 4. Verifica della Connessione

### Con mongosh (MongoDB CLI)

```bash
# Installa mongosh (se non hai)
# macOS: brew install mongosh
# Windows: scoop install mongosh

# Connetti con X.509
mongosh "mongodb+srv://cluster0.a1b2c3d3.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509" \
  --tlsCertificateKeyFile C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem

# Testa
> use qr-menu
> db.restaurants.countDocuments()
```

### Via Applicazione Go

Avvia il server:

```bash
go run main.go
```

Dovresti vedere:
```
✓ Connesso a MongoDB Atlas
✓ Indici creati
```

Se fallisce, controlla i logs per errori di:
- Certificato non trovato
- IP non whitelistato
- Connection string non valida

---

## 5. Collections Schema

### Collection: restaurants
```javascript
db.createCollection("restaurants", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["id", "username", "email", "password_hash"],
      properties: {
        id: { bsonType: "string", description: "UUID" },
        username: { bsonType: "string" },
        email: { bsonType: "string"},
        password_hash: { bsonType: "string" },
        role: { enum: ["admin", "owner", "staff"] },
        name: { bsonType: "string" },
        is_active: { bsonType: "bool" },
        created_at: { bsonType: "date" },
        last_login: { bsonType: "date" }
      }
    }
  }
})

// Indici
db.restaurants.createIndex({ id: 1 }, { unique: true })
db.restaurants.createIndex({ username: 1 }, { unique: true })
db.restaurants.createIndex({ email: 1 }, { unique: true })
```

### Collection: menus
```javascript
db.createCollection("menus", {
  validator: {
    $jsonSchema: {
      bsonType: "object",
      required: ["id", "restaurant_id", "name"],
      properties: {
        id: { bsonType: "string" },
        restaurant_id: { bsonType: "string" },
        name: { bsonType: "string" },
        meal_type: { enum: ["breakfast", "lunch", "dinner", "generic"] },
        is_active: { bsonType: "bool" },
        created_at: { bsonType: "date" }
      }
    }
  }
})

// Indici
db.menus.createIndex({ id: 1 }, { unique: true })
db.menus.createIndex({ restaurant_id: 1 })
db.menus.createIndex({ created_at: -1 })
```

### Collection: sessions
```javascript
db.createCollection("sessions")

// TTL Index (auto-delete dopo 7 giorni)
db.sessions.createIndex(
  { last_accessed: 1 },
  { expireAfterSeconds: 604800 }
)

db.sessions.createIndex({ id: 1 })
db.sessions.createIndex({ restaurant_id: 1 })
```

---

## 6. Backup & Restore

### Backup Automatico (MongoDB Atlas)

✅ **Incluso nel piano gratuito!**
- Backup orario della last 7 giorni
- Backup settimanale per 1 anno
- Trovabile in: **Backup** → **Restore** tab

### Backup Manuale (JSON)

```bash
# Esporta da MongoDB a JSON
# (la funzione è in db/mongo_migration.go)

# Nel codice Go:
err := db.MongoInstance.BackupToJSON("./backups")
```

Questo crea file JSON in `./backups/`:
```
backups/
├── restaurant_uuid1.json
├── restaurant_uuid2.json
├── menu_uuid1.json
└── menu_uuid2.json
```

### Restore da Backup JSON

```bash
# Riavvia con flag migrazione
MIGRATE_FROM_FILES=true go run main.go
```

---

## 7. Monitoring & Performance

### In Atlas Dashboard

1. **Metrics** → Visualizza CPU, RAM, operazioni
2. **Logs** → Leggi i log di query lente
3. **Performance Advisor** → Suggerimenti di indici
4. **Alerts** → Notifiche su anomalie

### Query Performance

```javascript
// Top 5 query più lente
db.system.profile.find({ millis: { $gt: 100 } })
  .sort({ ts: -1 })
  .limit(5)
  .pretty()
```

---

## 8. Troubleshooting

### ❌ "Error: server selection timed out"

**Cause:**
- Certificato non trovato
- IP non aggiunto alla whitelist
- Cluster non attivo

**Fix:**
```bash
# Controlla certificato
Test-Path "C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem"

# Aggiungi IP a Network Access
# Atlas → Security → Network Access → Add IP
```

### ❌ "ErrorCode:13"

**Significa:** Autenticazione fallita

**Fix:**
```bash
# Verifica che il CN del certificato sia corretto
# openssl x509 -in cert.pem -text -noout | grep "Subject:"
# Dovrebbe dire: "qr-menu-dev" o simile
```

### ❌ "EOF" error in logs

**Significa:** Connessione TLS interrotta

**Fix:**
- Genera nuovo certificato su Atlas
- Scarica e usa quello nuovo
- Riavvia applicazione

---

## 9. Security Best Practices

### ✅ DO:
- ✓ Usa X.509 certificate per autenticazione
- ✓ IP whitelist stretto in produzione
- ✓ Enable encryption in transit (default MongoDB Atlas)
- ✓ Enable encryption at rest (M10+ su Atlas)
- ✓ Rotate certificati annualmente
- ✓ Monitor audit logs

### ❌ DON'T:
- ✗ Non mettere certificato in repo Git
- ✗ Non usare `0.0.0.0/0` IP whitelist in prod
- ✗ Non commitare MONGODB_URI in code
- ✗ Non usare password deboli se usi auth password

---

## 10. Costi

### MongoDB Atlas M0 (Free)

| Feature | Limit |
|---------|-------|
| Storage | 512 MB |
| Throughput | Shared |
| Backup | 7 days |
| SLA | Community |
| Costo | **$0/mese** |

### Upgrade a Paid (M2+)

Se necessario scalare:

| Tier | RAM | Storage | Prezzo/mese |
|------|-----|---------|-------------|
| M2 | 2GB | 10GB | ~$9 |
| M5 | 4GB | 20GB | ~$57 |
| M10 | 10GB | 100GB | ~$96 |

---

## 11. Prossimi Passi

1. ✅ Crea cluster su MongoDB Atlas
2. ✅ Configura X.509 certificate
3. ✅ Aggiungi IP alla whitelist
4. ✅ Imposta environment variables
5. ✅ Esegui `MIGRATE_FROM_FILES=true go run main.go`
6. ✅ Verifica connessione nei logs
7. ✅ Controlla data migrata in Atlas UI

---

## Contatti Support

- **MongoDB Support**: https://support.mongodb.com
- **Docs**: https://docs.mongodb.com/drivers/go
- **Community**: https://community.mongodb.com
