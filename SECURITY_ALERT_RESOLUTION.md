# 🚨 SECURITY ALERT - Piano di Risoluzione

**Alert GitHub:** Secrets detected in gigliofr/qr-menu  
**File:** MONGODB_X509_CHECKLIST.md (riga 104)  
**Commit:** 1afcc465  
**Tipo:** MongoDB Atlas Database URI with credentials  
**Data:** 4 Marzo 2026

---

## ⚠️ Situazione Attuale

### Segreto Esposto
```
MONGODB_URI=mongodb+srv://qr-menu-user:PASSWORD@ac-d8zdak4.b9jfwmr.mongodb.net/...
```

**Cluster MongoDB:** `ac-d8zdak4.b9jfwmr.mongodb.net`

### Status
- ✅ File `MONGODB_X509_CHECKLIST.md` eliminato (commit 3e04454)
- ❌ Segreto ancora presente nella **cronologia Git**
- ⚠️ Chiunque con accesso read al repository può vedere il commit storico

---

## 🔒 Azioni Immediate Richieste

### 1. Verifica Autenticazione Attuale

Il progetto usa **X.509 Certificate Authentication**, NON username/password. Verifica su Railway:

```powershell
# Controlla quale metodo è configurato
echo $env:MONGODB_URI
```

**Dovrebbe contenere:**
- ✅ `authMechanism=MONGODB-X509` (sicuro)
- ❌ ~~`qr-menu-user:PASSWORD`~~ (esposto, da eliminare)

### 2. Rotazione Credenziali MongoDB Atlas

#### Opzione A: Se usi X.509 (RACCOMANDATO)
**Non c'è rischio!** L'URI esposto con password non è più usato.

✅ **Azione:** Dismissi l'alert su GitHub (vedi sotto)

#### Opzione B: Se usi ancora username/password
**AZIONE URGENTE RICHIESTA:**

1. **Vai su MongoDB Atlas:**
   - https://cloud.mongodb.com
   - Cluster: `ac-d8zdak4.b9jfwmr.mongodb.net`
   
2. **Elimina l'utente esposto:**
   - Security → Database Access
   - Trova utente `qr-menu-user`
   - Click "DELETE"

3. **Crea nuovo utente (o passa a X.509):**
   ```
   Username: qr-menu-prod
   Password: [genera password forte 32 caratteri]
   Privileges: readWrite su qr-menu database
   ```

4. **Aggiorna Railway:**
   - https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab
   - Variables → `MONGODB_URI`
   - Sostituisci con nuovo URI

---

## 📋 Dismissione Alert GitHub

### Dopo aver verificato la sicurezza:

1. **Vai su GitHub:**
   - https://github.com/gigliofr/qr-menu/security/secret-scanning

2. **Trova l'alert:** "MongoDB Atlas Database URI"

3. **Click "Dismiss alert" → Scegli motivo:**
   - ✅ **"Used in tests"** - Se era un esempio/placeholder
   - ✅ **"Revoked"** - Se hai eliminato l'utente MongoDB
   - ✅ **"False positive"** - Se l'URI non conteneva password reale

4. **Aggiungi commento:**
   ```
   URI con username/password non più utilizzato.
   Sistema migrato a X.509 certificate authentication.
   File rimosso nel commit 3e04454.
   ```

5. **Click "Dismiss alert"**

---

## 🧹 Pulizia Cronologia Git (OPZIONALE - DRASTICO)

### ⚠️ ATTENZIONE
Modificare la cronologia Git richiede **force push** e invalida tutti i clone esistenti.

**NON necessario se:**
- Le credenziali sono già state revocate
- Stai usando X.509 al posto di password

**Necessario SOLO se:**
- Credenziali ancora valide
- Repository pubblico
- Non puoi revocare le credenziali

### Se devi procedere:

```powershell
# BACKUP prima di tutto
git clone https://github.com/gigliofr/qr-menu.git qr-menu-backup

# Usa BFG Repo Cleaner (più sicuro di git filter-branch)
# Download: https://rtyley.github.io/bfg-repo-cleaner/

# 1. Clone bare
git clone --mirror https://github.com/gigliofr/qr-menu.git

# 2. Rimuovi file
cd qr-menu.git
java -jar bfg.jar --delete-files MONGODB_X509_CHECKLIST.md

# 3. Cleanup
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# 4. Force push (⚠️ DISTRUTTIVO)
git push --force

# 5. Tutti i collaboratori devono riclonare
```

