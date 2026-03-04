# 📋 Guida Migrazione Multi-Ristorante

## ⚠️ IMPORTANTE: Leggere Prima di Procedere

Questa migrazione converte il database dal vecchio modello (Restaurant con autenticazione incorporata) al nuovo modello (User + Restaurant separati) per supportare utenti con più ristoranti.

## 🔍 Pre-requisiti

1. **Backup del database**
2. **MongoDB Shell (mongosh) installato**
3. **Accesso al database MongoDB Atlas**
4. **Credenziali di connessione**

## 📦 Cosa Fa la Migrazione

### Conversione Dati
```
PRIMA (Restaurant monolitico):
Restaurant {
    _id: "abc123"
    username: "mario"
    email: "mario@test.com"
    password_hash: "..."
    name: "Pizzeria Mario"
    address: "Via Roma 1"
}

DOPO (User + Restaurant separati):
User {
    _id: "user_789"
    username: "mario"
    email: "mario@test.com"
    password_hash: "..."
    privacy_consent: true
    marketing_consent: false
}

Restaurant {
    _id: "abc123"
    owner_id: "user_789"  ← Link al User
    name: "Pizzeria Mario"
    address: "Via Roma 1"
}
```

### Modifiche Database

1. **Crea collezione `users`**
   - Estrae username, email, password_hash da restaurants
   - Aggiunge campi GDPR (privacy_consent, marketing_consent)
   - Crea indici unique su username e email

2. **Modifica collezione `restaurants`**
   - Aggiunge campo `owner_id` → User._id
   - Rimuove campi auth (username, email, password_hash, role, last_login)

3. **Aggiorna collezione `sessions`**
   - Aggiunge campo `user_id` a tutte le sessioni esistenti

4. **Crea indici performance**
   - users: username (unique), email (unique), is_active
   - restaurants: owner_id + is_active, owner_id + created_at
   - sessions: user_id + last_accessed, TTL index (30 giorni)

## 🚀 Procedura Step-by-Step

### Step 1: Backup Database

```bash
# Backup completo
mongodump --uri="mongodb+srv://USER:PASSWORD@cluster.mongodb.net/qr-menu" \
          --out=backup_$(date +%Y%m%d_%H%M%S)

# Verifica backup
ls -lh backup_*
```

### Step 2: Test su Database di Sviluppo (CONSIGLIATO)

```bash
# 1. Crea database di test
mongosh "mongodb+srv://..." --eval "use('qr-menu-test')"

# 2. Copia dati produzione
mongodump --uri="mongodb+srv://.../qr-menu" --out=temp_backup
mongorestore --uri="mongodb+srv://.../qr-menu-test" temp_backup/qr-menu

# 3. Esegui migrazione su test
mongosh "mongodb+srv://.../qr-menu-test" --file scripts/migrate_user_restaurant.js

# 4. Verifica risultati
mongosh "mongodb+srv://.../qr-menu-test"
```

In mongosh:
```javascript
use("qr-menu-test");

// Verifica users creati
db.users.countDocuments();
db.users.findOne();

// Verifica restaurants aggiornati
db.restaurants.findOne();  // Deve avere owner_id, NON username/email

// Verifica link corretti
const user = db.users.findOne();
db.restaurants.findOne({ owner_id: user._id });  // Deve restituire il ristorante
```

### Step 3: Esegui Migrazione su Produzione

```bash
# Connetti a MongoDB Atlas
mongosh "mongodb+srv://USERNAME:PASSWORD@cluster.mongodb.net/qr-menu"

# Esegui script migrazione
mongosh "mongodb+srv://USERNAME:PASSWORD@cluster.mongodb.net/qr-menu" \
        --file scripts/migrate_user_restaurant.js > migration_log_$(date +%Y%m%d_%H%M%S).txt

# Controlla il log
cat migration_log_*.txt
```

### Step 4: Verifica Post-Migrazione

In mongosh:
```javascript
use("qr-menu");

// ✅ Verifica 1: Numero users = numero restaurants
db.users.countDocuments();        // Es: 15
db.restaurants.countDocuments();  // Es: 15

// ✅ Verifica 2: Tutti i restaurants hanno owner_id
db.restaurants.countDocuments({ owner_id: { $exists: true, $ne: "" } });
// Deve essere uguale al totale restaurants

// ✅ Verifica 3: Nessun restaurant ha campi auth
db.restaurants.countDocuments({ username: { $exists: true } });
// Deve essere 0

// ✅ Verifica 4: Test login di un utente
const user = db.users.findOne({ username: "mario" });
print(user.email);  // Deve mostrare l'email
print(user.privacy_consent);  // Deve essere true

// ✅ Verifica 5: Test link User → Restaurants
const userRestaurants = db.restaurants.find({ owner_id: user._id }).toArray();
print(`User ${user.username} ha ${userRestaurants.length} ristoranti`);

// ✅ Verifica 6: Indici creati
db.users.getIndexes();
db.restaurants.getIndexes();
db.sessions.getIndexes();
```

### Step 5: Test Applicazione

1. **Avvia applicazione**
   ```bash
   cd C:\Users\gigli\GoWs\qr-menu
   go run main.go
   ```

2. **Test Login**
   - Vai a http://localhost:8080/login
   - Login con credenziali esistenti
   - Verifica redirect corretto (se 1 ristorante → /admin, se >1 → /select-restaurant)

