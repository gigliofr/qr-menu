# 🌱 Scripts di Database - Documentazione

## 📋 Indice

1. [seed_test_data.js](#seed_test_datajs) - Script MongoDB per creare dati di test
2. [setup_and_seed.ps1](#setup_and_seedps1) - Helper PowerShell per setup automatico
3. [migrate_user_restaurant.js](#migrate_user_restaurantjs) - Migrazione produzione

---

## 🧹 seed_test_data.js

**Scopo:** Pulisce completamente il database e crea dati di test per sviluppo.

### Cosa Fa

1. **Pulizia Totale:**
   - Elimina tutti i documenti da: users, restaurants, sessions, menus, analytics_events

2. **Crea 1 Utente Amministrativo:**
   - Username: `admin`
   - Password: `admin`
   - Email: `admin@qrmenu.local`
   - GDPR consents: Privacy ✅, Marketing ❌

3. **Crea 4 Ristoranti di Test:**

   | # | Nome | Tipo | Indirizzo | Piatti |
   |---|------|------|-----------|--------|
   | 1 | Pizzeria Napoletana | Pizza | Via Roma 123, Napoli | 10 |
   | 2 | Trattoria Toscana | Toscana | Piazza Duomo 45, Firenze | 7 |
   | 3 | Sushi-Ya Tokyo | Giapponese | Corso Venezia 88, Milano | 10 |
   | 4 | Burger House Americana | American | Via Garibaldi 77, Roma | 10 |

4. **Crea Menu Completi:**
   - **Pizzeria:** Margherita, Marinara, Diavola, 4 Stagioni, Bufala, Tartufo + Bevande
   - **Trattoria:** Crostini, Tagliere, Pici Cacio e Pepe, Pappardelle, Ribollita, Bistecca, Tagliata
   - **Sushi:** Nigiri (Sake, Maguro, Ebi, Unagi), Maki (California, Spicy Tuna, Dragon), Sashimi
   - **Burger:** Classic, Cheeseburger, Bacon, Truffle, Mexican, Veggie + Contorni

5. **Verifica Indici:**
   - Crea tutti gli indici per performance (se non esistono)

### Uso Manuale

```bash
# Se hai mongosh installato e MongoDB configurato:
mongosh "mongodb+srv://your-cluster.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509" \
  --tls \
  --tlsCertificateKeyFile "path/to/cert.pem" \
  --file scripts/seed_test_data.js
```

### Output Atteso

```
================================================
🧹 PULIZIA DATABASE
================================================

📊 Conteggio documenti PRIMA della pulizia:
   Users:            5
   Restaurants:      8
   Sessions:         12
   Menus:            7
   Analytics Events: 150

✅ Collezioni pulite

================================================
👤 CREAZIONE UTENTE AMMINISTRATIVO
================================================

✅ Utente admin creato:
   Username: admin
   Password: admin
   User ID:  admin_user_001

================================================
🏪 CREAZIONE 4 RISTORANTI DI TEST
================================================

✅ Ristoranti creati: 4
   📍 Pizzeria Napoletana (ID: rest_001)
   📍 Trattoria Toscana (ID: rest_002)
   📍 Sushi-Ya Tokyo (ID: rest_003)
   📍 Burger House Americana (ID: rest_004)

================================================
📋 CREAZIONE MENU PER OGNI RISTORANTE
================================================

✅ Menu creati: 4
✅ Menu attivati per ogni ristorante

   📋 Menu Pizzeria - Primavera 2026 → 3 sezioni, 10 piatti
   📋 Menu Toscano - Stagione → 3 sezioni, 7 piatti
   📋 Menu Sushi - Stagione 2026 → 3 sezioni, 10 piatti
   📋 Menu Burger 2026 → 3 sezioni, 10 piatti

================================================
📊 STATISTICHE FINALI
================================================

✅ Database popolato con successo!

👥 Users:            1
🏪 Restaurants:      4
🔑 Sessions:         0
📋 Menus:            4
📊 Analytics Events: 0

================================================
🎉 SEED COMPLETATO CON SUCCESSO!
================================================

🔐 CREDENZIALI AMMINISTRATORE:
   Username: admin
   Password: admin
   Email:    admin@qrmenu.local

🏪 RISTORANTI CREATI:
   1. Pizzeria Napoletana    (10 piatti)
   2. Trattoria Toscana      (7 piatti)
   3. Sushi-Ya Tokyo         (10 piatti)
   4. Burger House Americana (10 piatti)

🚀 Accedi a: http://localhost:8080/login
```

---

## 🚀 setup_and_seed.ps1

**Scopo:** Script helper PowerShell che semplifica il setup e l'esecuzione del seed.

### Cosa Fa

1. **Verifica Configurazione MongoDB:**
   - Controlla variabili d'ambiente: MONGODB_URI, MONGODB_CERT_PATH, MONGODB_DB_NAME
   - Se mancanti, guida l'utente nella configurazione

2. **Setup Interattivo:**
   - Chiede credenziali MongoDB se non configurate
   - Permette setup temporaneo (solo sessione) o permanente
   - Verifica esistenza del certificato X509

3. **Verifica mongosh:**
   - Controlla che MongoDB Shell sia installato
   - Fornisce istruzioni se mancante

4. **Esegue Seed:**
   - Chiede conferma prima di pulire il database
   - Esegue `seed_test_data.js` con le credenziali corrette
   - Mostra output colorato e formattato

5. **Post-Seed:**
   - Mostra credenziali admin
   - Lista ristoranti creati
   - Fornisce prossimi passi

### Uso

```powershell
# Esecuzione semplice
.\scripts\setup_and_seed.ps1

# Lo script ti guiderà attraverso:
# 1. Configurazione MongoDB (se necessaria)
# 2. Verifica mongosh
# 3. Conferma seed
# 4. Esecuzione e risultati
```

### Pre-requisiti

1. **MongoDB Shell (mongosh):**
   ```powershell
   # Installa con Chocolatey:
   choco install mongodb-shell
   
   # Oppure scarica da:
   # https://www.mongodb.com/try/download/shell
   ```

2. **Certificato X509:**
   - Scarica da MongoDB Atlas → Database Access → Certificates
   - Salva come `cert.pem` in un percorso sicuro

3. **Credenziali MongoDB Atlas:**
   - Connection String (URI) con X509 authentication
   - Formato: `mongodb+srv://cluster.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509`

### Opzioni di Setup

#### Opzione A: Setup Temporaneo (solo questa sessione PowerShell)

```powershell
$env:MONGODB_URI="mongodb+srv://cluster.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509"
$env:MONGODB_CERT_PATH="C:\path\to\X509-cert.pem"
$env:MONGODB_DB_NAME="qr-menu"

# Poi esegui
.\scripts\setup_and_seed.ps1
```

#### Opzione B: Setup Permanente (tutte le sessioni)

```powershell
[System.Environment]::SetEnvironmentVariable("MONGODB_URI", "mongodb+srv://...", "User")
[System.Environment]::SetEnvironmentVariable("MONGODB_CERT_PATH", "C:\path\to\cert.pem", "User")
[System.Environment]::SetEnvironmentVariable("MONGODB_DB_NAME", "qr-menu", "User")

# Riavvia terminal, poi
.\scripts\setup_and_seed.ps1
```

#### Opzione C: Setup Interattivo (lo script ti guida)

```powershell
# Esegui senza variabili, lo script chiederà tutto
.\scripts\setup_and_seed.ps1

# Rispondi 's' a "Vuoi configurare ora?"
# Inserisci: URI, path certificato, nome database
# Scegli se salvare in modo permanente
```

---

## 🔄 migrate_user_restaurant.js

**Scopo:** Migrazione dati produzione da vecchio modello al nuovo modello User/Restaurant.

**Documentazione Completa:** Vedi [MIGRATION_GUIDE.md](../MIGRATION_GUIDE.md)

### Differenza con seed_test_data.js

| Aspetto | seed_test_data.js | migrate_user_restaurant.js |
|---------|-------------------|----------------------------|
| Scopo | Dati di test sviluppo | Migrazione produzione |
| Pulizia | Elimina TUTTO | Converte dati esistenti |
| Dati | Crea dati fake | Preserva dati reali |
| Idempotente | No (distruttivo) | Sì (skip già migrati) |
| Rollback | Nessuno | Via mongorestore |
| Uso | Dev/Testing | Una volta in produzione |

### Quando Usare Quale

**Usa `seed_test_data.js` quando:**
- ✅ Stai sviluppando localmente
- ✅ Vuoi dati di test puliti
- ✅ Non ti importa di perdere dati esistenti
- ✅ Vuoi testare con 4 ristoranti e 37 piatti predefiniti

**Usa `migrate_user_restaurant.js` quando:**
- ✅ Hai dati REALI in produzione
- ✅ Vuoi preservare ristoranti e menu esistenti
- ✅ Stai migrando da vecchio modello (Restaurant con auth) a nuovo (User + Restaurant)
- ✅ Hai fatto backup del database

---

## 🧪 Testing Workflow Completo

### 1. Setup Ambiente di Sviluppo

```powershell
# Clone repo
git clone https://github.com/gigliofr/qr-menu.git
cd qr-menu

# Compila applicazione
go build -o qr-menu.exe .

# Configura MongoDB (setup temporaneo per test)
$env:MONGODB_URI="mongodb+srv://dev-cluster.mongodb.net/..."
$env:MONGODB_CERT_PATH="C:\dev\mongodb-cert.pem"
$env:MONGODB_DB_NAME="qr-menu-dev"
```

### 2. Seed Database di Test

```powershell
# Esegui seed (pulisce e ricrea)
.\scripts\setup_and_seed.ps1

# Oppure manuale:
mongosh "$env:MONGODB_URI" `
  --tls `
  --tlsCertificateKeyFile "$env:MONGODB_CERT_PATH" `
  --file scripts\seed_test_data.js
```

### 3. Avvia Applicazione

```powershell
.\qr-menu.exe

# Output atteso:
# 🔄 Connessione a MongoDB Atlas...
# ✅ MongoDB connesso
# 🚀 Server avviato su http://localhost:8080
```

### 4. Test nel Browser

```
http://localhost:8080/login

Username: admin
Password: admin

→ Redirect a /select-restaurant
→ Scegli "Pizzeria Napoletana"
→ Vai su /admin
→ Vedi il menu con 10 pizze
```

### 5. Test Multi-Ristorante

```
/select-restaurant → Scegli "Trattoria Toscana"
→ /admin → Menu cambiato (7 piatti toscani)

/select-restaurant → Scegli "Sushi-Ya Tokyo"
→ /admin → Menu giapponese (10 piatti)

/select-restaurant → Scegli "Burger House"
→ /admin → Menu americano (10 burger)
```

### 6. Test Isolamento Dati

```
Crea nuovo menu in "Pizzeria"
→ Switch a "Trattoria"
→ Verifica che menu di Pizzeria NON sia visibile ✅

Crea piatto in "Sushi-Ya"
→ Switch a "Burger House"
→ Verifica isolamento ✅
```

### 7. Re-seed per Test Puliti

```powershell
# Quando vuoi ricominciare da zero
.\scripts\setup_and_seed.ps1

# Conferma con 's'
# Database pulito e ripopolato in ~2-5 secondi
```

---

## 🚨 Troubleshooting

### Errore: "mongosh: comando non trovato"

**Causa:** MongoDB Shell non installato

**Soluzione:**
```powershell
# Installa con Chocolatey
choco install mongodb-shell

# Oppure scarica da:
# https://www.mongodb.com/try/download/shell
```

### Errore: "Authentication failed"

**Causa:** Credenziali MongoDB errate o certificato non valido

**Soluzione:**
```powershell
# Verifica URI
echo $env:MONGODB_URI

# Verifica certificato esista
Test-Path $env:MONGODB_CERT_PATH

# Testa connessione
mongosh "$env:MONGODB_URI" `
  --tls `
  --tlsCertificateKeyFile "$env:MONGODB_CERT_PATH" `
  --eval "db.version()"
```

### Errore: "Database not found"

**Causa:** Database name errato

**Soluzione:**
```powershell
# Verifica nome database
echo $env:MONGODB_DB_NAME

# Deve corrispondere a quello in MongoDB Atlas
# Default: qr-menu
```

### Script si blocca su "Vuoi configurare ora?"

**Causa:** Richiede input interattivo

**Soluzione:**
```powershell
# Opzione 1: Rispondi 's' o 'n'

# Opzione 2: Pre-configura variabili
$env:MONGODB_URI="..."
$env:MONGODB_CERT_PATH="..."
$env:MONGODB_DB_NAME="qr-menu"
.\scripts\setup_and_seed.ps1
```

### Seed completa ma app non vede dati

**Causa:** App connessa a database diverso

**Soluzione:**
```powershell
# Verifica app usa stesse variabili del seed
echo $env:MONGODB_URI
echo $env:MONGODB_DB_NAME

# Riavvia app
.\qr-menu.exe
```

---

## 📊 Dati Creati dal Seed

### Utente Admin

```javascript
{
  _id: "admin_user_001",
  username: "admin",
  email: "admin@qrmenu.local",
  password_hash: "$2a$10$...", // bcrypt hash di "admin"
  privacy_consent: true,
  marketing_consent: false,
  consent_date: ISODate("2026-03-04T..."),
  created_at: ISODate("2026-03-04T..."),
  last_login: ISODate("2026-03-04T..."),
  is_active: true
}
```

### Ristoranti (4)

```javascript
// rest_001 - Pizzeria Napoletana
{
  _id: "rest_001",
  owner_id: "admin_user_001",
  name: "Pizzeria Napoletana",
  description: "Autentica pizza napoletana con forno a legna...",
  address: "Via Roma 123, Napoli, 80100",
  phone: "+39 081 1234567",
  active_menu_id: "menu_001"
}

// rest_002 - Trattoria Toscana
// rest_003 - Sushi-Ya Tokyo
// rest_004 - Burger House Americana
```

### Menu (4 totali, 37 piatti)

**Menu 1 - Pizzeria (10 piatti):**
- Pizze Classiche: Margherita (€8), Marinara (€6.50), Diavola (€9.50), 4 Stagioni (€11)
- Pizze Speciali: Bufala e Pomodorini (€13), Tartufo Nero (€16)
- Bevande: Birra IPA (€5.50), Acqua (€2)

**Menu 2 - Trattoria (7 piatti):**
- Antipasti: Crostini (€7), Tagliere (€15)
- Primi: Pici Cacio e Pepe (€12), Pappardelle Cinghiale (€14), Ribollita (€10)
- Secondi: Bistecca Fiorentina (€45), Tagliata (€18)

**Menu 3 - Sushi (10 piatti):**
- Nigiri (2pz): Sake (€4.50), Maguro (€6), Ebi (€5), Unagi (€5.50)
- Maki (8pz): California Roll (€7), Spicy Tuna (€8.50), Dragon Roll (€12)
- Sashimi (5pz): Sake (€10), Maguro (€14), Mix (€25)

**Menu 4 - Burger (10 piatti):**
- Classic: American (€11), Cheeseburger (€13.50), Bacon (€12.50)
- Gourmet: Truffle (€16), Mexican (€14), Veggie (€12)
- Sides: Patatine (€4.50), Onion Rings (€5), Bibite (€3), Birra (€4.50)

---

## 🎯 Best Practices

### Development

1. **Seed frequentemente:** Ogni volta che cambi il data model
2. **Database separati:** Dev/Test/Prod con seed solo su Dev/Test
3. **Backup locale:** Prima di seed locale, backup se hai dati importanti
4. **Verifica indici:** Dopo seed, verifica con `db.collection.getIndexes()`

### Production

1. **MAI seed in produzione:** Usa solo `migrate_user_restaurant.js`
2. **Backup obbligatorio:** Prima di qualsiasi operazione
3. **Test su staging:** Seed su ambiente di staging identico a prod
4. **Monitoring post-seed:** Verifica performance e count documenti

### Testing

1. **Re-seed tra test:** Per test isolati e ripetibili
2. **Dati realistici:** I menu di test rappresentano casi d'uso reali
3. **Test multi-restaurant:** Verifica isolamento dati tra ristoranti
4. **Test performance:** Con 4 ristoranti e 37 piatti, testa query speed

---

## 📚 Riferimenti

- [QUICK_START.md](../QUICK_START.md) - Setup applicazione completo
- [MIGRATION_GUIDE.md](../MIGRATION_GUIDE.md) - Guida migrazione produzione
- [BEST_PRACTICES.md](../BEST_PRACTICES.md) - Best practices database
- [IMPLEMENTATION_SUMMARY.md](../IMPLEMENTATION_SUMMARY.md) - Dettagli tecnici

---

## ✅ Checklist Pre-Seed

Prima di eseguire lo script, verifica:

- [ ] MongoDB Atlas configurato
- [ ] Certificato X509 scaricato e salvato
- [ ] Variabili d'ambiente impostate (URI, CERT_PATH, DB_NAME)
- [ ] mongosh installato
- [ ] Database di TEST/DEV (non produzione!)
- [ ] Backup fatto (se hai dati da preservare)
- [ ] Applicazione NON in esecuzione (per evitare conflitti)

**Pronto per il seed? Esegui:**

```powershell
.\scripts\setup_and_seed.ps1
```

🎉 **Buon testing!**
