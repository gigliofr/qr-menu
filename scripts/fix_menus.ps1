# ===================================================================
# Script PowerShell per Correggere Menu nel Database
# ===================================================================

$ErrorActionPreference = "Stop"

Write-Host "`n================================================" -ForegroundColor Cyan
Write-Host "🔧 FIX MENU - Correzione Struttura Menu" -ForegroundColor Cyan
Write-Host "================================================`n" -ForegroundColor Cyan

# Path al certificato MongoDB
$certPath = "C:\Users\gigli\Desktop\X509-cert-4084673564018728353.pem"

# MongoDB URI
$mongoUri = "mongodb+srv://ac-d8zdak4.b9jfwmr.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=$certPath"

# Path allo script
$scriptPath = Join-Path $PSScriptRoot "fix_menus.js"

if (-not (Test-Path $scriptPath)) {
    Write-Host "❌ Errore: Script fix_menus.js non trovato" -ForegroundColor Red
    exit 1
}

Write-Host "📂 Script: $scriptPath" -ForegroundColor Yellow
Write-Host "🔐 Certificato: $certPath`n" -ForegroundColor Yellow

# Esegui lo script con mongosh
Write-Host "🚀 Esecuzione script di fix...`n" -ForegroundColor Green

try {
    & mongosh $mongoUri --quiet --file $scriptPath
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "`n✅ Menu corretti con successo!" -ForegroundColor Green
        Write-Host "`nOra puoi:" -ForegroundColor Cyan
        Write-Host "  1. Fare login su https://qr-menu-staging.up.railway.app/login" -ForegroundColor White
        Write-Host "  2. Username: admin  Password: admin" -ForegroundColor White
        Write-Host "  3. Selezionare un ristorante" -ForegroundColor White
        Write-Host "  4. Vedere i menu e generare QR code`n" -ForegroundColor White
    } else {
        Write-Host "`n❌ Errore durante l'esecuzione dello script" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "`n❌ Errore: $_" -ForegroundColor Red
    exit 1
}