3. **Test Registrazione**
   - Vai a http://localhost:8080/register
   - Crea nuovo account
   - Verifica checkbox GDPR presenti
   - Verifica creazione User + Restaurant in MongoDB

4. **Test Multi-Ristorante**
   - Login con account esistente
   - Vai a /add-restaurant
   - Aggiungi secondo ristorante
   - Logout e re-login
   - Verifica redirect a /select-restaurant
   - Seleziona ristorante 1 → verifica menu isolati
   - Switch a ristorante 2 → verifica menu diversi

## 🛡️ Rollback (In Caso di Problemi)

### Rollback Immediato

```bash
# Ripristina dal backup
mongorestore --uri="mongodb+srv://..." \
             --drop \
             backup_YYYYMMDD_HHMMSS
```

### Rollback Parziale (Solo Dati)

```javascript
use("qr-menu");

// 1. Elimina collezione users
db.users.drop();

// 2. Per ogni restaurant, ripristina campi auth (se hai backup)
// Questo è complicato, meglio fare restore completo
```

## 📊 Query Utili Post-Migrazione

### Dashboard Admin

```javascript
// Statistiche sistema
db.users.countDocuments({ is_active: true });
db.restaurants.countDocuments({ is_active: true });
db.sessions.countDocuments({ last_accessed: { $gte: new Date(Date.now() - 86400000) } });

// Utenti con più ristoranti
db.restaurants.aggregate([
    { $group: { 
        _id: "$owner_id", 
        count: { $sum: 1 },
        restaurants: { $push: "$name" }
    }},
    { $match: { count: { $gt: 1 } }},
    { $sort: { count: -1 }}
]);

// Top 10 utenti per numero ristoranti
db.restaurants.aggregate([
    { $group: { _id: "$owner_id", count: { $sum: 1 } }},
    { $sort: { count: -1 }},
    { $limit: 10 },
    { $lookup: {
        from: "users",
        localField: "_id",
        foreignField: "_id",
        as: "user"
    }},
    { $unwind: "$user" },
    { $project: { 
        username: "$user.username", 
        email: "$user.email", 
        restaurant_count: "$count" 
    }}
]);
```

### GDPR Compliance

```javascript
// Utenti che hanno dato consenso marketing
db.users.countDocuments({ marketing_consent: true });

// Utenti senza consenso marketing (possono ricevere solo comunicazioni tecniche)
db.users.countDocuments({ marketing_consent: false });

// Audit trail consensi (ultimi 30 giorni)
db.users.find({ 
    consent_date: { $gte: new Date(Date.now() - 2592000000) }
}).sort({ consent_date: -1 });
```

### Performance Monitoring

```javascript
// Query lente (enable profiling first)
db.setProfilingLevel(1, { slowms: 100 });
db.system.profile.find({ millis: { $gt: 100 }}).sort({ millis: -1 }).limit(10);

// Statistiche indici
db.users.stats().indexSizes;
db.restaurants.stats().indexSizes;
db.sessions.stats().indexSizes;
```

## ❓ Troubleshooting

### Errore: "username già esistente"

```javascript
// Trova duplicati
db.restaurants.aggregate([
    { $group: { _id: "$username", count: { $sum: 1 } }},
    { $match: { count: { $gt: 1 } }}
]);

// Rinomina username duplicati PRIMA della migrazione
db.restaurants.updateOne(
    { username: "mario", _id: "OLD_ID" },
    { $set: { username: "mario_old" } }
);
```

### Errore: "Alcuni restaurant senza owner_id"

```javascript
// Trova restaurant orfani
db.restaurants.find({ owner_id: { $exists: false } });

// Fix manuale (crea user temporaneo)
const tempUserID = new ObjectId().toString();
db.users.insertOne({
    _id: tempUserID,
    username: "orphan_owner",
    email: "orphan@system.local",
    password_hash: "",
    privacy_consent: true,
    marketing_consent: false
});

db.restaurants.updateMany(
    { owner_id: { $exists: false } },
    { $set: { owner_id: tempUserID } }
);
```

### Performance Degradation

```javascript
// Ricostruisci indici
db.users.reIndex();
db.restaurants.reIndex();
db.sessions.reIndex();

// Verifica query plan
db.restaurants.find({ owner_id: "user_123" }).explain("executionStats");
```

## 📞 Supporto

In caso di problemi:
1. Controlla il log della migrazione (`migration_log_*.txt`)
2. Verifica il database con le query di verifica sopra
3. Consulta la documentazione tecnica in ARCHITECTURE.md
4. Ripristina dal backup se necessario

## ✅ Checklist Finale

- [ ] Backup database completato
- [ ] Migrazione testata su database di sviluppo
- [ ] Script migrazione eseguito su produzione
- [ ] Verifiche post-migrazione TUTTE ✅
- [ ] Test login funzionante
- [ ] Test registrazione con GDPR funzionante
- [ ] Test multi-ristorante funzionante
- [ ] Indici database creati
- [ ] Performance monitorate
- [ ] Backup vecchio conservato per 30 giorni

## 📅 Timeline Consigliata

1. **Giorno 1**: Test su DB di sviluppo, fix eventuali problemi
2. **Giorno 2-3**: Migrazione in produzione durante orario di basso traffico (es: 2-4 AM)
3. **Giorno 4-7**: Monitoraggio intensivo, rollback immediato se problemi critici
4. **Giorno 30**: Elimina backup se tutto OK
