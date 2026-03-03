# Piano Migrazione MongoDB Atlas

## 1. Analisi Architettura Corrente

### Storage Attuale
- **API Layer**: Map in memoria (NO persistenza)
  - `apiMenus` map[string]*Menu
  - `apiRestaurants` map[string]*Restaurant
- **Web Handlers Layer**: File JSON in cartella `storage/`
  - `restaurant_*.json` → Collections `restaurants`
  - `session_*.json` → Collections `sessions`
  - `menu_*.json` → Collections `menus`

### Problemi
1. ❌ Dati API non persistenti (persi al restart)
2. ❌ Due sistemi di storage separati (API ≠ Handlers)
3. ❌ Scalabilità limitata (file system)
4. ❌ Backup/ripristino manuale

### Vantaggi MongoDB Atlas
1. ✅ Persistenza garantita
2. ✅ ACID transactions
3. ✅ Scalabilità orizzontale
4. ✅ Backup automatici
5. ✅ Replica set per HA
6. ✅ Indicizzazione automatica
7. ✅ X.509 authentication

---

## 2. Architettura Nuova

```
┌─────────────────────────────────────────────────────┐
│              QR-Menu Application                    │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────────┐      ┌──────────────────┐    │
│  │  API Handlers    │      │  Web Handlers    │    │
│  │ (api/*.go)       │      │ (handlers/*.go)  │    │
│  └────────┬─────────┘      └────────┬─────────┘    │
│           │                         │               │
│           └────────────┬────────────┘               │
│                        │                           │
│                  ┌─────▼──────┐                    │
│                  │  DB Layer  │                    │
│                  │(db/mongo.go)│                    │
│                  └─────┬──────┘                    │
│                        │                           │
│         ┌──────────────┼──────────────┐            │
│         │              │              │            │
│    ┌────▼────┐    ┌───▼────┐    ┌───▼────┐       │
│    │ Menus   │    │Sessions│    │Restau  │       │
│    │         │    │        │    │rants   │       │
│    └────┬────┘    └───┬────┘    └───┬────┘       │
│         └──────────────┼──────────────┘            │
│                        │                          │
└────────────────────────┼──────────────────────────┘
                         │
                    ┌────▼──────┐
                    │  MongoDB   │
                    │  Atlas     │
                    │  Cluster   │
                    └───────────┘
```

---

## 3. Dettagli Implementazione

### 3.1 Database Package (`db/mongo.go`)

#### Struttura
```go
type MongoClient struct {
    client  *mongo.Client
    db      *mongo.Database
    ctx     context.Context
    cancel  context.CancelFunc
}

// Collections
- restaurants
- menus
- sessions
- audit_logs
- analytics_events
```

#### Funzioni Principali
- `Connect()` → Connessione con X.509
- `Disconnect()` → Cleanup
- `CreateUser()` → Salva restaurant
- `GetUser()` → Recupera restaurant
- `CreateSession()` → Salva session
- `CreateMenu()` → Salva menu
- ecc...

### 3.2 File da Modificare

#### API Layer (`api/*.go`)
- `api/auth.go` → Usa DB per autenticazione
- `api/menu.go` → Usa DB per menu CRUD
- `api/restaurant.go` → Usa DB per restaurant

#### Web Handlers (`handlers/*.go`)
- `handlers/auth.go` → Usa DB per sessioni/autenticazione
- `handlers/handlers.go` → Usa DB per menu

#### Main Package
- `main.go` → Inizializza connessione MongoDB

### 3.3 Configurazione

#### Environment Variables
```
MONGODB_URI=mongodb+srv://qr-menu-dev:[certificate]@cluster0.XXXX.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509
MONGODB_CERT_PATH=C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem
MONGODB_DB_NAME=qr-menu
```

#### Connection String con X.509
```
mongodb+srv://
  <username>:<password>@<cluster>.<region>.mongodb.net/
  ?authSource=$external
  &authMechanism=MONGODB-X509
  &retryWrites=true
```

---

## 4. Work Breakdown Structure

