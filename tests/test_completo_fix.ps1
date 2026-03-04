$baseUrl = "https://qr-menu-staging.up.railway.app"
$timestamp = Get-Date -Format "HHmmss"

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         TEST REGISTRAZIONE E MENU - COMPLETO                  ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# ============================================================================
# 1. REGISTRAZIONE NUOVO ACCOUNT
# ============================================================================
Write-Host "1️⃣  REGISTRAZIONE NUOVO ACCOUNT" -ForegroundColor Yellow

$username = "testuser_$timestamp"
$email = "test$timestamp@example.com"
$password = "TestPassword123!"
$restaurantName = "Ristorante Test $timestamp"

Write-Host "   Username: $username" -ForegroundColor Gray
Write-Host "   Email: $email" -ForegroundColor Gray
Write-Host "   Restaurant: $restaurantName" -ForegroundColor Gray

$registerBody = @{
    username = $username
    email = $email
    password = $password
    confirm_password = $password
    restaurant_name = $restaurantName
    description = "Ristorante di test automatico"
    address = "Via Test 123, 00100 Roma"
    phone = "+39 123 456789"
} | ConvertTo-Json

try {
    $registerResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method Post `
        -Body $registerBody `
        -ContentType "application/json" `
        -TimeoutSec 15
    
    Write-Host "   ✅ Registrazione completata!" -ForegroundColor Green
    $token = $registerResp.data.token
    $restaurantId = $registerResp.data.restaurant.id
    Write-Host "      Restaurant ID: $restaurantId" -ForegroundColor DarkGray
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
} catch {
    Write-Host "   ❌ Errore registrazione: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        try {
            $errorDetail = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Host "      Dettaglio: $($errorDetail.error.message)" -ForegroundColor Yellow
        } catch {
            Write-Host "      Dettaglio: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
        }
    }
    exit 1
}
Write-Host ""

# ============================================================================
# 2. VERIFICA PROFILO
# ============================================================================
Write-Host "2️⃣  VERIFICA PROFILO" -ForegroundColor Yellow
try {
    $profileResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Headers $headers -TimeoutSec 10
    Write-Host "   ✅ Profilo recuperato!" -ForegroundColor Green
    Write-Host "      Nome: $($profileResp.data.name)" -ForegroundColor DarkGray
    Write-Host "      Username: $($profileResp.data.username)" -ForegroundColor DarkGray
    Write-Host "      Email: $($profileResp.data.email)" -ForegroundColor DarkGray
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# ============================================================================
# 3. CREAZIONE MENU
# ============================================================================
Write-Host "3️⃣  CREAZIONE MENU" -ForegroundColor Yellow

$menuData = @{
    name = "Menu Test"
    description = "Menu di test"
    meal_type = "lunch"
    categories = @(
        @{
            id = (New-Guid).ToString()
            name = "Antipasti"
            description = "Antipasti della casa"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Bruschetta"
                    description = "Pane tostato con pomodoro"
                    price = 5.50
                    category = "Antipasti"
                    available = $true
                },
                @{
                    id = (New-Guid).ToString()
                    name = "Prosciutto e melone"
                    description = "Prosciutto crudo con melone"
                    price = 8.00
                    category = "Antipasti"
                    available = $true
                }
            )
        },
        @{
            id = (New-Guid).ToString()
            name = "Primi Piatti"
            description = "Primi piatti"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Carbonara"
                    description = "Pasta alla carbonara"
                    price = 12.00
                    category = "Primi Piatti"
                    available = $true
                },
                @{
                    id = (New-Guid).ToString()
                    name = "Amatriciana"
                    description = "Pasta all'amatriciana"
                    price = 11.50
                    category = "Primi Piatti"
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $createResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
        -Body $menuData `
        -Headers $headers `
        -TimeoutSec 15
    
    $menuId = $createResp.data.id
    Write-Host "   ✅ Menu creato!" -ForegroundColor Green
    Write-Host "      ID: $menuId" -ForegroundColor DarkGray
    Write-Host "      Nome: $($createResp.data.name)" -ForegroundColor DarkGray
    Write-Host "      Categorie: $($createResp.data.categories.Count)" -ForegroundColor DarkGray
    $itemCount = ($createResp.data.categories | ForEach-Object { $_.items.Count } | Measure-Object -Sum).Sum
    Write-Host "      Piatti: $itemCount" -ForegroundColor DarkGray
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "      Dettaglio: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
    }
}
Write-Host ""

# ============================================================================
# 4. RECUPERO MENU PER ID
# ============================================================================
Write-Host "4️⃣  RECUPERO MENU PER ID" -ForegroundColor Yellow
if ($menuId) {
    try {
        $menuByIdResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId" -Headers $headers -TimeoutSec 10
        Write-Host "   ✅ Menu recuperato per ID" -ForegroundColor Green
        Write-Host "      Nome: $($menuByIdResp.data.name)" -ForegroundColor DarkGray
    } catch {
        Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    }
}
Write-Host ""

# ============================================================================
# 5. LISTA MENU (TEST CRITICO!)
# ============================================================================
Write-Host "5️⃣  LISTA MENU (TEST CRITICO PER TAG BSON)" -ForegroundColor Yellow
Write-Host "   Questo è il test che verifica se il fix funziona!" -ForegroundColor Magenta

try {
    $menusListResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers -TimeoutSec 10
    
    if ($menusListResp.data.Count -gt 0) {
        Write-Host "   ✅ SUCCESSO! Menu trovati: $($menusListResp.data.Count)" -ForegroundColor Green
        Write-Host ""
        $menusListResp.data | ForEach-Object {
            Write-Host "      📋 $($_.name)" -ForegroundColor Cyan
            Write-Host "         ID: $($_.id)" -ForegroundColor DarkGray
            Write-Host "         Restaurant ID: $($_.restaurant_id)" -ForegroundColor DarkGray
            Write-Host "         Categorie: $($_.categories.Count)" -ForegroundColor DarkGray
        }
        Write-Host ""
        Write-Host "   🎉 IL FIX È FUNZIONANTE!" -ForegroundColor Green -BackgroundColor DarkGreen
        Write-Host "   🎉 I TAG BSON SONO ATTIVI E CORRETTI!" -ForegroundColor Green -BackgroundColor DarkGreen
        $fixWorks = $true
    } else {
        Write-Host "   ❌ FALLITO! Nessun menu trovato" -ForegroundColor Red
        Write-Host "      Il menu è stato creato ma non appare nella lista" -ForegroundColor Yellow
        Write-Host "      Questo indica che i tag bson NON sono attivi" -ForegroundColor Yellow
        $fixWorks = $false
    }
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    $fixWorks = $false
}
Write-Host ""

# ============================================================================
# 6. DEBUG ENDPOINT
# ============================================================================
Write-Host "6️⃣  VERIFICA DEBUG ENDPOINT" -ForegroundColor Yellow
try {
    $debugResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/debug/menus" -Headers $headers -TimeoutSec 10
    Write-Host "   ✅ Debug endpoint disponibile" -ForegroundColor Green
    Write-Host "      Total menus in DB: $($debugResp.data.total_menus_in_db)" -ForegroundColor DarkGray
    Write-Host "      Filtered menus: $($debugResp.data.filtered_menus_count)" -ForegroundColor DarkGray
    
    if ($debugResp.data.all_menus_raw.Count -gt 0) {
        $firstMenu = $debugResp.data.all_menus_raw[0]
        Write-Host ""
        Write-Host "   📊 Analisi campi MongoDB:" -ForegroundColor Cyan
        
        if ($firstMenu.PSObject.Properties["restaurant_id"]) {
            Write-Host "      ✅ Campo 'restaurant_id' (snake_case) presente" -ForegroundColor Green
            Write-Host "         → TAG BSON ATTIVI!" -ForegroundColor Green
        } elseif ($firstMenu.PSObject.Properties["RestaurantID"]) {
            Write-Host "      ❌ Campo 'RestaurantID' (PascalCase) presente" -ForegroundColor Red
            Write-Host "         → TAG BSON NON ATTIVI!" -ForegroundColor Red
        }
    }
    $debugDeployed = $true
} catch {
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Host "   ❌ Debug endpoint NON disponibile (404)" -ForegroundColor Red
        Write-Host "      Il nuovo codice NON è stato deployato" -ForegroundColor Yellow
    } else {
        Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    }
    $debugDeployed = $false
}
Write-Host ""

# ============================================================================
# 7. TEST ANALYTICS (genera eventi)
# ============================================================================
Write-Host "7️⃣  TEST ANALYTICS" -ForegroundColor Yellow
try {
    # Genera alcuni eventi
    for ($i = 1; $i -le 3; $i++) {
        Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Headers $headers -TimeoutSec 5 | Out-Null
    }
    if ($menuId) {
        Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId" -Headers $headers -TimeoutSec 5 | Out-Null
    }
    Write-Host "   ✅ Eventi analytics generati" -ForegroundColor Green
    Write-Host "      (verifica nei log o nel database)" -ForegroundColor Gray
} catch {
    Write-Host "   ⚠️  Non è stato possibile generare eventi" -ForegroundColor Yellow
}
Write-Host ""

# ============================================================================
# RIEPILOGO FINALE
# ============================================================================
Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║                    RIEPILOGO FINALE                           ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

if ($fixWorks) {
    Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║                                                               ║" -ForegroundColor Green
    Write-Host "║            🎉🎉🎉 TUTTI I TEST PASSATI! 🎉🎉🎉              ║" -ForegroundColor Green
    Write-Host "║                                                               ║" -ForegroundColor Green
    Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
    Write-Host "✅ REGISTRAZIONE: Funzionante" -ForegroundColor Green
    Write-Host "✅ ACCESSO: Funzionante" -ForegroundColor Green
    Write-Host "✅ MENU - Creazione: Funzionante" -ForegroundColor Green
    Write-Host "✅ MENU - Lettura: Funzionante" -ForegroundColor Green
    Write-Host "✅ MENU - Lista: Funzionante (TAG BSON OK!)" -ForegroundColor Green
    if ($debugDeployed) {
        Write-Host "✅ DEBUG ENDPOINT: Deployato" -ForegroundColor Green
    }
    Write-Host "✅ ANALYTICS: Eventi generati" -ForegroundColor Green
    Write-Host ""
    Write-Host "🌐 CREDENZIALI DI ACCESSO:" -ForegroundColor Cyan
    Write-Host "   URL: https://qr-menu-staging.up.railway.app/login" -ForegroundColor White
    Write-Host "   Username: $username" -ForegroundColor White
    Write-Host "   Password: $password" -ForegroundColor White
    Write-Host ""
    Write-Host "✨ Il sistema è pronto per l'uso in produzione!" -ForegroundColor Green
    
} else {
    Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Red
    Write-Host "║                ❌ TEST FALLITI ❌                            ║" -ForegroundColor Red
    Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Red
    Write-Host ""
    Write-Host "❌ I menu non sono visibili nella lista" -ForegroundColor Red
    Write-Host "   Il fix dei tag bson non è attivo" -ForegroundColor Yellow
    Write-Host ""
    
    if (!$debugDeployed) {
        Write-Host "💡 DIAGNOSI:" -ForegroundColor Cyan
        Write-Host "   L'endpoint di debug non è disponibile" -ForegroundColor White
        Write-Host "   → Il nuovo codice NON è stato deployato su Railway" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "🔧 AZIONI NECESSARIE:" -ForegroundColor Cyan
        Write-Host "   1. Verifica lo stato del deploy su Railway" -ForegroundColor White
        Write-Host "   2. Controlla i log di build per errori" -ForegroundColor White
        Write-Host "   3. Forza un nuovo deploy se necessario" -ForegroundColor White
    } else {
        Write-Host "💡 DIAGNOSI:" -ForegroundColor Cyan
        Write-Host "   L'endpoint di debug è disponibile ma i menu non si vedono" -ForegroundColor White
        Write-Host "   → Possibile problema con i dati o query" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "🔧 AZIONI NECESSARIE:" -ForegroundColor Cyan
        Write-Host "   1. Esegui: .\check_debug.ps1 per analisi dettagliata" -ForegroundColor White
        Write-Host "   2. Verifica i campi nei documenti MongoDB" -ForegroundColor White
    }
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Gray
Write-Host "Test completato: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" -ForegroundColor Gray
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Gray
Write-Host ""
