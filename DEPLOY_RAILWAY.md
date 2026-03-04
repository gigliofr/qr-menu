# 🚀 Deploy su Railway - Guida Rapida

## ✅ Pre-requisiti

- Account Railway: https://railway.app (gratuito)
- Repository GitHub già pushato ✅

## 📋 Passaggi Deploy (5 minuti)

### 1. Crea Nuovo Progetto Railway

1. Vai su https://railway.app
2. Click **"New Project"**
3. Scegli **"Deploy from GitHub repo"**
4. Seleziona: **gigliofr/qr-menu**
5. Click **"Deploy Now"**

### 2. Configura Variabili d'Ambiente

Nel progetto Railway appena creato:

1. Vai su **"Variables"** (tab in alto)
2. Click **"+ New Variable"**
3. Aggiungi queste 4 variabili:

#### Variabile 1: MONGODB_URI
```
mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authMechanism=MONGODB-X509&authSource=$external&retryWrites=true&w=majority
```

#### Variabile 2: MONGODB_DB_NAME
```
qr-menu
```

#### Variabile 3: MONGODB_CERT_CONTENT
```
-----BEGIN CERTIFICATE-----
<Copia TUTTO il contenuto del file>
C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem
<Includi BEGIN e END CERTIFICATE>
-----END CERTIFICATE-----
```

#### Variabile 4: PORT
```
8080
```

### 3. Deploy Automatico

Railway rileva automaticamente il `Dockerfile` e fa il build!

- ⏳ Build: ~2-3 minuti
- ✅ Deploy: automatico
- 🌐 URL pubblico: generato automaticamente

### 4. Ottieni URL Pubblico

1. Nel progetto Railway, vai su **"Settings"**
2. Sezione **"Domains"**
3. Click **"Generate Domain"**
4. Copia l'URL (es: `qr-menu-production.up.railway.app`)

### 5. Test Online

Vai su:
```
https://<tuo-dominio>.up.railway.app/login
```

Credenziali:
- **Username:** admin
- **Password:** admin123

## 🎯 Alternativa: Deploy da CLI

Se hai Railway CLI installato:

```bash
# Installa Railway CLI
npm i -g @railway/cli

# Login
railway login

# Deploy
railway up

# Configura env vars
railway variables set MONGODB_URI="mongodb+srv://..."
railway variables set MONGODB_DB_NAME="qr-menu"
railway variables set MONGODB_CERT_CONTENT="$(cat C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem)"
railway variables set PORT="8080"
```

## 📊 Monitoraggio

Railway Dashboard mostra:
- 📈 Deployment logs
- 💾 Uso risorse
- 🌐 Traffico
- 🐛 Errori runtime

## 🔧 Troubleshooting

### Build Failed
- Verifica che `Dockerfile` sia presente
- Controlla i logs di build su Railway

### MongoDB Connection Error
- Verifica `MONGODB_URI` (deve includere `authMechanism=MONGODB-X509`)
- Controlla `MONGODB_CERT_CONTENT` (deve avere BEGIN/END CERTIFICATE)
- MongoDB Atlas IP whitelist: aggiungi `0.0.0.0/0` per Railway

### App Non Risponde
- Verifica `PORT=8080` nelle variabili
- Controlla deployment logs su Railway

## 💰 Costi

Piano gratuito Railway:
- ✅ 500 ore/mese
- ✅ 100GB traffico
- ✅ Sufficiente per test e demo

---

## 🎉 Deploy Completato

Una volta online:

1. **Test Login:** https://tuo-dominio.railway.app/login
2. **Selezione Ristorante:** 4 ristoranti di test disponibili
3. **Admin Panel:** Gestione menu completa
4. **QR Code:** Genera QR per ogni ristorante

**Condividi l'URL pubblico per far testare l'app! 🚀**
