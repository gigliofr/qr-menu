# MongoDB Integration Implementation Guide

## Panoramica

Questo documento spiega come è stata integrata la persistenza MongoDB nell'applicazione qr-menu e come usare il database layer.

## Architettura

```
┌─────────────────────────────┐
│   API Handlers / Web UI      │
│  (api/*.go, handlers/*.go)   │
└──────────────┬──────────────┘
               │ Chiama
               ▼
┌─────────────────────────────┐
│   Database Layer            │
│   (db/mongo.go)             │
│                             │
│   - CreateRestaurant()      │
│   - GetRestaurantByID()     │
│   - CreateMenu()            │
│   - GetMenuByID()           │
│   - CreateSession()         │
│   - GetSessionByID()        │
│   ... (altre operazioni)    │
└──────────────┬──────────────┘
               │ Usa
               ▼
┌─────────────────────────────┐
│  MongoDB Atlas              │
│  (Cloud Database)           │
│                             │
│  Collections:               │
│  - restaurants              │
│  - menus                    │
│  - sessions                 │
└─────────────────────────────┘
```

## File Nuovi

### `db/mongo.go` - MongoDB Client Wrapper
Contiene:
- **MongoClient struct**: Wrapper per il client MongoDB
- **Connect()**: Stabilisce connessione con X.509
- **Disconnect()**: Chiude la connessione
- **CRUD methods**: Per restaurants, menus, sessions
- **createIndexes()**: Crea indici MongoDB

### `db/mongo_migration.go` - Data Migration
Contiene:
- **MigrateFromFileStorage()**: Migra da JSON file a MongoDB
- **migrateRestaurants()**: Migra ristoranti
- **migrateMenus()**: Migra menu
- **migrateSessions()**: Migra sessioni
- **BackupToJSON()**: Esporta da MongoDB a JSON

## Come Usare nel Codice

### Accesso al Database

```go
import "qr-menu/db"

// Nel tuo handler
func MyHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // Leggi un ristorante da MongoDB
    restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, restaurantID)
    if err != nil {
        http.Error(w, "Database error", 500)
        return
    }
    
    // Modifica
    restaurant.LastLogin = time.Now()
    
    // Salva su MongoDB
    if err := db.MongoInstance.UpdateRestaurant(ctx, restaurant); err != nil {
        http.Error(w, "Cannot save", 500)
        return
    }
}
```

### Operazioni CRUD - Restaurants

```go
ctx := context.Background()

// CREATE
restaurant := &models.Restaurant{
    ID: uuid.New().String(),
    Username: "owner1",
    Email: "owner1@example.com",
    // ... altri campi
}
err := db.MongoInstance.CreateRestaurant(ctx, restaurant)

// READ
restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, id)
restaurant, err := db.MongoInstance.GetRestaurantByUsername(ctx, "owner1")
restaurant, err := db.MongoInstance.GetRestaurantByEmail(ctx, "owner1@example.com")

// UPDATE
restaurant.Name = "New Name"
err := db.MongoInstance.UpdateRestaurant(ctx, restaurant)

// READ ALL
allRestaurants, err := db.MongoInstance.GetAllRestaurants(ctx)
```

### Operazioni CRUD - Menus

```go
ctx := context.Background()

// CREATE
menu := &models.Menu{
    ID: uuid.New().String(),
    RestaurantID: restaurantID,
    Name: "Lunch Menu",
    MealType: "lunch",
    // ...
}
err := db.MongoInstance.CreateMenu(ctx, menu)

// READ
menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)
menus, err := db.MongoInstance.GetMenusByRestaurantID(ctx, restaurantID)

// UPDATE
menu.Name = "Updated Name"
err := db.MongoInstance.UpdateMenu(ctx, menu)

// DELETE
err := db.MongoInstance.DeleteMenu(ctx, menuID)
```

### Operazioni CRUD - Sessions

```go
ctx := context.Background()

// CREATE
session := &models.Session{
    ID: uuid.New().String(),
    RestaurantID: restaurantID,
    CreatedAt: time.Now(),
    LastAccessed: time.Now(),
    IPAddress: r.RemoteAddr,
    UserAgent: r.UserAgent(),
}
err := db.MongoInstance.CreateSession(ctx, session)

// READ
session, err := db.MongoInstance.GetSessionByID(ctx, sessionID)
sessions, err := db.MongoInstance.GetSessionsByRestaurantID(ctx, restaurantID)

// UPDATE
session.LastAccessed = time.Now()
err := db.MongoInstance.UpdateSession(ctx, session)

// DELETE
err := db.MongoInstance.DeleteSession(ctx, sessionID)

// DELETE EXPIRED (>24h)
err := db.MongoInstance.DeleteExpiredSessions(ctx)
```

## Migrare API Handlers (Esempio: api/menu.go)

### Prima (in-memory):
```go
var apiMenus = make(map[string]*models.Menu)

func CreateMenuHandler(w http.ResponseWriter, r *http.Request) {
    // ...
    apiMenus[menu.ID] = menu  // ❌ Non persistente!
}

func GetMenuHandler(w http.ResponseWriter, r *http.Request) {
    menu, exists := apiMenus[menuID]  // ❌ In memoria
}
```

