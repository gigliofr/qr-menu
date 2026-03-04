$baseUrl = "https://qr-menu-staging.up.railway.app"

# Credenziali utente
$username = "testuser_075850"
$password = "TestPassword123!"

Write-Host ""
Write-Host "╔════════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║         VERIFICA API MENU - Debug Approfondito               ║" -ForegroundColor Cyan
Write-Host "╚════════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
Write-Host ""

# Login
Write-Host "1️⃣  LOGIN" -ForegroundColor Yellow
$loginBody = @{
    username = $username
    password = $password
} | ConvertTo-Json

try {
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResp.data.token
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    Write-Host "   ✅ Login OK" -ForegroundColor Green
    Write-Host "      Restaurant ID: $($loginResp.data.restaurant.id)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ Login fallito: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# GET /menus
Write-Host "2️⃣  GET /api/v1/menus" -ForegroundColor Yellow
try {
    $menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
    Write-Host "   ✅ Risposta ricevuta" -ForegroundColor Green
    Write-Host "      Status: 200 OK" -ForegroundColor Gray
    Write-Host "      Menu Count: $($menusResp.data.Count)" -ForegroundColor Gray
    Write-Host ""
    
    if ($menusResp.data.Count -gt 0) {
        Write-Host "   📋 MENU TROVATI:" -ForegroundColor Green
        $menusResp.data | ForEach-Object {
            Write-Host "      • $($_.name)" -ForegroundColor Cyan
            Write-Host "        ID: $($_.id)" -ForegroundColor DarkGray
            Write-Host "        Restaurant ID: $($_.restaurant_id)" -ForegroundColor DarkGray
            Write-Host "        Active: $($_.is_active)" -ForegroundColor DarkGray
            Write-Host "        Completed: $($_.is_completed)" -ForegroundColor DarkGray
            Write-Host "        Categorie: $($_.categories.Count)" -ForegroundColor DarkGray
            Write-Host ""
        }
    } else {
        Write-Host "   ⚠️  Nessun menu trovato!" -ForegroundColor Yellow
    }
    
    Write-Host "   📊 RISPOSTA COMPLETA:" -ForegroundColor Cyan
    $menusResp | ConvertTo-Json -Depth 10 | Write-Host -ForegroundColor DarkGray
    
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "      Dettaglio: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
    }
}
Write-Host ""

# GET /restaurant/profile
Write-Host "3️⃣  GET /api/v1/restaurant/profile" -ForegroundColor Yellow
try {
    $profileResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/restaurant/profile" -Headers $headers
    Write-Host "   ✅ Profilo recuperato" -ForegroundColor Green
    Write-Host "      Restaurant ID: $($profileResp.data.id)" -ForegroundColor Gray
    Write-Host "      Username: $($profileResp.data.username)" -ForegroundColor Gray
    Write-Host "      Email: $($profileResp.data.email)" -ForegroundColor Gray
    Write-Host "      Active Menu ID: $($profileResp.data.active_menu_id)" -ForegroundColor Gray
    Write-Host ""
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Debug endpoint
Write-Host "4️⃣  GET /api/v1/debug/menus" -ForegroundColor Yellow
try {
    $debugResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/debug/menus" -Headers $headers
    Write-Host "   ✅ Debug info recuperata" -ForegroundColor Green
    Write-Host "      Total menus in DB: $($debugResp.data.total_menus_in_db)" -ForegroundColor Gray
    Write-Host "      Filtered menus: $($debugResp.data.filtered_menus_count)" -ForegroundColor Gray
    Write-Host "      Restaurant ID filter: $($debugResp.data.restaurant_id)" -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "   📊 RAW MENUS:" -ForegroundColor Cyan
    $debugResp.data.all_menus_raw | ForEach-Object {
        Write-Host "      Menu: $($_.name)" -ForegroundColor DarkCyan
        Write-Host "         Campi disponibili:" -ForegroundColor DarkGray
        $_.PSObject.Properties | ForEach-Object {
            Write-Host "            $($_.Name): $($_.Value)" -ForegroundColor DarkGray
        }
        Write-Host ""
    }
} catch {
    Write-Host "   ❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Gray
Write-Host "Verifica completata" -ForegroundColor Gray
Write-Host "═══════════════════════════════════════════════════════════════" -ForegroundColor Gray
