# 🧪 Test delle Nuove Funzionalità

# Test 1: Verifica che il server sia avviato
Write-Host "🔍 Verifica server..." -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080" -UseBasicParsing -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Server online e funzionante" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Server non raggiungibile" -ForegroundColor Red
    Write-Host "Assicurati che .\qr-menu.exe sia in esecuzione" -ForegroundColor Yellow
    exit 1
}

# Test 2: Verifica template moderni
Write-Host "`n🎨 Verifica template moderni..." -ForegroundColor Cyan
$adminResponse = Invoke-WebRequest -Uri "http://localhost:8080/login" -UseBasicParsing
if ($adminResponse.Content -match "glass-morphism|backdrop-filter|Inter") {
    Write-Host "✅ Template moderni caricati correttamente" -ForegroundColor Green
} else {
    Write-Host "⚠️  Template potrebbero non essere aggiornati" -ForegroundColor Yellow
}

# Test 3: Verifica header di sicurezza
Write-Host "`n🛡️ Verifica header di sicurezza..." -ForegroundColor Cyan
$headers = (Invoke-WebRequest -Uri "http://localhost:8080/login" -UseBasicParsing).Headers
$securityHeaders = @(
    "X-Content-Type-Options",
    "X-Frame-Options", 
    "X-XSS-Protection",
    "Content-Security-Policy"
)

foreach ($header in $securityHeaders) {
    if ($headers.ContainsKey($header)) {
        Write-Host "✅ $header configurato" -ForegroundColor Green
    } else {
        Write-Host "❌ $header mancante" -ForegroundColor Red
    }
}

# Test 4: Verifica directory per immagini
Write-Host "`n📁 Verifica directory..." -ForegroundColor Cyan
$directories = @(
    "storage",
    "static/images",
    "static/images/dishes",
    "static/qrcodes"
)

foreach ($dir in $directories) {
    if (Test-Path $dir) {
        Write-Host "✅ Directory $dir esistente" -ForegroundColor Green
    } else {
        Write-Host "❌ Directory $dir mancante" -ForegroundColor Red
    }
}

# Test 5: Verifica endpoints API
Write-Host "`n🔗 Verifica endpoint API..." -ForegroundColor Cyan
$endpoints = @(
    "/login",
    "/register", 
    "/admin"
)

foreach ($endpoint in $endpoints) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080$endpoint" -UseBasicParsing -TimeoutSec 3
        if ($response.StatusCode -eq 200) {
            Write-Host "✅ Endpoint $endpoint accessibile" -ForegroundColor Green
        }
    } catch {
        Write-Host "❌ Endpoint $endpoint non raggiungibile" -ForegroundColor Red
    }
}

# Apertura interfacce per test manuale
Write-Host "`n🚀 Apertura interfacce per test manuale..." -ForegroundColor Cyan
Write-Host "Aprendo le seguenti pagine nel browser:" -ForegroundColor Yellow
Write-Host "- Login: http://localhost:8080/login" -ForegroundColor White
Write-Host "- Registrazione: http://localhost:8080/register" -ForegroundColor White
Write-Host "- Admin: http://localhost:8080/admin" -ForegroundColor White

# Apri le pagine nel browser
Start-Process "http://localhost:8080/login"
Start-Sleep 2
Start-Process "http://localhost:8080/register"

Write-Host "`n🎉 Test completati!" -ForegroundColor Green
Write-Host "Verifica manualmente:" -ForegroundColor Cyan
Write-Host "1. 📱 Registra un nuovo ristorante" -ForegroundColor White
Write-Host "2. 🍽️  Crea un menu con piatti" -ForegroundColor White  
Write-Host "3. 📸 Carica immagini per i piatti" -ForegroundColor White
Write-Host "4. 📱 Testa la condivisione social" -ForegroundColor White
Write-Host "5. 🎨 Verifica il design moderno" -ForegroundColor White