### Dopo (MongoDB):
```go
import "qr-menu/db"

func CreateMenuHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    // ...
    err := db.MongoInstance.CreateMenu(ctx, &menu)  // ✅ Persistente!
    if err != nil {
        ErrorResponse(w, 500, "DB_ERROR", "Cannot save menu", err.Error())
        return
    }
}

func GetMenuHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    menu, err := db.MongoInstance.GetMenuByID(ctx, menuID)  // ✅ Da MongoDB
    if err != nil {
        ErrorResponse(w, 500, "DB_ERROR", "Cannot fetch menu", err.Error())
        return
    }
}
```

## Migrare Web Handlers (Esempio: handlers/auth.go)

### Prima (file JSON):
```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    for _, rest := range restaurants {  // ❌ Map in memoria
        if rest.Username == username {
            // ...
        }
    }
}

func loadRestaurantsFromStorage() {
    // ❌ Legge file JSON da disk
    files, _ := filepath.Glob("storage/restaurant_*.json")
}
```

### Dopo (MongoDB):
```go
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    
    // ✅ Legge da MongoDB
    restaurant, err := db.MongoInstance.GetRestaurantByUsername(ctx, username)
    if err != nil {
        // Handle error
        return
    }
    if restaurant == nil {
        // User not found
        return
    }
    
    // Verifica password, crea sessione, etc.
}

// No more need for loadRestaurantsFromStorage()!
// MongoDB è il source of truth
```

## Context Handling

MongoDB Go driver richiede `context.Context`. Best practices:

```go
// ✅ BUONO: Timeout per operazioni
func MyHandler(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, id)
}

// ❌ CATTIVO: Nessun timeout
func MyHandler(w http.ResponseWriter, r *http.Request) {
    ctx := context.Background()
    // Potrebbe stare in attesa indefinitamente se MongoDB è offline
    restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, id)
}

// ✅ BUONO: Usa il context della request HTTP
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // r.Context() ha già il timeout impostato dal server HTTP
    restaurant, err := db.MongoInstance.GetRestaurantByID(r.Context(), id)
}
```

## Error Handling

### Errori Comuni

```go
import "go.mongodb.org/mongo-driver/mongo"

// Documento non trovato
restaurant, err := db.MongoInstance.GetRestaurantByID(ctx, id)
if err != nil {
    log.Printf("Database error: %v", err)
    http.Error(w, "Server error", 500)
    return
}
if restaurant == nil {
    // Record not found - NOT an error, just empty result
    http.Error(w, "Not found", 404)
    return
}

// Violazione di unique index
err := db.MongoInstance.CreateRestaurant(ctx, restaurant)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key error") {
        // Username o email già esiste
        http.Error(w, "Username already taken", 400)
        return
    }
    // Altro errore database
    http.Error(w, "Database error", 500)
    return
}
```

## Performance Tips

1. **Usa indici**: Sono creati automaticamente da `createIndexes()`
2. **Limita query**: Usa `GetMenusByRestaurantID()` invece di `GetAllMenus()` se possibile
3. **Batch operations**: Per migrazioni, usa batch insert se disponibile nel driver
4. **Connection pooling**: Automatico in MongoDB Go driver
5. **TTL Indexes**: Sessions auto-delete dopo 7 giorni

## Testing

```go
// Test con mock MongoDB (optional)
import "go.mongodb.org/mongo-driver/mongo/mocks"

func TestCreateRestaurant(t *testing.T) {
    // Puoi fare unit test senza MongoDB usando mock
    // O usare MongoDB Atlas test tier
}
```

## Troubleshooting

### "connection refused"
- Verifica che MongoDB Atlas cluster sia online
- Controlla IP whitelist in Network Access
- Assicurati certificato X.509 sia valido

### "unauthorized" error
- Controll che il certificato sia nella posizione corretta
- Verifica MONGODB_CERT_PATH environment variable

### "timeout" error
- Aumento timeout in contex (vedi esempi sopra)
- Controlla MongoDB Atlas status

### "database or collection not found"
- Non è un errore! MongoDB crea automaticamente
- Esegui `createIndexes()` per creare indici

## Prossimi Step

1. ✅ Completare migrazione di `api/menu.go` (usare db.MongoInstance)
2. ✅ Completare migrazione di `api/auth.go` (usare db.MongoInstance)
3. ✅ Completare migrazione di `handlers/auth.go` (usare db.MongoInstance)
4. ✅ Completare migrazione di `handlers/handlers.go` (usare db.MongoInstance)
5. ✅ Rimuovere codice vecchio di file storage
6. ✅ Testare persistenza end-to-end
7. ✅ Setup backup automatici in Atlas

## Reference

- MongoDB Go Driver: https://pkg.go.dev/go.mongodb.org/mongo-driver
- MongoDB Atlas: https://mongodb.com/cloud/atlas
- BSON Types: https://pkg.go.dev/go.mongodb.org/mongo-driver/bson
