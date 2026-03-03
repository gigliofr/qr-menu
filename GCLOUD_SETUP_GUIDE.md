# 🔧 Setup Google Cloud SDK - QR-Menu Deployment

## ⚙️ Passo 1: Installa Google Cloud SDK

### Opzione A: Installazione Automatica (Consigliato)

#### Su Windows con winget:
```powershell
# Esegui PowerShell come Amministratore

winget install Google.CloudSDK -e
```

#### Su Windows senza winget:
```powershell
# Scarica da: https://cloud.google.com/sdk/docs/install-windows
# Esegui l'installer MSI standard
```

#### Su macOS:
```bash
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

#### Su Linux:
```bash
curl https://sdk.cloud.google.com | bash
exec -l $SHELL
```

---

## ✅ Passo 2: Verifica Installazione

```powershell
gcloud --version
```

Aspettato output:
```
Google Cloud SDK ...
```

---

## 🔐 Passo 3: Autentica con Google Cloud

```powershell
gcloud auth login
```

Questo aprirà il browser per autenticarti. Segui i passaggi:
1. Accedi con il tuo account Google
2. Dai il permesso a gcloud
3. Copia il codice di autenticazione (se richiesto)
4. Torna al terminale

---

## 📁 Passo 4: Crea il Progetto Google Cloud

```powershell
# Crea nuovo project
gcloud projects create qr-menu-prod --name="QR Menu Production"

# Imposta come project attuale
gcloud config set project qr-menu-prod

# Verifica
gcloud config get-value project
```

Aspettato output:
```
qr-menu-prod
```

---

## 🚀 Passo 5: Abilita le API Necessarie

```powershell
# Cloud Run API
gcloud services enable run.googleapis.com

# Container Registry
gcloud services enable containerregistry.googleapis.com

# Cloud Build (per build automatico)
gcloud services enable cloudbuild.googleapis.com

# Secret Manager (per certificato X.509)
gcloud services enable secretmanager.googleapis.com
```

Aspettato output:
```
Operation "operations/..." finished successfully.
```

---

## 🔑 Passo 6: Configura il Certificato X.509 nel Secret Manager

```powershell
# Carica il certificato come secret
gcloud secrets create mongodb-x509-cert `
  --replication-policy="automatic" `
  --data-file="C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem"

# Verifica
gcloud secrets describe mongodb-x509-cert
```

Aspettato output:
```
Created version [1] of the secret [mongodb-x509-cert].
```

---

## ✨ Passo 7: Crea un Service Account (opzionale ma consigliato)

```powershell
# Crea service account
gcloud iam service-accounts create qr-menu-runner `
  --display-name="QR Menu Cloud Run Runtime"

# Dai permesso di accedere al secret
gcloud secrets add-iam-policy-binding mongodb-x509-cert `
  --member="serviceAccount:qr-menu-runner@qr-menu-prod.iam.gserviceaccount.com" `
  --role="roles/secretmanager.secretAccessor"
```

---

## ✅ Verifica Setup

```powershell
# Mostra configurazione attuale
gcloud config list

# Controlla progetti disponibili
gcloud projects list

# Controlla API abilitate
gcloud services list --enabled
```

Aspettato output per ogni comando:
```
✅ project = qr-menu-prod
✅ run.googleapis.com ENABLED
✅ secretmanager.googleapis.com ENABLED
```

---

## 🎯 Risultato Finale

Se tutto è verde ✅:
- ✅ gcloud CLI installato
- ✅ Autenticato con Google Cloud
- ✅ Project "qr-menu-prod" creato
- ✅ API abilitate
- ✅ Certificato X.509 caricato nel Secret Manager
- ✅ Service Account configurato

**Pronto per Task 4: Deploy su Cloud Run! 🚀**

---

## 🚨 Troubleshooting

### "gcloud command not found"
- Riavvia PowerShell/terminal dopo l'installazione
- Verifica il PATH: `$env:PATH`

### "Access Denied" su secrets
- Assicurati di aver eseguito come Admin
- Verifica IAM permissions nel Google Cloud console

### "Project creation failed"
- Verifica di avere una billing account attiva
- Usa `gcloud projects list` per vedere progetti esistenti

---

## 📞 Link Utili

- Cloud SDK Docs: https://cloud.google.com/sdk/docs
- Cloud Run Docs: https://cloud.google.com/run/docs
- Secret Manager: https://cloud.google.com/secret-manager/docs
- gcloud reference: https://cloud.google.com/sdk/gcloud/reference

---

**Status: Pronto per procedere?** ✅
