# 🚂 Railway Quick Setup Script
# Genera SESSION_SECRET e configura Railway environment variables

Write-Host "`n🚀 QR Menu - Railway Setup Wizard`n" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray

# Step 1: Generate SESSION_SECRET
Write-Host "📝 Step 1: Generating SESSION_SECRET..." -ForegroundColor Yellow

$sessionSecret = -join ((1..32) | ForEach-Object {'{0:x2}' -f (Get-Random -Maximum 256)})

Write-Host "`n✅ SESSION_SECRET generato:`n" -ForegroundColor Green
Write-Host "   $sessionSecret`n" -ForegroundColor White
Write-Host "   ⚠️  SALVA QUESTO VALORE! Te lo servirà per Railway`n" -ForegroundColor Yellow

# Step 2: Check Railway CLI
Write-Host "📦 Step 2: Verifying Railway CLI..." -ForegroundColor Yellow

$railwayCli = Get-Command railway -ErrorAction SilentlyContinue

if (-not $railwayCli) {
    Write-Host "`n❌ Railway CLI non trovato!`n" -ForegroundColor Red
    Write-Host "   Installazione richiesta:" -ForegroundColor Yellow
    Write-Host "   npm install -g @railway/cli`n" -ForegroundColor Cyan
    Write-Host "   Oppure configura manualmente su Railway Dashboard:" -ForegroundColor Yellow
    Write-Host "   https://railway.app/dashboard`n" -ForegroundColor Cyan
    
    # Copy to clipboard
    Set-Clipboard -Value $sessionSecret
    Write-Host "✅ SESSION_SECRET copiato negli appunti!" -ForegroundColor Green
    Write-Host "   Incolla su Railway Dashboard → Variables → SESSION_SECRET`n" -ForegroundColor White
    
    exit 0
}

Write-Host "✅ Railway CLI trovato: $($railwayCli.Source)`n" -ForegroundColor Green

# Step 3: Check if logged in
Write-Host "🔐 Step 3: Checking Railway authentication..." -ForegroundColor Yellow

railway whoami 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "`n⚠️  Non sei loggato a Railway`n" -ForegroundColor Yellow
    Write-Host "   Esegui: railway login`n" -ForegroundColor Cyan
    
    $login = Read-Host "Vuoi fare login ora? (y/n)"
    if ($login -eq 'y' -or $login -eq 'Y') {
        railway login
        if ($LASTEXITCODE -ne 0) {
            Write-Host "`n❌ Login fallito" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "`n❌ Setup annullato" -ForegroundColor Red
        exit 0
    }
}

Write-Host "✅ Autenticato su Railway`n" -ForegroundColor Green

# Step 4: Check project link
Write-Host "🔗 Step 4: Checking project link..." -ForegroundColor Yellow

railway status 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "`n⚠️  Nessun progetto Railway linkato in questa directory`n" -ForegroundColor Yellow
    Write-Host "   Opzioni:" -ForegroundColor Yellow
    Write-Host "   1. Link progetto esistente: railway link" -ForegroundColor Cyan
    Write-Host "   2. Crea nuovo progetto: railway init`n" -ForegroundColor Cyan
    
    $action = Read-Host "Vuoi linkare un progetto esistente? (y/n)"
    if ($action -eq 'y' -or $action -eq 'Y') {
        railway link
        if ($LASTEXITCODE -ne 0) {
            Write-Host "`n❌ Link fallito" -ForegroundColor Red
            exit 1
        }
    } else {
        Write-Host "`n❌ Setup annullato - configura manualmente" -ForegroundColor Red
        Write-Host "   SESSION_SECRET copiato negli appunti" -ForegroundColor Yellow
        Set-Clipboard -Value $sessionSecret
        exit 0
    }
}

Write-Host "✅ Progetto Railway linkato`n" -ForegroundColor Green

# Step 5: Environment selection
Write-Host "🌍 Step 5: Select environment..." -ForegroundColor Yellow
Write-Host "`n   Quale ambiente stai configurando?" -ForegroundColor White
Write-Host "   1. Production (ENVIRONMENT=production, LOG_LEVEL=INFO)" -ForegroundColor Cyan
Write-Host "   2. Staging (ENVIRONMENT=staging, LOG_LEVEL=DEBUG)" -ForegroundColor Yellow
Write-Host "   3. Development (ENVIRONMENT=development, LOG_LEVEL=DEBUG)`n" -ForegroundColor Green

$envChoice = Read-Host "Scelta (1/2/3)"