### Phase 1: Setup Infrastructure ✓
- [ ] Installare driver MongoDB (`go.mod`)
- [ ] Creare `db/mongo.go` con client wrapper
- [ ] Implementare `Connect()` con X.509
- [ ] Implementare `Disconnect()`

### Phase 2: Core Data Access ✓
- [ ] CRUD operations per Restaurants
- [ ] CRUD operations per Menus
- [ ] CRUD operations per Sessions
- [ ] CRUD operations per Audit Logs

### Phase 3: API Integration ✓
- [ ] Aggiornare `api/auth.go`
- [ ] Aggiornare `api/menu.go`
- [ ] Aggiornare `api/restaurant.go`
- [ ] Testare endpoint API

### Phase 4: Web Handlers Integration ✓
- [ ] Aggiornare `handlers/auth.go`
- [ ] Aggiornare `handlers/handlers.go`
- [ ] Testare web interface

### Phase 5: Testing & Migration ✓
- [ ] Migration script (file JSON → MongoDB)
- [ ] Testare persistenza
- [ ] Testare failover
- [ ] Performance testing

### Phase 6: Documentation ✓
- [ ] Update `README.md`
- [ ] Create `MONGODB_SETUP.md`
- [ ] Create `MIGRATION_GUIDE.md`
- [ ] Update deployment docs

---

## 5. Schema MongoDB

### Collection: restaurants
```json
{
  "_id": ObjectId,
  "id": "uuid",
  "username": "string",
  "email": "string",
  "password_hash": "string",
  "role": "admin|owner|staff",
  "name": "string",
  "description": "string",
  "address": "string",
  "phone": "string",
  "logo": "string",
  "active_menu_id": "string",
  "created_at": ISODate,
  "last_login": ISODate,
  "is_active": boolean
}
Indexes: {username: 1}, {email: 1}, {id: 1}
```

### Collection: menus
```json
{
  "_id": ObjectId,
  "id": "uuid",
  "restaurant_id": "uuid",
  "name": "string",
  "description": "string",
  "meal_type": "breakfast|lunch|dinner|generic",
  "categories": [
    {
      "id": "uuid",
      "name": "string",
      "description": "string",
      "items": [...]
    }
  ],
  "created_at": ISODate,
  "updated_at": ISODate,
  "is_completed": boolean,
  "is_active": boolean,
  "qr_code_path": "string",
  "public_url": "string"
}
Indexes: {restaurant_id: 1}, {id: 1}, {created_at: -1}
```

### Collection: sessions
```json
{
  "_id": ObjectId,
  "id": "uuid",
  "restaurant_id": "uuid",
  "created_at": ISODate,
  "last_accessed": ISODate,
  "ip_address": "string",
  "user_agent": "string"
}
Indexes: {restaurant_id: 1}, {id: 1}, {last_accessed: -1}
TTL: 604800 (7 days)
```

---

## 6. Timeline

| Fase | Durata | Status |
|------|--------|--------|
| 1. Infrastructure | 30 min | 🔄 In Progress |
| 2. Core Data Access | 1 ora | Pending |
| 3. API Integration | 1 ora | Pending |
| 4. Web Handlers | 45 min | Pending |
| 5. Testing | 1 ora | Pending |
| 6. Documentation | 30 min | Pending |
| **TOTALE** | **~5 ore** | |

---

## 7. Rollback Plan

Se qualcosa va storto:
1. Mantenere storage JSON come backup
2. Aggiungere environment variable `USE_FILE_STORAGE=true`
3. Fallback automatico ai file se MongoDB è down

---

## 8. Benefici Post-Migrazione

- 📈 **Scalabilità**: Replica set per alta disponibilità
- 🔄 **Sincronizzazione**: API e Web usano same source of truth
- 💾 **Persistenza**: Nessun dato perso al restart
- 🔐 **Backup**: Automatici ogni ora da Atlas
- 📊 **Analytics**: Query complesse su MongoDB
- 🚀 **Deployment**: Più semplice (no file system)
