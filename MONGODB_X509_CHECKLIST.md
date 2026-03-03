# MongoDB X.509 Setup Checklist

## 1. Railway Environment Variables

Verifica in Railway Dashboard → Variables:

### MONGODB_URI (CRITICO)
```
mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authMechanism=MONGODB-X509&authSource=$external&retryWrites=true&w=majority
```

**Elementi obbligatori:**
- ❌ NO username:password nella URI
- ✅ `authMechanism=MONGODB-X509` (case-sensitive!)
- ✅ `authSource=$external`

### MONGODB_CERT_CONTENT
Copia ESATTAMENTE il contenuto del file .pem (including newlines):
```
-----BEGIN CERTIFICATE-----
MIIFCzCCAvOgAwIBAgIIOK+s9wbD3aEwDQYJKoZIhvcNAQELBQAwSTEhMB8GA1UE
...tutto il certificato...
-----END CERTIFICATE-----
-----BEGIN PRIVATE KEY-----
MIIJQQIBADANBgkqhkiG9w0BAQEFAASCCSswggknAgEAAoICAQCxpxNuMbA9D/nc
...tutta la chiave...
-----END PRIVATE KEY-----
```

**Attenzione:** Railway NON richiede escape di newlines (copia diretto)

### MONGODB_DB_NAME (opzionale)
```
qr-menu
```

---

## 2. MongoDB Atlas - Database Access

1. Vai a: **Security → Database Access**
2. Clicca: **Certificate** tab
3. Verifica che il certificato con CN=`qr-menu-dev` sia presente
4. Se non presente:
   - Clicca "Add New Certificate"
   - Incolla il contenuto del certificato (solo la sezione CERTIFICATE, non la private key)
   - MongoDB mostrerà il CN estratto: **CN=qr-menu-dev**
   - Salva

**Username per X.509:** `CN=qr-menu-dev`

---

## 3. MongoDB Atlas - Network Access

1. Vai a: **Security → Network Access**
2. Verifica che sia presente: **0.0.0.0/0** (Allow access from anywhere)
3. Railway usa IP dinamici → serve wildcard

Se non presente:
- Clicca "Add IP Address"
- Seleziona "Allow access from anywhere"
- Conferma `0.0.0.0/0`

---

## 4. Test Locale (Opzionale)

Se vuoi testare la connessione dal tuo PC:

```powershell
# Salva certificato in file
$cert = @"
-----BEGIN CERTIFICATE-----
...
-----END PRIVATE KEY-----
"@
$cert | Out-File -Encoding ASCII cert.pem

# Test connessione
$env:MONGODB_URI = "mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authMechanism=MONGODB-X509&authSource=`$external"
$env:MONGODB_CERT_PATH = "cert.pem"
$env:MONGODB_DB_NAME = "qr-menu"

# Avvia app
go run main.go
```

---

## Alternativa: Username/Password (RAPIDO)

Se X.509 continua a dare problemi, passa a autenticazione standard:

### MongoDB Atlas
1. Security → Database Access → Add New Database User
2. Scegli: **Password Authentication**
3. Username: `qr-menu-user`
4. Password: (genera una sicura)
5. Database User Privileges: **Read and write to any database**

### Railway Variables
```bash
MONGODB_URI=mongodb+srv://qr-menu-user:PASSWORD@ac-d8zdak4.b9jfwmr.mongodb.net/qr-menu?retryWrites=true&w=majority
MONGODB_DB_NAME=qr-menu
```

**RIMUOVI:** `MONGODB_CERT_CONTENT` (il codice userà password se non trova cert)

---

## Debug: Estrai CN dal Certificato

```bash
# Linux/Mac/WSL
openssl x509 -in cert.pem -noout -subject

# Output atteso:
# subject=CN = qr-menu-dev
```

Il CN **DEVE** corrispondere esattamente a quello registrato in Atlas.

---

## Prossimo Step

Quale metodo preferisci?

**OPZIONE A - Fix X.509:**
1. Verifica che MONGODB_URI includa `authMechanism=MONGODB-X509`
2. Verifica che il certificato sia registrato in Atlas con CN=qr-menu-dev
3. Rideploy Railway

**OPZIONE B - Switch a Password (5 minuti):**
1. Crea utente password in Atlas
2. Aggiorna MONGODB_URI in Railway
3. Rimuovi MONGODB_CERT_CONTENT
4. Rideploy Railway

Fammi sapere quale preferisci!
