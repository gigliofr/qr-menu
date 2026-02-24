# Piano di Test Utente - QR Menu System
*Versione 2.0.0 | Test Completo End-to-End*

---

## 🎯 Obiettivo

Validare tutte le funzionalità del sistema dal punto di vista dell'utente finale, coprendo tutti i casi d'uso principali e verificando l'integrazione tra i componenti.

---

## 📋 Prerequisiti Test

### Setup Ambiente di Test

```bash
# 1. Avvia server
cd C:\Users\gigli\GoWs\qr-menu
.\qr-menu.exe

# Server avviato su http://localhost:8080
```

### Utenti di Test

| Username | Password | Ruolo | Scopo |
|----------|----------|-------|-------|
| `admin` | `admin123` | Super Admin | Test funzioni amministrative |
| `owner1` | `pass123` | Restaurant Owner | Test creazione menu |
| `staff1` | `pass123` | Staff | Test visualizzazione limitata |

---

## 🧪 Test Suite

### FASE 1: Autenticazione & Accesso

#### Test 1.1: Registrazione Nuovo Utente
**Obiettivo**: Verificare registrazione nuovo ristoratore

**Steps:**
1. Apri browser: `http://localhost:8080/register`
2. Compila form:
   - Username: `testuser1`
   - Password: `Test123!`
   - Conferma password: `Test123!`
   - Nome ristorante: `Pizzeria Test`
3. Clicca "Registra"

**Risultato Atteso**: ✅
- Reindirizzamento a `/login`
- Messaggio "Registrazione completata"
- Utente creato nel sistema

**Risultato Effettivo**: _____________

---

#### Test 1.2: Login
**Obiettivo**: Accesso con credenziali valide

**Steps:**
1. Apri `http://localhost:8080/login`
2. Username: `admin`
3. Password: `admin123`
4. Clicca "Login"

**Risultato Atteso**: ✅
- Reindirizzamento a `/admin`
- Cookie JWT impostato
- Dashboard visibile

**Risultato Effettivo**: _____________

---

#### Test 1.3: Login Fallito
**Obiettivo**: Gestione credenziali errate

**Steps:**
1. Apri `/login`
2. Username: `admin`
3. Password: `wrong_password`
4. Clicca "Login"

**Risultato Atteso**: ✅
- Messaggio errore "Credenziali non valide"
- Rimane su pagina login
- Nessun cookie impostato

**Risultato Effettivo**: _____________

---

### FASE 2: Gestione Menu

#### Test 2.1: Creazione Menu Base
**Obiettivo**: Creare menu semplice

**Steps:**
1. Login come `owner1`
2. Vai a `/admin/menu/create`
3. Compila:
   - Nome menu: `Menu Primavera 2026`
   - Descrizione: `Menu stagionale primaverile`
4. Clicca "Crea Menu"

**Risultato Atteso**: ✅
- Menu creato con ID univoco
- Redirect a pagina modifica menu
- Menu visibile in lista

**Risultato Effettivo**: _____________

---

#### Test 2.2: Aggiunta Categoria
**Obiettivo**: Organizzare menu in categorie

**Steps:**
1. Nella pagina modifica menu
2. Sezione "Aggiungi Categoria"
3. Nome: `Antipasti`
4. Descrizione: `Antipasti freschi di stagione`
5. Clicca "Aggiungi"

**Risultato Atteso**: ✅
- Categoria aggiunta al menu
- Visibile nella lista categorie
- Possibilità di aggiungere item

**Risultato Effettivo**: _____________

---

#### Test 2.3: Aggiunta Item con Immagine
**Obiettivo**: Creare piatto completo

**Steps:**
1. Nella categoria "Antipasti"
2. Clicca "Aggiungi Piatto"
3. Compila:
   - Nome: `Bruschetta al Pomodoro`
   - Descrizione: `Pane tostato con pomodori freschi e basilico`
   - Prezzo: `6.50`
   - Allergeni: `Glutine`
   - Disponibile: ☑️
4. Upload immagine (png/jpg)
5. Salva

**Risultato Atteso**: ✅
- Item creato con tutti i campi
- Immagine caricata e visualizzata
- Prezzo formattato correttamente (€ 6,50)

**Risultato Effettivo**: _____________

---

#### Test 2.4: Duplicazione Item
**Obiettivo**: Velocizzare creazione piatti simili

**Steps:**
1. Trova item "Bruschetta al Pomodoro"
2. Clicca "Duplica"
3. Modifica nome in "Bruschetta ai Funghi"
4. Modifica descrizione e prezzo
5. Salva

