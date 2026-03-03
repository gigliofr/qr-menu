# Monitora il deploy Railway e testa quando è pronto

$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "🚀 Monitoraggio Deploy Railway" -ForegroundColor Cyan
Write-Host "=" * 60 -ForegroundColor Gray
Write-Host ""
Write-Host "Commit locale: fab5785 (appena pushato)" -ForegroundColor Yellow
Write-Host "Attendo che Railway completi il rebuild..." -ForegroundColor Gray
Write-Host ""

$deployCompleted = $false
$maxAttempts = 20  # 20 tentativi = ~  2 minuti
$attemptCount = 0

while (-not $deployCompleted -and $attemptCount -lt $maxAttempts) {
    $attemptCount++
    Start-Sleep -Seconds 10
    
    try {
        $health = Invoke-RestMethod -Uri "$baseUrl/health" -TimeoutSec 5
        $timestamp = $health.data.timestamp
        
        Write-Host "[$attemptCount/$maxAttempts] Health check OK - $timestamp" -ForegroundColor Gray
        
        # Se l'app risponde, il deploy è probabilmente completato
        # Testa subito se il fix funziona
        if ($attemptCount -ge 2) {  # Aspetta almeno 20 secondi
            Write-Host ""
            Write-Host "🧪 Test del fix bson tags..." -ForegroundColor Cyan
            
            # Login
            $loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
            $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
            $token = $loginResp.data.token
            $headers = @{"Authorization" = "Bearer $token"}
            
            # Recupera menu
            $menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
            
            if ($menusResp.data.Count -gt 0) {
                Write-Host ""
                Write-Host "✅ SUCCESSO! Menu visibili: $($menusResp.data.Count)" -ForegroundColor Green
                Write-Host ""
                $menusResp.data | ForEach-Object {
                    Write-Host "   📋 $($_.name)" -ForegroundColor White
                    Write-Host "      ID: $($_.id)" -ForegroundColor Gray
                    Write-Host "      Restaurant: $($_.restaurant_id)" -ForegroundColor Gray
                    Write-Host "      Categorie: $($_.categories.Count)" -ForegroundColor Gray
                }
                $deployCompleted = $true
            } else {
                Write-Host "   ⏳ Deploy ancora in corso (menu non visibili)..." -ForegroundColor Yellow
            }
        }
    } catch {
        Write-Host "[$attemptCount/$maxAttempts] App non disponibile (rebuild in corso)..." -ForegroundColor Yellow
    }
}

if (-not $deployCompleted) {
    Write-Host ""
    Write-Host "⚠️  Timeout raggiunto" -ForegroundColor Red
    Write-Host "   Il deploy potrebbe richiedere più tempo del previsto." -ForegroundColor Yellow
    Write-Host "   Vai su Railway per verificare lo stato:" -ForegroundColor Yellow
    Write-Host "   https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab" -ForegroundColor White
    Write-Host ""
    Write-Host "   Oppure riprova manualmente tra qualche minuto:" -ForegroundColor Yellow
    Write-Host "   .\test_menu_visibility.ps1" -ForegroundColor Gray
}

Write-Host ""
Write-Host "🏁 Monitoraggio terminato" -ForegroundColor Cyan
