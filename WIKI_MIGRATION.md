# 📚 Migrazione Documentazione su GitHub Wiki

## 🎯 Obiettivo

Spostare la documentazione estesa dalla root del progetto alla **GitHub Wiki** per mantenere il repository pulito e focalizzato sul codice.

---

## 📋 Struttura Wiki Proposta

### **Home** (Pagina principale)
```markdown
# QR Menu - Wiki

Benvenuto nella documentazione completa di QR Menu!

## 📖 Indice

### Setup e Configurazione
- [Configurazione Progetto Non Commerciale](Configurazione-Non-Commerciale)
- [Guida Compliance Legale Italiana](Guida-Compliance-Italiana)

### Architettura e Sviluppo
- [Architettura del Sistema](Architettura-Sistema)
- [Modulo Security](Security-Module)
- [Testing Guide](Testing-Guide)

### Troubleshooting
- [Bug Menu Non Visibili - RISOLTO](Troubleshooting-Menu-Visibili)
- [Risoluzione Alert Sicurezza GitHub](Security-Alert-Resolution)

### Deployment
- [Deploy su Railway](Deploy-Railway)
- [Deploy su Google Cloud Run](Deploy-Cloud-Run)

---

Torna al [Repository principale →](https://github.com/gigliofr/qr-menu)
```

---

## 📄 Contenuto Pagine Wiki

### 1. **Configurazione-Non-Commerciale**
**File sorgente:** `SETUP_NON_COMMERCIAL.md`

Copiare il contenuto completo di `SETUP_NON_COMMERCIAL.md` in questa pagina wiki.

---

### 2. **Guida-Compliance-Italiana**
**File sorgente:** `LEGAL_COMPLIANCE_IT.md`

Copiare il contenuto completo di `LEGAL_COMPLIANCE_IT.md` in questa pagina wiki.

---

### 3. **Architettura-Sistema**
**File sorgente:** `ARCHITECTURE.md`

Copiare il contenuto completo di `ARCHITECTURE.md` in questa pagina wiki.

---

### 4. **Security-Module**
**File sorgente:** `security/README.md`

Copiare il contenuto completo di `security/README.md` in questa pagina wiki.
**Nota:** Mantenere anche il file originale nella cartella `security/`.

---

### 5. **Testing-Guide**
**File sorgente:** `tests/README.md`

Copiare il contenuto completo di `tests/README.md` in questa pagina wiki.
**Nota:** Mantenere anche il file originale nella cartella `tests/`.

---

### 6. **Troubleshooting-Menu-Visibili**
**File sorgente:** `RISOLUZIONE_MENU_VISIBILI.md`

Copiare il contenuto completo di `RISOLUZIONE_MENU_VISIBILI.md` in questa pagina wiki.

---

### 7. **Security-Alert-Resolution**
**File sorgente:** `SECURITY_ALERT_RESOLUTION.md` + `DISMISS_ALERT.md`

**Unire i due file** in un'unica pagina:

```markdown
# Risoluzione Alert Sicurezza GitHub

[Contenuto di SECURITY_ALERT_RESOLUTION.md]

---

## Guida Rapida: Dismissal Alert

[Contenuto di DISMISS_ALERT.md]
```

---

### 8. **Deploy-Railway** (Nuovo)
Creare una nuova pagina con le istruzioni deployment Railway:

```markdown
# Deploy su Railway

## Setup

1. Crea account su [Railway.app](https://railway.app)
2. Connetti il repository GitHub
3. Configura variabili d'ambiente:

```bash
MONGODB_URI=your_connection_string
MONGODB_CERT_CONTENT=your_certificate_pem
MONGODB_DB_NAME=qr-menu
PORT=8080
```

4. Deploy automatico ad ogni push su `main`

## URL

Production: https://qr-menu-staging.up.railway.app
```

---

### 9. **Deploy-Cloud-Run** (Nuovo)
Creare una nuova pagina con le istruzioni deployment Google Cloud Run:

```markdown
# Deploy su Google Cloud Run

## Prerequisiti

```bash
# Installa gcloud CLI
winget install Google.CloudSDK

# Login
gcloud auth login

# Set project
gcloud config set project qr-menu-20241217
```

## Deploy

```bash
gcloud run deploy qr-menu \
  --source . \
  --platform managed \
  --region us-central1 \
  --port 8080 \
  --allow-unauthenticated
```

## Variabili d'Ambiente

Configura su Cloud Console o via CLI:
- MONGODB_URI
- MONGODB_CERT_CONTENT
- MONGODB_DB_NAME
```

---

## 🚀 Step per Step: Creazione Wiki

### Passo 1: Abilita Wiki su GitHub

1. Vai su: https://github.com/gigliofr/qr-menu/settings
2. Scroll fino a "Features"
3. Spunta **"Wikis"** ✅
4. Salva

### Passo 2: Crea Pagina Home

1. Vai su: https://github.com/gigliofr/qr-menu/wiki
2. Click su **"Create the first page"**
3. Titolo: `Home`
4. Copia il contenuto della sezione **Home** sopra
5. Salva con: **"Create Page"**

### Passo 3: Crea Pagine Documentazione

Per ogni pagina elencata sopra:

1. Click su **"New Page"**
2. Inserisci il titolo (es: `Configurazione-Non-Commerciale`)
3. Copia il contenuto dal file sorgente
4. Click **"Save Page"**

Ripeti per tutte le 9 pagine.

### Passo 4: Verifica Link

- Clicca su ogni link dalla Home
- Verifica che le pagine si aprano correttamente
- Correggi eventuali link rotti

---

## 🗑️ Pulizia Repository

**Dopo aver creato la wiki**, esegui:

```powershell
# Rimuovi file spostati in wiki
git rm ARCHITECTURE.md
git rm DISMISS_ALERT.md
git rm LEGAL_COMPLIANCE_IT.md
git rm RISOLUZIONE_MENU_VISIBILI.md
git rm SECURITY_ALERT_RESOLUTION.md
git rm SETUP_NON_COMMERCIAL.md
git rm WIKI_MIGRATION.md

# Commit
git commit -m "docs: Move extended documentation to GitHub Wiki"

# Push
git push origin main
```

---

## ✅ Checklist Migrazione

- [ ] Wiki abilitata su GitHub
- [ ] Pagina **Home** creata
- [ ] Pagina **Configurazione-Non-Commerciale** creata
- [ ] Pagina **Guida-Compliance-Italiana** creata
- [ ] Pagina **Architettura-Sistema** creata
- [ ] Pagina **Security-Module** creata
- [ ] Pagina **Testing-Guide** creata
- [ ] Pagina **Troubleshooting-Menu-Visibili** creata
- [ ] Pagina **Security-Alert-Resolution** creata (unione 2 file)
- [ ] Pagina **Deploy-Railway** creata
- [ ] Pagina **Deploy-Cloud-Run** creata
- [ ] Link verificati dalla Home
- [ ] README.md aggiornato con link wiki
- [ ] File .md rimossi dal repository
- [ ] Commit e push finale

---

## 📍 Link Utili

- **Repository:** https://github.com/gigliofr/qr-menu
- **Wiki:** https://github.com/gigliofr/qr-menu/wiki
- **Railway:** https://qr-menu-staging.up.railway.app

---

**Tempo stimato:** 15-20 minuti
**Difficoltà:** ⭐ Facile