**Alternative più sicure:**
- Usa GitHub's "Remove sensitive data" tool
- Contatta GitHub Support per assistenza

---

## ✅ Checklist Risoluzione

### Immediato
- [ ] Verificare quale autenticazione è configurata su Railway
- [ ] Se X.509: Dismissi alert (non c'è rischio)
- [ ] Se password: Elimina utente MongoDB esposto
- [ ] Crea nuovo utente/certificato se necessario
- [ ] Aggiorna `MONGODB_URI` su Railway
- [ ] Testa connessione con nuovo URI

### GitHub
- [ ] Vai su Security → Secret Scanning
- [ ] Dismissi alert con motivazione appropriata
- [ ] Aggiungi commento che spiega la risoluzione
- [ ] Verifica che non ci siano altri alert

### Documentazione
- [ ] Aggiorna README.md con best practices
- [ ] Non committare mai credenziali in chiaro
- [ ] Usa sempre variabili d'ambiente
- [ ] Aggiungi `.env` a `.gitignore`

### Prevenzione Futura
- [ ] Abilita GitHub Secret Scanning (già attivo ✅)
- [ ] Considera GitHub Advanced Security
- [ ] Usa git-secrets o gitleaks localmente
- [ ] Code review obbligatorio per file di configurazione

---

## 🔍 Verifica Sicurezza Attuale

### Test Connessione
```powershell
# Verifica che il sistema usi X.509
cd tests
.\verifica_api_menu.ps1

# Controlla nei log
# Dovrebbe mostrare: "MongoDB connected" senza errori
```

### Controlla Railway Logs
```
https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab
→ Deployments → Latest → View Logs

# Cerca:
✅ "MongoDB connection successful"
❌ "authentication failed"
```

---

## 📞 Contatti Emergenza

### Se sospetti accesso non autorizzato:

1. **MongoDB Atlas:**
   - Vai su Security → Network Access
   - Verifica IP whitelist
   - Controlla Access Logs per connessioni sospette

2. **Railway:**
   - Controlla deployment logs
   - Verifica variabili d'ambiente non modificate
   - Check access logs (se disponibili)

3. **GitHub:**
   - Settings → Security log
   - Verifica accessi recenti al repository

---

## 📚 Best Practices per il Futuro

### ✅ DA FARE
- Usa X.509 certificates per MongoDB
- Variabili d'ambiente per tutti i segreti
- File `.env` in `.gitignore`
- GitHub Secret Scanning attivo
- Rotazione credenziali periodica (90 giorni)
- Audit logs su MongoDB Atlas

### ❌ NON FARE
- ~~Committare credenziali in chiaro~~
- ~~URI completi nei file di documentazione~~
- ~~Password in esempi di codice~~
- ~~File di configurazione con segreti~~
- ~~Condividere credenziali via chat/email~~

---

## 🎯 Recommended Action Plan

**Se usi X.509 (molto probabile):**
```
1. Verifica Railway usa MONGODB-X509
2. Dismissi alert GitHub come "Revoked" o "Used in tests"
3. Fine! ✅
```

**Se usi password:**
```
1. ELIMINA subito utente qr-menu-user su MongoDB Atlas
2. Crea nuovo utente con password forte
3. Aggiorna MONGODB_URI su Railway
4. Redeploy applicazione
5. Testa con verifica_api_menu.ps1
6. Dismissi alert su GitHub
7. Considera migrazione a X.509
```

---

## 📝 Note Tecniche

### Perché l'alert persiste?
- Git memorizza **tutta la cronologia**
- Eliminare un file NON rimuove il contenuto dai commit passati
- GitHub Secret Scanning scansiona tutto il repository inclusa la storia

### Come GitHub trova i segreti?
- Pattern matching su stringhe
- Database di pattern noti (MongoDB URI, AWS keys, etc.)
- Machine learning per identificare credenziali

### L'alert è grave?
- **SÌ se:** Le credenziali sono ancora valide
- **NO se:** Già revocate o mai state valide (placeholder)
- **FORSE se:** Repository privato con accesso limitato

### Next Steps
1. Segui il "Recommended Action Plan" sopra
2. Documenta le azioni prese
3. Dismissi l'alert con motivazione chiara
4. Implementa prevenzione per il futuro

---

**Creato:** 4 Marzo 2026  
**Ultima modifica:** 4 Marzo 2026  
**Status:** In attesa di azione utente
