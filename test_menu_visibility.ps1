$baseUrl = "https://qr-menu-staging.up.railway.app"
$loginBody = @{
    username = "trattoria_roma"
    password = "RomaTest2026!"
} | ConvertTo-Json

Write-Host "🔐 Tentativo di login con account trattoria_roma..." -ForegroundColor Cyan
try {
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResp.data.token
    Write-Host "✅ Login riuscito! Token: $($token.Substring(0,20))..." -ForegroundColor Green
    
    Write-Host ""
    Write-Host "📋 Recupero menu..." -ForegroundColor Cyan
    $headers = @{"Authorization" = "Bearer $token"}
    $menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
    
    Write-Host "Menus trovati: $($menusResp.data.Count)" -ForegroundColor Yellow
    if ($menusResp.data.Count -gt 0) {
        Write-Host ""
        Write-Host "✅ MENU VISIBILI!" -ForegroundColor Green
        $menusResp.data | ForEach-Object {
            Write-Host "   - $($_.name) (ID: $($_.id))" -ForegroundColor White
            Write-Host "     Categorie: $($_.categories.Count)" -ForegroundColor Gray
        }
    } else {
        Write-Host "❌ Nessun menu trovato!" -ForegroundColor Red
        Write-Host "Response completa:" -ForegroundColor Yellow
        $menusResp | ConvertTo-Json -Depth 10
    }
} catch {
    Write-Host "❌ Errore: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "Dettagli: $($_.ErrorDetails.Message)" -ForegroundColor Yellow
    }
}
