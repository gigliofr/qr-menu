$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json

Write-Host "🔐 Login..." -ForegroundColor Cyan
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"}
Write-Host "✅ Login OK" -ForegroundColor Green
Write-Host ""

# Crea un menu semplice
Write-Host "📝 Creazione menu di test..." -ForegroundColor Cyan
$menuData = @{
    name = "Menu Test Fix $(Get-Date -Format 'HH:mm:ss')"
    description = "Test dopo eliminazione collections"
    meal_type = "lunch"
    categories = @(
        @{
            id = (New-Guid).ToString()
            name = "Primi"
            description = "Primi piatti"
            items = @(
                @{
                    id = (New-Guid).ToString()
                    name = "Carbonara"
                    description = "Classica"
                    price = 12.0
                    category = "Primi"
                    available = $true
                },
                @{
                    id = (New-Guid).ToString()
                    name = "Amatriciana"
                    description = "Con guanciale"
                    price = 11.0
                    category = "Primi"
                    available = $true
                }
            )
        }
    )
} | ConvertTo-Json -Depth 10

try {
    $createResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post -Body $menuData -Headers $headers
    Write-Host "✅ Menu creato!" -ForegroundColor Green
    Write-Host "   ID: $($createResp.data.id)" -ForegroundColor Gray
    Write-Host "   Nome: $($createResp.data.name)" -ForegroundColor Gray
    $menuId = $createResp.data.id
} catch {
    Write-Host "❌ Errore creazione: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🔍 Test 1: Recupero per ID specifico..." -ForegroundColor Cyan
try {
    $menuById = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus/$menuId" -Headers $headers
    Write-Host "✅ Menu trovato per ID!" -ForegroundColor Green
    Write-Host "   Nome: $($menuById.data.name)" -ForegroundColor Gray
    Write-Host "   Restaurant ID nel documento: $($menuById.data.restaurant_id)" -ForegroundColor Gray
} catch {
    Write-Host "❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "🔍 Test 2: Recupero tutti i menu (query per restaurant_id)..." -ForegroundColor Cyan
$allMenus = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers

Write-Host ""
Write-Host "=" * 70 -ForegroundColor Cyan
if ($allMenus.data.Count -gt 0) {
    Write-Host "✅ ✅ ✅ FUNZIONA! ✅ ✅ ✅" -ForegroundColor Green
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Menu trovati: $($allMenus.data.Count)" -ForegroundColor Yellow
    $allMenus.data | ForEach-Object {
        Write-Host "  📋 $($_.name)" -ForegroundColor White
        Write-Host "     Categorie: $($_.categories.Count)" -ForegroundColor Gray
    }
    Write-Host ""
    Write-Host "🎉 IL FIX È FUNZIONANTE!" -ForegroundColor Green
    Write-Host "   I tag bson sono attivi" -ForegroundColor White
    Write-Host "   La query per restaurant_id funziona" -ForegroundColor White
} else {
    Write-Host "❌ ❌ ❌ PROBLEMA PERSISTE ❌ ❌ ❌" -ForegroundColor Red
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Il menu è stato creato ma non viene trovato dalla GET." -ForegroundColor Yellow
    Write-Host "Railway probabilmente non ha ancora deployato il codice con i tag bson." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Aspetta 1-2 minuti e riprova:" -ForegroundColor Cyan
    Write-Host ".\check_debug.ps1" -ForegroundColor Gray
}