**Risultato Atteso**: ✅
- Nuovo item creato con stessi campi base
- ID diverso
- Modifiche applicate

**Risultato Effettivo**: _____________

---

#### Test 2.5: Generazione QR Code
**Obiettivo**: Creare QR code per menu

**Steps:**
1. Completa menu con almeno 3 categorie e 10 item
2. Clicca "Completa Menu"
3. Clicca "Genera QR Code"

**Risultato Atteso**: ✅
- QR code generato
- Immagine PNG salvata in `/static/qrcodes/`
- Link pubblico al menu: `/menu/{id}`

**Risultato Effettivo**: _____________

---

#### Test 2.6: Visualizzazione Menu Pubblico
**Obiettivo**: Verificare esperienza cliente

**Steps:**
1. Apri browser in modalità incognito
2. Vai a `http://localhost:8080/menu/{menu_id}`
   (sostituisci {menu_id} con ID reale)
3. Naviga tra categorie
4. Visualizza dettagli piatti

**Risultato Atteso**: ✅
- Menu visibile senza login
- Tutte le categorie mostrate
- Immagini caricate
- Prezzi formattati
- Layout responsive (test mobile)

**Risultato Effettivo**: _____________

---

### FASE 3: Analytics & Tracking

#### Test 3.1: Tracciamento Visualizzazioni
**Obiettivo**: Verificare conteggio visite

**Steps:**
1. Apri menu pubblico in 3 browser differenti
2. Vai a `/admin/analytics`
3. Seleziona ristorante
4. Verifica metrics

**Risultato Atteso**: ✅
- Total Views ≥ 3
- Grafico visualizzazioni aggiornato
- Device types mostrati

**Risultato Effettivo**: _____________

---

#### Test 3.2: Item Popolari
**Obiettivo**: Tracking item più visti

**Steps:**
1. Nel menu pubblico, clicca su 5 item diversi
2. Clicca 3 volte su "Bruschetta al Pomodoro"
3. Vai a `/admin/analytics`
4. Sezione "Item Popolari"

**Risultato Atteso**: ✅
- "Bruschetta al Pomodoro" in top 3
- Numero views corretto
- Ordinamento per popolarità

**Risultato Effettivo**: _____________

---

#### Test 3.3: Share Tracking
**Obiettivo**: Tracciare condivisioni social

**Steps:**
1. Menu pubblico
2. Clicca bottone "Condividi su WhatsApp"
3. Torna a `/admin/analytics`
4. Verifica "Share Stats"

**Risultato Atteso**: ✅
- WhatsApp count incrementato
- Timestamp registrato
- Grafico aggiornato

**Risultato Effettivo**: _____________

---

#### Test 3.4: API Analytics
**Obiettivo**: Verificare endpoint JSON

**Steps:**
```bash
# PowerShell
$token = "your_jwt_token"
Invoke-RestMethod -Uri "http://localhost:8080/api/analytics?restaurant_id=test" `
  -Headers @{"Authorization"="Bearer $token"}
