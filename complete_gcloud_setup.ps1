# 🚀 Complete Google Cloud Setup - Step 2-7
# Run this script AFTER restarting PowerShell

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host "🚀 Google Cloud Complete Setup - QR-Menu Deployment" -ForegroundColor Cyan
Write-Host "============================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Verifica gcloud
Write-Host "📋 STEP 1: Verifica gcloud..." -ForegroundColor Yellow
gcloud --version
Write-Host ""

# Step 2: Autentica
Write-Host "🔐 STEP 2: Autenticazione Google Cloud..." -ForegroundColor Yellow
Write-Host "   → Si aprirà un browser per l'autenticazione" -ForegroundColor Gray
Write-Host "   → Accedi con il tuo account Google" -ForegroundColor Gray
Write-Host "   → Torna qui e premi invio" -ForegroundColor Gray
Read-Host "Premi INVIO quando sei pronto"

gcloud auth login

Write-Host ""
Write-Host "✅ Autenticazione completata!" -ForegroundColor Green
Write-Host ""

# Step 3: Crea project
Write-Host "📁 STEP 3: Creo il project Google Cloud..." -ForegroundColor Yellow
gcloud projects create qr-menu-prod --name="QR Menu Production"
gcloud config set project qr-menu-prod

Write-Host "✅ Project creato!" -ForegroundColor Green
Write-Host ""

# Step 4: Abilita API
Write-Host "⚙️  STEP 4: Abilito le API necessarie..." -ForegroundColor Yellow
Write-Host "   → Cloud Run" -ForegroundColor Gray
gcloud services enable run.googleapis.com

Write-Host "   → Container Registry" -ForegroundColor Gray
gcloud services enable containerregistry.googleapis.com

Write-Host "   → Cloud Build" -ForegroundColor Gray
gcloud services enable cloudbuild.googleapis.com

Write-Host "   → Secret Manager" -ForegroundColor Gray
gcloud services enable secretmanager.googleapis.com

Write-Host "✅ API abilitate!" -ForegroundColor Green
Write-Host ""

# Step 5: Upload certificato
Write-Host "🔑 STEP 5: Carico il certificato X.509..." -ForegroundColor Yellow
$certPath = "C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem"

if (Test-Path $certPath) {
    gcloud secrets create mongodb-x509-cert `
        --replication-policy="automatic" `
        --data-file=$certPath
    Write-Host "✅ Certificato caricato!" -ForegroundColor Green
} else {
    Write-Host "❌ Certificato non trovato in: $certPath" -ForegroundColor Red
    Write-Host "   Scaricalo da MongoDB Atlas e riprova" -ForegroundColor Yellow
}
Write-Host ""

# Step 6: Crea service account
Write-Host "👤 STEP 6: Creo il Service Account..." -ForegroundColor Yellow
gcloud iam service-accounts create qr-menu-runner `
    --display-name="QR Menu Cloud Run Runtime"

# Step 7: Assegna permessi
Write-Host "🔐 STEP 7: Assegno i permessi..." -ForegroundColor Yellow
gcloud secrets add-iam-policy-binding mongodb-x509-cert `
    --member="serviceAccount:qr-menu-runner@qr-menu-prod.iam.gserviceaccount.com" `
    --role="roles/secretmanager.secretAccessor"

Write-Host "✅ Service Account configurato!" -ForegroundColor Green
Write-Host ""

# Verifica finale
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host "✅ SETUP COMPLETATO CON SUCCESSO!" -ForegroundColor Green
Write-Host "===========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Configurazione:" -ForegroundColor Yellow
gcloud config list

Write-Host ""
Write-Host "🚀 PROSSIMO STEP: Task 4 - Deploy su Cloud Run" -ForegroundColor Cyan
Write-Host ""
Write-Host "Esegui:" -ForegroundColor Yellow
Write-Host "  cd C:\Users\gigli\GoWs\qr-menu"
Write-Host "  gcloud run deploy"
Write-Host ""

Read-Host "Premi INVIO per continuare"
