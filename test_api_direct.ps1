# Test API dirette per debug
$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "`n=== TEST API REGISTRAZIONE E LOGIN ===" -ForegroundColor Cyan

# Genera dati unici
$ts = Get-Date -Format "HHmmss"
$username = "apitest$ts"
$email = "apitest$ts@test.com"
$password = "TestPass123!"

Write-Host "`nDati test:" -ForegroundColor Yellow
Write-Host "  Username: $username"
Write-Host "  Email: $email"
Write-Host "  Password: $password"

# Test 1: API Register
Write-Host "`n1. Test API /api/v1/auth/register..." -ForegroundColor Yellow
try {
    $registerBody = @{
        username = $username
        email = $email
        password = $password
        confirm_password = $password
        restaurant_name = "Ristorante API Test $ts"
        description = "Test automatico"
    } | ConvertTo-Json

    Write-Host "   Body: $registerBody" -ForegroundColor Gray
    
    $registerResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/register" -Method Post `
        -Body $registerBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "   ✅ Registrazione API riuscita!" -ForegroundColor Green
    Write-Host "   Response: $($registerResp | ConvertTo-Json -Depth 3)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Gray
    }
}

# Test 2: Login immediato
Write-Host "`n2. Test API /api/v1/auth/login (con username)..." -ForegroundColor Yellow
Start-Sleep -Seconds 2
try {
    $loginBody = @{
        username = $username
        password = $password
    } | ConvertTo-Json

    Write-Host "   Body: $loginBody" -ForegroundColor Gray
    
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post `
        -Body $loginBody `
        -ContentType "application/json" `
        -ErrorAction Stop

    Write-Host "   ✅ Login riuscito!" -ForegroundColor Green
    Write-Host "   Token: $($loginResp.data.token.Substring(0, 30))..." -ForegroundColor Gray
    Write-Host "   User: $($loginResp.data.user.username)" -ForegroundColor Gray
    
    $token = $loginResp.data.token
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # Test 3: Crea menu
    Write-Host "`n3. Test creazione menu..." -ForegroundColor Yellow
    $menuBody = @{
        name = "Menu Test API $ts"
        description = "Menu creato via API"
        active = $true
        categories = @(
            @{
                name = "Piatti"
                items = @(
                    @{
                        name = "Test Dish"
                        description = "Piatto di test"
                        price = 10.00
                        available = $true
                    }
                )
            }
        )
    } | ConvertTo-Json -Depth 10

    $menuResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Post `
        -Body $menuBody `
        -Headers $headers

    Write-Host "   ✅ Menu creato! ID: $($menuResp.data._id)" -ForegroundColor Green

    # Test 4: Get menus
    Write-Host "`n4. Test recupero menus..." -ForegroundColor Yellow
    $menus = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Method Get -Headers $headers
    Write-Host "   ✅ Menus: $($menus.data.Count)" -ForegroundColor Green

    # Test 5: Get profilo
    Write-Host "`n5. Test recupero profilo..." -ForegroundColor Yellow
    $profile = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Method Get -Headers $headers
    Write-Host "   ✅ Ristorante: $($profile.data.name)" -ForegroundColor Green

    Write-Host "`n✅ TUTTI I TEST PASSATI!" -ForegroundColor Green
    Write-Host "`n🔐 CREDENZIALI FUNZIONANTI:" -ForegroundColor Cyan
    Write-Host "   Username: $username" -ForegroundColor White
    Write-Host "   Email:    $email" -ForegroundColor White
    Write-Host "   Password: $password" -ForegroundColor White

} catch {
    Write-Host "   ❌ Login fallito: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "   Details: $($_.ErrorDetails.Message)" -ForegroundColor Gray
    }
}

Write-Host ""