```

**Risultato Atteso**: ✅
- JSON response con tutte le metriche
- Status 200
- Dati consistenti con dashboard

**Risultato Effettivo**: _____________

---

### FASE 4: Notifiche

#### Test 4.1: Invio Notifica Email
**Obiettivo**: Sistema notifiche funzionante

**Steps:**
```bash
POST /api/notifications/send
{
  "restaurant_id": "test",
  "type": "email",
  "title": "Nuovo Menu Disponibile",
  "body": "Scopri il nostro menu primaverile!",
  "data": {
    "menu_id": "menu123"
  }
}
```

**Risultato Atteso**: ✅
- Status 200
- Notifica in coda
- Log registrato

**Risultato Effettivo**: _____________

---

#### Test 4.2: Preferenze Notifiche
**Obiettivo**: Gestione preferenze utente

**Steps:**
1. `GET /api/notifications/preferences` (logged in)
2. `PUT /api/notifications/preferences`:
```json
{
  "email_enabled": true,
  "push_enabled": false,
  "quiet_hours": {
    "enabled": true,
    "start": "22:00",
    "end": "08:00"
  }
}
```
3. Verifica con GET

**Risultato Atteso**: ✅
- Preferenze salvate
- Quiet hours rispettate
- GET ritorna preferenze aggiornate

**Risultato Effettivo**: _____________

---

### FASE 5: Backup & Restore

#### Test 5.1: Creazione Backup Manuale
**Obiettivo**: Backup on-demand

**Steps:**
```bash
POST /api/backup/create
```

**Risultato Atteso**: ✅
- Backup ZIP creato in `/backups/`
- Metadata salvato
- Dimensione file > 0

**Risultato Effettivo**: _____________

---

#### Test 5.2: Lista Backup
**Obiettivo**: Visualizzare backup disponibili

**Steps:**
```bash
GET /api/backup/list
```

**Risultato Atteso**: ✅
- Array di backup
- Ogni backup ha: id, timestamp, size, type
- Ordinati per data (più recente primo)

**Risultato Effettivo**: _____________

---

#### Test 5.3: Restore da Backup
**Obiettivo**: Ripristino dati

**Steps:**
1. Crea backup
2. Modifica alcuni menu
3. `POST /api/backup/restore` con backup_id
4. Verifica menu ripristinati

**Risultato Atteso**: ✅
- Restore completato
- Dati tornati allo stato del backup
- Log restore registrato

**Risultato Effettivo**: _____________

---

### FASE 6: Localizzazione (i18n)

#### Test 6.1: Cambio Lingua UI
**Obiettivo**: Interfaccia multilingua

**Steps:**
1. Menu pubblico in italiano (default)
2. Clicca selector lingua
3. Seleziona "English"
4. Verifica traduzioni

**Risultato Atteso**: ✅
- UI tradotta in inglese
- Pulsanti, label, messaggi in EN
- Preferenza salvata in cookie/session

**Risultato Effettivo**: _____________

---

#### Test 6.2: Formattazione Valuta
**Obiettivo**: Prezzi localizzati

**Steps:**
1. Locale IT: 6.50 → "€ 6,50"
2. Locale EN: 6.50 → "$6.50"
3. Locale DE: 6.50 → "6,50 €"

**API:**
```bash
GET /api/localization/format-currency?locale=it&amount=6.50
```

**Risultato Atteso**: ✅
- Formati corretti per ogni locale
- Simbolo valuta posizione corretta

**Risultato Effettivo**: _____________

---

### FASE 7: PWA (Progressive Web App)

#### Test 7.1: Installazione PWA
**Obiettivo**: App installabile

**Steps:**
1. Chrome desktop: vai a home
2. Barra indirizzo: icona "Installa"
3. Clicca "Installa QR Menu"

**Risultato Atteso**: ✅
- App installata
- Icona desktop/menu Start
- Apre in finestra standalone

**Risultato Effettivo**: _____________

---

#### Test 7.2: Funzionalità Offline
**Obiettivo**: Accesso senza connessione

**Steps:**
1. Visita home e menu
2. Chrome DevTools → Network → Offline
3. Ricarica pagina
4. Naviga nel menu

**Risultato Atteso**: ✅
- Pagine cache visualizzate
- Service Worker attivo
- Messaggi offline chiari

**Risultato Effettivo**: _____________

---

#### Test 7.3: Manifest.json
**Obiettivo**: Configurazione PWA

**Steps:**
1. Apri `http://localhost:8080/manifest.json`
2. Verifica campi

**Risultato Atteso**: ✅
```json
{
  "name": "QR Menu System",
  "short_name": "QR Menu",
  "start_url": "/",
  "theme_color": "#2E7D32",
  "icons": [...]
}
```

**Risultato Effettivo**: _____________

---

### FASE 8: Machine Learning

#### Test 8.1: Recommendations
**Obiettivo**: Suggerimenti personalizzati

**Steps:**
1. Come utente, visualizza 5 item diversi
2. `GET /api/v1/ml/recommendations?limit=5`

**Risultato Atteso**: ✅
- 5 item raccomandati
- Diversi da quelli già visti
- Score di similarity

**Risultato Effettivo**: _____________

---

#### Test 8.2: Forecasting
**Obiettivo**: Previsione domanda

**Steps:**
```bash
# Aggiungi dati storici
POST /api/v1/ml/data-points (x30 giorni)

# Forecast
GET /api/v1/ml/forecast?metric=orders&periods=7
```

**Risultato Atteso**: ✅
- 7 giorni di previsioni
- Confidence intervals
- Trend visibile

**Risultato Effettivo**: _____________

---

#### Test 8.3: A/B Testing
**Obiettivo**: Esperimenti variant testing

**Steps:**
1. Crea esperimento:
```json
POST /api/v1/ml/experiments
{
  "name": "Menu Layout Test",
  "variants": [
    {"id": "control", "name": "Grid", "traffic": 0.5},
    {"id": "variant_a", "name": "List", "traffic": 0.5}
  ]
}
```
2. Start esperimento
3. Assegna 100 utenti random
4. Track 50 conversioni
5. Get results

