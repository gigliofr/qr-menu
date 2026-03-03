$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"}

Write-Host "🔬 Test Diagnostico BSON Tags" -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Gray
Write-Host ""

# 1. Crea un nuovo menu
Write-Host "STEP 1: Creazione menu di test..." -ForegroundColor Yellow
$menuData = @{
    name = "Menu Diagnostico $(Get-Date -Format 'HH:mm:ss')"
    description = "Test BSON"
    meal_type = "lunch"
    categories = @(
        @{
            id = (New-Guid).ToString()
            name = "Test"
            description = "Test"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Piatto Test"
                    description = "Test"
                    price = 10.0
                    category = "Test"
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

$createResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post -Body $menuData -Headers $headers
$menuId = $createResp.data.id
$restaurantIdFromCreate = $createResp.data.restaurant_id

Write-Host "✅ Menu creato" -ForegroundColor Green
Write-Host "   ID menu: $menuId" -ForegroundColor White
Write-Host "   Restaurant ID nella risposta: $restaurantIdFromCreate" -ForegroundColor White

# 2. Recupera lo stesso menu per ID
Write-Host ""
Write-Host "STEP 2: Recupero menu per ID specifico..." -ForegroundColor Yellow
try {
    $menuByIdResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId" -Headers $headers
    Write-Host "✅ Menu recuperato per ID!" -ForegroundColor Green
    Write-Host "   Nome: $($menuByIdResp.data.name)" -ForegroundColor White
    Write-Host "   Restaurant ID nel documento: $($menuByIdResp.data.restaurant_id)" -ForegroundColor White
    
    $restaurantIdInDb = $menuByIdResp.data.restaurant_id
} catch {
    Write-Host "❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    $restaurantIdInDb = $null
}

# 3. Recupera tutti i menu (per restaurant_id)
Write-Host ""
Write-Host "STEP 3: Recupero tutti i menu (query per restaurant_id)..." -ForegroundColor Yellow
$allMenusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
Write-Host "Menu trovati: $($allMenusResp.data.Count)" -ForegroundColor $(if ($allMenusResp.data.Count -gt 0) {"Green"} else {"Red"})

# 4. Verifica restaurant ID dal profilo
Write-Host ""
Write-Host "STEP 4: Verifica Restaurant ID dal profilo..." -ForegroundColor Yellow
$profileResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Headers $headers
$restaurantIdFromProfile = $profileResp.data.id
Write-Host "Restaurant ID dal profilo: $restaurantIdFromProfile" -ForegroundColor White

# 5. Analisi
Write-Host ""
Write-Host "=" * 60 -ForegroundColor Gray
Write-Host "📊 ANALISI" -ForegroundColor Cyan
Write-Host ""

if ($restaurantIdFromCreate -eq $restaurantIdFromProfile) {
    Write-Host "✅ Restaurant ID nella risposta POST corrisponde al profilo" -ForegroundColor Green
} else {
    Write-Host "❌ PROBLEMA: Restaurant ID nella risposta POST NON corrisponde!" -ForegroundColor Red
    Write-Host "   POST response: $restaurantIdFromCreate" -ForegroundColor Yellow
    Write-Host "   Profilo:       $restaurantIdFromProfile" -ForegroundColor Yellow
}

if ($restaurantIdInDb) {
    if ($restaurantIdInDb -eq $restaurantIdFromProfile) {
        Write-Host "✅ Restaurant ID nel DB (recuperato per ID) corrisponde" -ForegroundColor Green
    } else {
        Write-Host "❌ PROBLEMA: Restaurant ID nel DB NON corrisponde!" -ForegroundColor Red
        Write-Host "   Nel DB:  $restaurantIdInDb" -ForegroundColor Yellow
        Write-Host "   Profilo: $restaurantIdFromProfile" -ForegroundColor Yellow
    }
}

if ($allMenusResp.data.Count -eq 0) {
    Write-Host ""
    Write-Host "❌ PROBLEMA PRINCIPALE: Query per restaurant_id non trova menu!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Possibili cause:" -ForegroundColor Yellow
    Write-Host "   1. Railway non ha ancora deployato il codice con tag bson" -ForegroundColor White
    Write-Host "   2. Il campo 'restaurant_id' nel DB ha un nome diverso" -ForegroundColor White
    Write-Host "   3. C'è un bug nella query MongoDB" -ForegroundColor White
    Write-Host ""
    Write-Host "AZIONE CONSIGLIATA:" -ForegroundColor Cyan
    Write-Host "   Attendi il deploy Railway (2-3 minuti) ed esegui:" -ForegroundColor White
    Write-Host "   .\wait_for_deploy.ps1" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "✅ FIX FUNZIONANTE! La query per restaurant_id trova i menu!" -ForegroundColor Green
}

Write-Host ""
