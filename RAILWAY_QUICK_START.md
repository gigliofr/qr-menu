# 🚀 Deploy su Railway - ISTRUZIONI RAPIDE

## Veloce: 10 minuti totali

### Passo 1: Accedi a Railway
1. Vai su **https://railway.app**
2. Clicca **"Login"** (usa GitHub account)

### Passo 2: Crea Nuovo Progetto
1. Clicca **"New Project"**
2. Seleziona **"Deploy from GitHub repo"**
3. Autorizza Railway ad accedere ai tuoi repo GitHub

### Passo 3: Seleziona il Repo
1. Cerca il repo **`qr-menu`**
2. Clicca per selezionarlo

### Passo 4: Configura il Servizio
Railway detecta automaticamente il Dockerfile.

Lo vedi nella dashboard:
- **Service**: qr-menu
- **Status**: Building... (5-10 minuti)

### Passo 5: Aggiungi Variabili d'Ambiente (IMPORTANTE!)
Una volta nel progetto Railway:
1. Clicca sul servizio **qr-menu**
2. Vai a **"Variables"**
3. Aggiungi una nuova variabile:
   - **Key**: `MONGODB_URI`
   - **Value**: `mongodb+srv://qr-menu-dev@cluster0.b9jfwmr.mongodb.net/qr-menu?authSource=$external&authMechanism=MONGODB-X509`

### Passo 6: Carica il Certificato
**Opzione A** (Consigliato): Via Secret
1. In Railway → **"Variables"** → **"Add Secret"**
2. **Key**: `MONGODB_CERT_PATH`
3. **Value**: Incolla il contenuto del file `X509-cert-4084673564018728353.pem`

**Opzione B**: Carica il file
1. In Railway → **"Deployments"**
2. Uploda il certificato nella root della build

### Passo 7: Attendere Deploy
- Train di build completato (5-10 minuti)
- Railway assegna URL automatico:
  ```
  https://qr-menu-xxx.railway.app
  ```

---

## ✅ Dopo il Deploy

Test immediato:
```bash
curl https://qr-menu-xxx.railway.app/health
```

Dovresti ricevere:
```json
{"status": "ok"}
```

---

## 🎯 Prossimo Passo

1. Vai su **https://railway.app**
2. Registrati/Login con GitHub
3. Crea nuovo progetto da `qr-menu` repo
4. Fammelo sapere quando vedi il build in corso! ✅

---

## Se qualcosa non funziona

Posso anche fare il deploy via file `.railway.json` che autopropagates le variabili.

Fammelo sapere quando sei pronto!