**Risultato Atteso**: ✅
- Esperimento creato
- Utenti distribuiti 50/50
- Risultati con p-value
- Winner determinato (se significativo)

**Risultato Effettivo**: _____________

---

### FASE 9: Security & GDPR

#### Test 9.1: Rate Limiting
**Obiettivo**: Protezione DDoS

**Steps:**
```bash
# PowerShell - 150 richieste in 1 minuto
1..150 | ForEach-Object {
  Invoke-WebRequest http://localhost:8080/api/menus
}
```

**Risultato Atteso**: ✅
- Prime 100 req: 200 OK
- Successive: 429 Too Many Requests
- Header `X-RateLimit-Remaining` decrescente

**Risultato Effettivo**: _____________

---

#### Test 9.2: Audit Log
**Obiettivo**: Tracciamento eventi di sicurezza

**Steps:**
1. Login (successo)
2. Login (fallito)
3. Create menu
4. Delete item
5. `GET /api/v1/security/audit-logs`

**Risultato Atteso**: ✅
- Tutti gli eventi registrati
- Timestamp, user_id, IP, action
- Outcome (success/failure)

**Risultato Effettivo**: _____________

---

#### Test 9.3: GDPR Export
**Obiettivo**: Diritto di accesso dati

**Steps:**
```bash
GET /api/v1/gdpr/export
```

**Risultato Atteso**: ✅
- JSON con tutti i dati utente
- Menus, analytics, preferences, history
- Formato machine-readable

**Risultato Effettivo**: _____________

---

#### Test 9.4: GDPR Delete
**Obiettivo**: Diritto all'oblio

**Steps:**
1. `DELETE /api/v1/gdpr/delete`
2. Conferma
3. Logout
4. Tenta login

**Risultato Atteso**: ✅
- Tutti i dati utente cancellati
- Login fallisce
- Account non esiste

**Risultato Effettivo**: _____________

---

### FASE 10: Mobile App (Flutter)

#### Test 10.1: QR Scanner
**Obiettivo**: Scansione QR code

**Steps:**
1. Apri app mobile
2. Tap "Scansiona QR"
3. Inquadra QR code generato

**Risultato Atteso**: ✅
- Camera aperta
- QR code rilevato
- Redirect a menu

**Risultato Effettivo**: _____________

---

#### Test 10.2: Menu Navigation
**Obiettivo**: Navigazione mobile

**Steps:**
1. Visita menu da app
2. Tap categorie
3. Swipe tra item
4. Zoom immagini

**Risultato Atteso**: ✅
- Smooth transitions
- Immagini caricate
- Tap responsive
- Layout ottimizzato mobile

**Risultato Effettivo**: _____________

---

## 📊 Riepilogo Test

### Checklist Finale

| Area | Test Passed | Test Failed | Coverage |
|------|-------------|-------------|----------|
| Auth & Access | __ / 3 | __ | __% |
| Menu Management | __ / 6 | __ | __% |
| Analytics | __ / 4 | __ | __% |
| Notifications | __ / 2 | __ | __% |
| Backup | __ / 3 | __ | __% |
| i18n | __ / 2 | __ | __% |
| PWA | __ / 3 | __ | __% |
| ML/AI | __ / 3 | __ | __% |
| Security | __ / 4 | __ | __% |
| Mobile | __ / 2 | __ | __% |
| **TOTALE** | **__ / 32** | **__** | **__%** |

### Rating Qualità

- ✅ **Eccellente**: 32/32 test passed (100%)
- 🟢 **Buono**: 28-31 test passed (87-96%)
- 🟡 **Accettabile**: 24-27 test passed (75-84%)
- 🔴 **Insufficiente**: < 24 test passed (< 75%)

---

## 🐛 Bug Tracking

| ID | Descrizione | Severità | Status | Assegnato | Data |
|----|-------------|----------|--------|-----------|------|
| 001 | ... | High | Open | ... | ... |
| 002 | ... | Medium | Fixed | ... | ... |

---

## 📝 Note Tester

**Data Test**: _______________
**Tester**: _______________
**Ambiente**: Production / Staging / Local
**Browser**: Chrome / Firefox / Safari / Edge
**OS**: Windows / macOS / Linux

**Commenti Aggiuntivi**:
_______________________________________________
_______________________________________________
_______________________________________________

---

## ✅ Sign-off

**Test Completati da**: _______________
**Data**: _______________
**Firma**: _______________

**Approvazione QA**: _______________
**Data**: _______________

---

*Fine Piano di Test*
