# Fix Tag BSON - Problema Menu Non Visibili

## 🐛 Problema Identificato

Il problema dei menu non visibili era causato da un **mismatch nei nomi dei campi MongoDB**.

### Root Cause

I modelli Go (`models/menu.go`, `models/billing.go`, `models/webhook.go`) non avevano tag `bson`, quindi:

```go
// PRIMA (senza tag bson)
type Menu struct {
    ID           string    `json:"id"`
    RestaurantID string    `json:"restaurant_id"` 
    Name         string    `json:"name"`
    // ...
}
```

Quando MongoDB salvava i documenti:
- Usava i **nomi dei campi Go** (es. `RestaurantID` con R maiuscola)

Quando MongoDB faceva query:
```go
coll.Find(ctx, bson.M{"restaurant_id": restaurantID})
```
- Cercava `restaurant_id` (minuscolo con underscore)
- **NON TROVAVA NULLA** perché il campo era salvato come `RestaurantID`!

### Risultato

✅ **POST /api/v1/menus** → 200 OK (menu salvato)  
❌ **GET /api/v1/menus** → 200 OK ma array vuoto (query non trova niente)

---

## ✅ Soluzione Applicata

Aggiunti tag `bson` a tutti i modelli per garantire consistenza:

```go
// DOPO (con tag bson)
type Menu struct {
    ID           string    `json:"id" bson:"id"`
    RestaurantID string    `json:"restaurant_id" bson:"restaurant_id"` 
    Name         string    `json:"name" bson:"name"`
    // ...
}
```

**Commit:** [`84295b3`](https://github.com/gigliofr/qr-menu/commit/84295b3)

---

## 🔄 Migrazione Dati Esistenti

⚠️ **IMPORTANTE:** I documenti GIÀ PRESENTI in MongoDB hanno i vecchi nomi di campo (es. `RestaurantID`), quindi NON saranno visibili con il nuovo codice.

### Opzione 1: Elimina e Ricrea (RACCOMANDATO per staging)

1. Vai su [MongoDB Atlas](https://cloud.mongodb.com)
2. Seleziona il cluster: `ac-d8zdak4.b9jfwmr.mongodb.net`
3. Clicca su **Browse Collections**
4. Database: `qr-menu`
5. Elimina le collections:
   - `restaurants`
   - `menus`
   - `sessions`
   - `analytics_events` (opzionale)
   - `audit_logs` (opzionale)
6. Esegui lo script di setup:
   ```powershell
   .\setup_ristorante_completo.ps1
   ```

### Opzione 2: Usa lo Script di Migrazione

```powershell
.\migrate_bson_fields.ps1
```

Lo script ti guiderà attraverso il processo manuale di eliminazione delle collections.

---

## 🧪 Test Post-Fix

Dopo la migrazione/eliminazione, verifica che tutto funzioni:

```powershell
# 1. Crea nuovo account e menu
.\setup_ristorante_completo.ps1

# 2. Verifica che GET /menus restituisca i dati
$token = "..." # Token dalla risposta login
$headers = @{"Authorization" = "Bearer $token"}
Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/api/v1/menus" -Headers $headers
```

**Output atteso:**
```json
{
  "success": true,
  "data": [
    {
      "id": "...",
      "restaurant_id": "...",
      "name": "Menu Pranzo",
      "categories": [...]
    }
  ]
}
```

---

## 📁 File Modificati

- ✅ `models/menu.go` - Aggiunti tag bson a Menu, Restaurant, Session, MenuItem, MenuCategory
- ✅ `models/billing.go` - Aggiunti tag bson a BillingPlan, BillingSubscription  
- ✅ `models/webhook.go` - Aggiunti tag bson a WebhookEndpoint, WebhookDelivery, WebhookEvent

---

## 🚀 Deploy Status

- **Commit:** 84295b3
- **Branch:** main  
- **Railway:** Auto-deploy in corso...
- **URL:** https://qr-menu-staging.up.railway.app

Una volta completato il deploy Railway, segui i passi di migrazione sopra.

---

## 🔍 Verifiche Complete

### 1. Health Check
```bash
curl https://qr-menu-staging.up.railway.app/health
```
Expected: `"database": "connected"`

### 2. Registrazione
```powershell
$body = @{
    username = "test_user"
    email = "test@example.com"
    password = "Test123!"
    restaurant_name = "Test Restaurant"
} | ConvertTo-Json

Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/api/v1/auth/register" -Method Post -Body $body -ContentType "application/json"
```

### 3. Login
```powershell
$body = @{username="test_user"; password="Test123!"} | ConvertTo-Json
$resp = Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/api/v1/auth/login" -Method Post -Body $body -ContentType "application/json"
$token = $resp.data.token
```

### 4. Crea Menu
```powershell
$headers = @{"Authorization" = "Bearer $token"}
$menu = @{
    name = "Test Menu"
    description = "Menu di test"
    meal_type = "lunch"
    categories = @(
        @{
            id = (New-Guid).ToString()
            name = "Antipasti"
            description = "Stuzzichini"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Bruschetta"
                    description = "Pane tostato con pomodoro"
                    price = 5.50
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/api/v1/menus" -Method Post -Body $menu -Headers $headers -ContentType "application/json"
```

### 5. Recupera Menu (QUESTO ERA IL BUG!)
```powershell
$menus = Invoke-RestMethod -Uri "https://qr-menu-staging.up.railway.app/api/v1/menus" -Headers $headers
$menus.data.Count  # Deve essere > 0!
```

Se `$menus.data.Count` è maggiore di 0, il fix è FUNZIONANTE ✅

---

## 📝 Note Tecniche

### Perché MongoDB non usava automaticamente i tag json?

MongoDB driver Go (go.mongodb.org/mongo-driver) usa **solo** i tag `bson` per la serializzazione. I tag `json` sono ignorati completamente.

### Cosa succede senza tag bson?

Il driver usa reflection per ottenere i nomi dei campi Go struct (es. `RestaurantID`), che possono essere diversi dai nomi JSON (`restaurant_id`).

### Perché le POST funzionavano ma le GET no?

- **POST (Insert):** Salva il documento con i nomi campo Go struct
- **GET (Find):** Cerca con i nomi hardcoded nella query (`restaurant_id`)
- **Mismatch:** I due nomi non corrispondono → query vuota

---

## ✅ Checklist Post-Fix

- [ ] Railway deploy completato
- [ ] MongoDB collections eliminate o migrate  
- [ ] Nuovo account registrato via API
- [ ] Menu creato via API
- [ ] Menu visibile in GET /api/v1/menus ✨
- [ ] Login UI funziona
- [ ] Dashboard mostra i menu ✨
- [ ] QR code generabile
- [ ] Analytics e audit logs popolati

---

**Data Fix:** 2025-01-XX  
**Commit:** 84295b3  
**Status:** ✅ RISOLTO (in attesa di migrazione dati)
