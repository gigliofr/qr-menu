$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"}

Write-Host "🔍 Chiamando endpoint di debug..." -ForegroundColor Cyan
try {
    $debugResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/debug/menus" -Headers $headers
    Write-Host "✅ Endpoint di debug disponibile!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Restaurant ID: $($debugResp.data.restaurant_id)" -ForegroundColor Yellow
    Write-Host "Total menus in DB: $($debugResp.data.total_menus_in_db)" -ForegroundColor Yellow
    Write-Host "Filtered menus: $($debugResp.data.filtered_menus_count)" -ForegroundColor Yellow
    Write-Host ""
    
    if ($debugResp.data.total_menus_in_db -gt 0) {
        Write-Host "📄 Menu nel database:" -ForegroundColor Cyan
        $debugResp.data.all_menus_raw | ForEach-Object {
            Write-Host "  - $($_['name'])" -ForegroundColor White
            Write-Host "    restaurant_id: $($_['restaurant_id'])" -ForegroundColor Gray
            Write-Host "    RestaurantID: $($_['RestaurantID'])" -ForegroundColor Gray
            
            # Mostra tutti i campi
            Write-Host "    Tutti i campi:" -ForegroundColor DarkGray
            $_.Keys | Sort-Object | ForEach-Object {
                Write-Host "      $_" -ForegroundColor DarkGray
            }
            Write-Host ""
        }
    } else {
        Write-Host "❌ Nessun menu nel database!" -ForegroundColor Red
        Write-Host "   Significa che setup_ristorante_completo.ps1 non ha creato menu" -ForegroundColor Yellow
    }
    
    Write-Host "Debug completo:" -ForegroundColor Gray
    $debugResp.data | ConvertTo-Json -Depth 10
} catch {
    Write-Host "❌ Endpoint di debug non disponibile: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Railway sta ancora deployando il nuovo codice..." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "   Aspetta 30-60 secondi e riprova:" -ForegroundColor Cyan
    Write-Host "   .\check_debug.ps1" -ForegroundColor Gray
}
