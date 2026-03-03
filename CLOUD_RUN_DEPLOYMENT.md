# Google Cloud Run Deployment Guide - QR-Menu

## 📋 Setup Google Cloud Project

### STEP 1: Crea/Accedi al Google Cloud Project

```bash
# Se non hai gcloud CLI, installalo da:
# https://cloud.google.com/sdk/docs/install

# Login
gcloud auth login

# Crea nuovo project (o usa uno esistente)
gcloud projects create qr-menu-prod --name="QR Menu Production"

# Imposta project attuale
gcloud config set project qr-menu-prod
```

### STEP 2: Abilita le API necessarie

```bash
gcloud services enable run.googleapis.com
gcloud services enable containerregistry.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

### STEP 3: Configura le variabili d'ambiente

Cloud Run supporta environment variables. Crea un file `.env.yaml`:

```yaml
# environment.yaml per Cloud Run
MONGODB_URI: "mongodb+srv://qr-menu-dev@cluster0.b9jfwmr.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509"
MONGODB_DB_NAME: "qr-menu"
MONGODB_CERT_PATH: "/secrets/X509-cert.pem"
PORT: "8080"
```

**⚠️ IMPORTANTE**: Il certificato X.509 va gestito con Google Secret Manager!

```bash
# Crea secret per il certificato
gcloud secrets create mongodb-x509-cert \
  --data-file=C:/Users/gigli/Desktop/X509-cert-4084673564018728353.pem

# Dai permesso a Cloud Run di accedere al secret
gcloud projects add-iam-policy-binding qr-menu-prod \
  --member=serviceAccount:qr-menu@qr-menu-prod.iam.gserviceaccount.com \
  --role=roles/secretmanager.secretAccessor
```

### STEP 4: Configura il Secret nel main.go

Aggiorna il codice per leggere il certificato dal Secret Manager:

```go
// Nel file db/mongo.go, sostituisci la lettura del certificato locale con:
import "cloud.google.com/go/secretmanager/apiv1"

func getCertificateFromSecret(ctx context.Context) ([]byte, error) {
    hostname := os.Getenv("GOOGLE_CLOUD_PROJECT")
    if hostname == "" {
        // Fallback per local testing
        return ioutil.ReadFile("C:/Users/gigli/Desktop/X509-cert-4084673564018728353.pem")
    }
    
    client, err := secretmanager.NewClient(ctx)
    if err != nil {
        return nil, err
    }
    
    req := &secretmanagerpb.AccessSecretVersionRequest{
        Name: fmt.Sprintf("projects/%s/secrets/mongodb-x509-cert/versions/latest", hostname),
    }
    
    result, err := client.AccessSecretVersion(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return result.Payload.Data, nil
}
```

### STEP 5: Deploy su Cloud Run

```bash
# Posizionati nella cartella del progetto
cd c:\Users\gigli\GoWs\qr-menu

# Deploy (prima volta crea il servizio)
gcloud run deploy qr-menu \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --port 8080 \
  --memory 512Mi \
  --cpu 1 \
  --timeout 3600 \
  --set-env-vars MONGODB_URI="mongodb+srv://qr-menu-dev@cluster0.b9jfwmr.mongodb.net/?authSource=$external&authMechanism=MONGODB-X509",MONGODB_DB_NAME="qr-menu",MONGODB_CERT_PATH="/run/secrets/mongodb-x509-cert"

# Per deployment successivi, puoi usare:
gcloud run deploy qr-menu --source .
```

### STEP 6: Verifica il deployment

```bash
# Check status
gcloud run services describe qr-menu --region us-central1

# Visualizza i log
gcloud run logs read qr-menu --region us-central1 --limit 50

# Test della URL pubblica
curl https://qr-menu-[PROJECT-ID].a.run.app/health
```

## 📊 Costi Stimati (Gratuiti!)

- **Compute**: 2M requests/mese FREE (più che sufficiente per test)
- **Storage**: 5GB FREE per database
- **Network**: 1GB/mese outbound FREE
- **Total**: **$0** (finché non superi i limiti free)

## 🔐 Secret Management

### Gestire il certificato X.509 in produzione:

```bash
# Vedi il secret
gcloud secrets versions list mongodb-x509-cert

# Aggiorna il secret
gcloud secrets versions add mongodb-x509-cert \
  --data-file=/path/to/new/cert.pem

# Nel deployment, Cloud Run automaticamente lo monta in /run/secrets/
```

## 🚨 Troubleshooting

### Certificate validation errors
```
Soluzione: Controlla che il certificato sia correttamente caricato nel Secret Manager
```

### Timeout di connessione MongoDB
```
Soluzione: Aggiungi l'IP di Cloud Run alla whitelist di MongoDB Atlas
```

### Out of quota
```
Soluzione: Richiedi aumento della quota in gcloud console
```

## 📈 Monitoraggio

```bash
# Dashboard in Cloud Console:
# https://console.cloud.google.com/run

# Metriche disponibili:
# - Request rate
# - Error rate
# - Response latency
# - CPU/Memory usage
```

---

**Prossimo step**: Eseguire gli step 1-5 sopra in sequenza, poi verificare il deployment!