switch ($envChoice) {
    "1" {
        $environment = "production"
        $logLevel = "INFO"
        $seedData = "false"
    }
    "2" {
        $environment = "staging"
        $logLevel = "DEBUG"
        $seedData = "true"
    }
    "3" {
        $environment = "development"
        $logLevel = "DEBUG"
        $seedData = "true"
    }
    default {
        Write-Host "`n❌ Scelta non valida" -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n✅ Ambiente selezionato: $environment`n" -ForegroundColor Green

# Step 6: Set variables
Write-Host "⚙️  Step 6: Setting variables on Railway...`n" -ForegroundColor Yellow

$variables = @{
    "SESSION_SECRET" = $sessionSecret
    "ENVIRONMENT" = $environment
    "LOG_LEVEL" = $logLevel
    "ENABLE_SEED_DATA" = $seedData
}

foreach ($key in $variables.Keys) {
    $value = $variables[$key]
    Write-Host "   Setting $key = $value..." -ForegroundColor Cyan
    
    railway variables set "$key=$value" 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ $key impostato" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  Errore setting $key (potrebbe già esistere)" -ForegroundColor Yellow
    }
}

Write-Host ""

# Step 7: Verify MongoDB variables
Write-Host "🗄️  Step 7: Verifying MongoDB variables..." -ForegroundColor Yellow

$requiredVars = @("MONGODB_URI", "MONGODB_CERT_CONTENT", "MONGODB_DB_NAME")
$missingVars = @()

foreach ($var in $requiredVars) {
    railway variables 2>$null | Select-String $var >$null
    if ($LASTEXITCODE -ne 0) {
        $missingVars += $var
    }
}

if ($missingVars.Count -gt 0) {
    Write-Host "`n⚠️  ATTENZIONE: Variabili MongoDB mancanti:`n" -ForegroundColor Yellow
    foreach ($var in $missingVars) {
        Write-Host "   ❌ $var" -ForegroundColor Red
    }
    Write-Host "`n   Configura manualmente su Railway Dashboard:" -ForegroundColor Yellow
    Write-Host "   Vedi: RAILWAY_SETUP_GUIDE.md per dettagli`n" -ForegroundColor Cyan
} else {
    Write-Host "`n✅ Tutte le variabili MongoDB presenti`n" -ForegroundColor Green
}

# Step 8: Summary
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor DarkGray
Write-Host "`n✅ SETUP COMPLETATO!`n" -ForegroundColor Green

Write-Host "📊 Variabili configurate:" -ForegroundColor Cyan
Write-Host "   • SESSION_SECRET: ✅ (64 char hex)" -ForegroundColor White
Write-Host "   • ENVIRONMENT: $environment" -ForegroundColor White
Write-Host "   • LOG_LEVEL: $logLevel" -ForegroundColor White
Write-Host "   • ENABLE_SEED_DATA: $seedData`n" -ForegroundColor White

Write-Host "🚀 Prossimi step:" -ForegroundColor Yellow
Write-Host "   1. Verifica variabili MongoDB (se mancanti)" -ForegroundColor White
Write-Host "   2. Deploy/Redeploy Railway:" -ForegroundColor White
Write-Host "      railway up" -ForegroundColor Cyan
Write-Host "      # O push su GitHub per auto-deploy`n" -ForegroundColor DarkGray

Write-Host "   3. Monitor logs:" -ForegroundColor White
Write-Host "      railway logs --follow`n" -ForegroundColor Cyan

Write-Host "   4. Test health check:" -ForegroundColor White
Write-Host "      Invoke-RestMethod 'https://your-domain.up.railway.app/api/v1/health'`n" -ForegroundColor Cyan

Write-Host "📚 Documentazione completa: RAILWAY_SETUP_GUIDE.md`n" -ForegroundColor White

# Save config to file
$configFile = "railway_config_$environment.txt"
@"
Railway Configuration - $environment
Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

SESSION_SECRET=$sessionSecret
ENVIRONMENT=$environment
LOG_LEVEL=$logLevel
ENABLE_SEED_DATA=$seedData

MongoDB Variables Required:
- MONGODB_URI
- MONGODB_CERT_CONTENT
- MONGODB_DB_NAME

Note: Keep this file SECRET and SECURE!
Do NOT commit to repository!
"@ | Out-File $configFile -Encoding UTF8

Write-Host "💾 Config salvato in: $configFile" -ForegroundColor Green
Write-Host "   ⚠️  NON committare questo file nel repository!`n" -ForegroundColor Yellow

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor DarkGray
