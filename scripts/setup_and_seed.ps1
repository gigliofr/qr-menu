# ===================================================================
# Script Helper - Setup e Seed Database di Test
# ===================================================================
# 
# Questo script configura MongoDB e popola il database con dati di test
#

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "🧹 SETUP DATABASE DI TEST" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# ===================================================================
# STEP 1: VERIFICA CONFIGURAZIONE MONGODB
# ===================================================================

Write-Host "📋 Verifica configurazione MongoDB..." -ForegroundColor Yellow
Write-Host ""

$hasMongoUri = $env:MONGODB_URI -ne $null
$hasMongoCert = ($env:MONGODB_CERT_PATH -ne $null) -or ($env:MONGODB_CERT_CONTENT -ne $null)
$hasMongoDb = $env:MONGODB_DB_NAME -ne $null

if (-not $hasMongoUri -or -not $hasMongoCert) {
    Write-Host "⚠️  MongoDB non configurato completamente" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Per procedere, configura le seguenti variabili d'ambiente:" -ForegroundColor White
    Write-Host ""
    Write-Host "Opzione A - Setup Temporaneo (solo questa sessione):" -ForegroundColor Cyan
    Write-Host '  $env:MONGODB_URI="mongodb+srv://your-cluster.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509"'
    Write-Host '  $env:MONGODB_CERT_PATH="C:\path\to\X509-cert.pem"'
    Write-Host '  $env:MONGODB_DB_NAME="qr-menu"'
    Write-Host ""
    Write-Host "Opzione B - Setup Permanente:" -ForegroundColor Cyan
    Write-Host '  [System.Environment]::SetEnvironmentVariable("MONGODB_URI", "mongodb+srv://...", "User")'
    Write-Host '  [System.Environment]::SetEnvironmentVariable("MONGODB_CERT_PATH", "C:\path\to\cert.pem", "User")'
    Write-Host '  [System.Environment]::SetEnvironmentVariable("MONGODB_DB_NAME", "qr-menu", "User")'
    Write-Host ""
    
    $response = Read-Host "Vuoi configurare ora? (s/n)"
    
    if ($response -eq "s" -or $response -eq "S") {
        Write-Host ""
        Write-Host "📝 Configurazione MongoDB:" -ForegroundColor Cyan
        Write-Host ""
        
        $mongoUri = Read-Host "MongoDB URI"
        $mongoCertPath = Read-Host "Percorso certificato X509 (.pem)"
        $mongoDbName = Read-Host "Nome database (default: qr-menu)"
        
        if ([string]::IsNullOrWhiteSpace($mongoDbName)) {
            $mongoDbName = "qr-menu"
        }
        
        # Verifica che il certificato esista
        if (-not (Test-Path $mongoCertPath)) {
            Write-Host ""
            Write-Host "❌ Errore: Il certificato non esiste in: $mongoCertPath" -ForegroundColor Red
            Write-Host ""
            Write-Host "💡 Scarica il certificato da:" -ForegroundColor Yellow
            Write-Host "   MongoDB Atlas → Database Access → Certificates → Download X.509 Certificate" -ForegroundColor Yellow
            Write-Host ""
            exit 1
        }
        
        # Imposta variabili per questa sessione
        $env:MONGODB_URI = $mongoUri
        $env:MONGODB_CERT_PATH = $mongoCertPath
        $env:MONGODB_DB_NAME = $mongoDbName
        
        Write-Host ""
        Write-Host "✅ Configurazione impostata per questa sessione" -ForegroundColor Green
        Write-Host ""
        
        $permanent = Read-Host "Vuoi salvare in modo permanente? (s/n)"
        if ($permanent -eq "s" -or $permanent -eq "S") {
            [System.Environment]::SetEnvironmentVariable("MONGODB_URI", $mongoUri, "User")
            [System.Environment]::SetEnvironmentVariable("MONGODB_CERT_PATH", $mongoCertPath, "User")
            [System.Environment]::SetEnvironmentVariable("MONGODB_DB_NAME", $mongoDbName, "User")
            Write-Host "✅ Configurazione salvata permanentemente" -ForegroundColor Green
            Write-Host "   (Riavvia il terminale per usarla in altre sessioni)" -ForegroundColor Yellow
        }
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "❌ Setup annullato" -ForegroundColor Red
        Write-Host "   Configura MongoDB e riprova" -ForegroundColor Yellow
        Write-Host ""
        exit 1
    }
} else {
    Write-Host "✅ MongoDB configurato correttamente" -ForegroundColor Green
    Write-Host "   URI:       $($env:MONGODB_URI.Substring(0, 40))..." -ForegroundColor Gray
    Write-Host "   Cert:      $env:MONGODB_CERT_PATH" -ForegroundColor Gray
    Write-Host "   Database:  $env:MONGODB_DB_NAME" -ForegroundColor Gray
    Write-Host ""
}

