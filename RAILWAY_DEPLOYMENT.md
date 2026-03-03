# 🚀 Deployment su Railway - Guida Rapida

## Cosa è Railway?
- Platform di deployment moderno per app con Docker
- Support per Go automatico
- Integrazione GitHub (auto-deploy on push)
- FREE $5 di credito al mese

## Passi:

### 1️⃣ Registrati su Railway
Apri: https://railway.app

Puoi usare:
- GitHub account (più veloce)
- Email

### 2️⃣ Collega il GitHub repository

Una volta registrato:
1. Clicca **"New Project"**
2. Seleziona **"Deploy from GitHub repo"**
3. Autorizza Railway ad accedere ai tuoi repo
4. Seleziona il repo **`qr-menu`**

### 3️⃣ Railway detecta il Dockerfile

Railway vede automaticamente il nostro `Dockerfile` e:
- Builderà l'immagine Docker
- Deployerà il servizio
- Assegnerà un URL pubblico

### 4️⃣ Configura le Variabili d'Ambiente (IMPORTANTE!)

Nel progetto Railway, aggiungi una nuova **Environment Variable**:

```
MONGODB_CERT_PATH=/opt/render/project/src/X509-cert-4084673564018728353.pem
```

Oppure carica il certificato di MongoDB:
1. Nella dashboard Railway
2. **Service** → **Variables**
3. Carica il file `X509-cert-4084673564018728353.pem`

### 5️⃣ Deploy Automatico

Una volta connesso:
- Ogni `git push` triggera automaticamente il deploy
- Railway buildda e deploya in ~5-10 minuti
- Puoi seguire i log in tempo reale

---

## 📝 Passaggi Rapidi per Connessione MongoDB

Nel nostro `db/mongo.go`, la connessione è già configurata per cercasse il certificato.

Se Railway non trova il file, puoi anche passare il certificato come:
1. Secret in Railway (cifrato)
2. Variabile d'ambiente encodata in base64

---

## 🔗 URL del Servizio

Una volta deployato, Railway ti darà un URL come:
```
https://qr-menu-abc123.railway.app
```

Puoi testare:
```bash
curl https://qr-menu-abc123.railway.app/health
```

---

## Alternative se Railway non funziona

Se Railway crea problemi con il certificato MongoDB, possiamo usare:
1. **Fly.io** - Simile a Railway, molto veloce
2. **Render** - Più stabilito
3. **Google Cloud** - Aspettare che il billing si propaghi

---

## ✅ Continuazione

Registrati su Railway adesso, connetti il repo e fammelo sapere quando il deploy è in corso!
