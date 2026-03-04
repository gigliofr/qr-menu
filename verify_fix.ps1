$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json

Write-Host "🔐 Login..." -ForegroundColor Cyan
$loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginResp.data.token
$headers = @{"Authorization" = "Bearer $token"}

Write-Host "✅ Login OK" -ForegroundColor Green
Write-Host ""

Write-Host "📋 Recupero menu..." -ForegroundColor Cyan
$menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers

Write-Host ""
Write-Host "=" * 70 -ForegroundColor Cyan
if ($menusResp.data.Count -gt 0) {
    Write-Host "✅ SUCCESSO! Menu trovati: $($menusResp.data.Count)" -ForegroundColor Green
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host ""
    
    $menusResp.data | ForEach-Object {
        Write-Host "📋 $($_.name)" -ForegroundColor White
        Write-Host "   ID: $($_.id)" -ForegroundColor Gray
        Write-Host "   Restaurant ID: $($_.restaurant_id)" -ForegroundColor Gray
        Write-Host "   Categorie: $($_.categories.Count)" -ForegroundColor Gray
        Write-Host "   Piatti totali: $(($_.categories | ForEach-Object { $_.items.Count } | Measure-Object -Sum).Sum)" -ForegroundColor Gray
        Write-Host "   Creato: $($_.created_at)" -ForegroundColor Gray
        Write-Host ""
    }
    
    Write-Host "🎉 IL PROBLEMA È RISOLTO!" -ForegroundColor Green
    Write-Host "   I tag bson funzionano correttamente" -ForegroundColor White
    Write-Host "   La query per restaurant_id trova i menu" -ForegroundColor White
    Write-Host ""
    Write-Host "✅ Ora puoi accedere all'app e vedere i menu!" -ForegroundColor Green
    Write-Host "   URL: https://qr-menu-staging.up.railway.app/login" -ForegroundColor White
} else {
    Write-Host "❌ Nessun menu trovato" -ForegroundColor Red
    Write-Host "=" * 70 -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Questo potrebbe significare:" -ForegroundColor Yellow
    Write-Host "1. Railway non ha ancora deployato il codice con i tag bson" -ForegroundColor White
    Write-Host "2. C'è ancora un problema nel codice" -ForegroundColor White
    Write-Host ""
    Write-Host "Response completa:" -ForegroundColor Gray
    $menusResp | ConvertTo-Json -Depth 10
}