# ===================================================================
# STEP 2: VERIFICA MONGOSH
# ===================================================================

Write-Host "🔍 Verifica mongosh (MongoDB Shell)..." -ForegroundColor Yellow

$mongoshPath = Get-Command mongosh -ErrorAction SilentlyContinue

if (-not $mongoshPath) {
    Write-Host "❌ mongosh non trovato nel PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "📥 Installa MongoDB Shell:" -ForegroundColor Yellow
    Write-Host "   1. Vai su: https://www.mongodb.com/try/download/shell" -ForegroundColor White
    Write-Host "   2. Scarica MongoDB Shell (mongosh) per Windows" -ForegroundColor White
    Write-Host "   3. Estrai e aggiungi al PATH" -ForegroundColor White
    Write-Host ""
    Write-Host "OPPURE usa Chocolatey:" -ForegroundColor Yellow
    Write-Host "   choco install mongodb-shell" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host "✅ mongosh trovato: $($mongoshPath.Source)" -ForegroundColor Green
Write-Host ""

# ===================================================================
# STEP 3: ESECUZIONE SCRIPT SEED
# ===================================================================

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "🌱 ESECUZIONE SEED DATABASE" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "⚠️  ATTENZIONE: Questo script:" -ForegroundColor Yellow
Write-Host "   - Eliminerà TUTTI i dati esistenti" -ForegroundColor Yellow
Write-Host "   - Creerà 1 utente admin" -ForegroundColor Yellow
Write-Host "   - Creerà 4 ristoranti di test" -ForegroundColor Yellow
Write-Host "   - Creerà 37 piatti totali nei menu" -ForegroundColor Yellow
Write-Host ""

$confirm = Read-Host "Vuoi procedere? (s/n)"

if ($confirm -ne "s" -and $confirm -ne "S") {
    Write-Host ""
    Write-Host "❌ Seed annullato" -ForegroundColor Red
    Write-Host ""
    exit 0
}

Write-Host ""
Write-Host "🚀 Esecuzione seed script..." -ForegroundColor Cyan
Write-Host ""

# Crea URI completo con certificato per mongosh
$mongoshUri = $env:MONGODB_URI

try {
    # Esegui lo script MongoDB
    $result = & mongosh $mongoshUri `
        --tls `
        --tlsCertificateKeyFile $env:MONGODB_CERT_PATH `
        --file "scripts\seed_test_data.js" `
        2>&1
    
    # Mostra output
    $result | ForEach-Object { Write-Host $_ }
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "================================================" -ForegroundColor Green
        Write-Host "🎉 SEED COMPLETATO CON SUCCESSO!" -ForegroundColor Green
        Write-Host "================================================" -ForegroundColor Green
        Write-Host ""
        Write-Host "🔐 CREDENZIALI AMMINISTRATORE:" -ForegroundColor Cyan
        Write-Host "   Username: admin" -ForegroundColor White
        Write-Host "   Password: admin123" -ForegroundColor White
        Write-Host "   Email:    admin@qrmenu.local" -ForegroundColor White
        Write-Host ""
        Write-Host "🏪 RISTORANTI CREATI:" -ForegroundColor Cyan
        Write-Host "   1. Pizzeria Napoletana    (10 piatti + bevande)" -ForegroundColor White
        Write-Host "   2. Trattoria Toscana      (7 piatti toscani)" -ForegroundColor White
        Write-Host "   3. Sushi-Ya Tokyo         (10 specialità giapponesi)" -ForegroundColor White
        Write-Host "   4. Burger House Americana (10 burger + contorni)" -ForegroundColor White
        Write-Host ""
        Write-Host "🚀 PROSSIMI PASSI:" -ForegroundColor Yellow
        Write-Host "   1. Avvia applicazione:  .\qr-menu.exe" -ForegroundColor White
        Write-Host "   2. Apri browser:        http://localhost:8080" -ForegroundColor White
        Write-Host "   3. Login:               admin / admin123" -ForegroundColor White
        Write-Host "   4. Seleziona ristorante da /select-restaurant" -ForegroundColor White
        Write-Host ""
    } else {
        Write-Host ""
        Write-Host "❌ Errore durante il seed" -ForegroundColor Red
        Write-Host "   Exit code: $LASTEXITCODE" -ForegroundColor Red
        Write-Host ""
        Write-Host "💡 Possibili cause:" -ForegroundColor Yellow
        Write-Host "   - Credenziali MongoDB errate" -ForegroundColor White
        Write-Host "   - Certificato X509 non valido" -ForegroundColor White
        Write-Host "   - Problemi di rete con MongoDB Atlas" -ForegroundColor White
        Write-Host ""
    }
    
} catch {
    Write-Host ""
    Write-Host "❌ Errore durante l'esecuzione:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    exit 1
}

Write-Host ""
