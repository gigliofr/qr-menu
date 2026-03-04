$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "🚀 Monitoraggio Deploy Railway" -ForegroundColor Cyan
Write-Host "=" * 70 -ForegroundColor Gray
Write-Host ""
Write-Host "Commit pushato: a56b898" -ForegroundColor Yellow
Write-Host "Railway sta rebuilding l'applicazione..." -ForegroundColor Gray
Write-Host ""

$maxAttempts = 40  # 40 * 10s = ~6-7 minuti
$attempt = 0
$deployOK = $false

while ($attempt -lt $maxAttempts -and -not $deployOK) {
    $attempt++
    Start-Sleep -Seconds 10
    
    Write-Host "[$attempt/$maxAttempts] Controllo deploy..." -ForegroundColor Gray
    
    try {
        # Prova a chiamare l'endpoint di debug
        $loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
        $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json" -TimeoutSec 5
        $token = $loginResp.data.token
        $headers = @{"Authorization" = "Bearer $token"}
        
        $debugResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/debug/menus" -Headers $headers -TimeoutSec 5
        
        # Se arriviamo qui, l'endpoint è disponibile
        Write-Host ""
        Write-Host "✅ Deploy completato! (dopo $($attempt * 10) secondi)" -ForegroundColor Green
        Write-Host ""
        Write-Host "=" * 70 -ForegroundColor Cyan
        Write-Host "DEBUG INFO" -ForegroundColor Cyan
        Write-Host "=" * 70 -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Restaurant ID: $($debugResp.data.restaurant_id)" -ForegroundColor Yellow
        Write-Host "Total menus in DB: $($debugResp.data.total_menus_in_db)" -ForegroundColor Yellow
        Write-Host "Filtered menus: $($debugResp.data.filtered_menus_count)" -ForegroundColor Yellow
        Write-Host ""
        
        if ($debugResp.data.filtered_menus_count -gt 0) {
            Write-Host "🎉 SUCCESSO! I menu sono visibili!" -ForegroundColor Green
            Write-Host "=" * 70 -ForegroundColor Green
            Write-Host ""
            $debugResp.data.filtered_menus_decoded | ForEach-Object {
                Write-Host "📋 $($_.name)" -ForegroundColor White
                Write-Host "   ID: $($_.id)" -ForegroundColor Gray
                Write-Host "   Categorie: $($_.categories.Count)" -ForegroundColor Gray
            }
            Write-Host ""
            Write-Host "✅ Ora puoi accedere all'app:" -ForegroundColor Green
            Write-Host "   https://qr-menu-staging.up.railway.app/login" -ForegroundColor White
        } elseif ($debugResp.data.total_menus_in_db -gt 0) {
            Write-Host "⚠️  Menu presenti ma non filtrati correttamente" -ForegroundColor Yellow
            Write-Host "   Questo significa che c'è ancora un problema nel codice" -ForegroundColor Red
            Write-Host ""
            Write-Host "Menu nel DB:" -ForegroundColor Cyan
            $debugResp.data.all_menus_raw | ForEach-Object {
                Write-Host "  - $($_['name'])" -ForegroundColor White
                Write-Host "    restaurant_id: $($_['restaurant_id'])" -ForegroundColor Gray
            }
        } else {
            Write-Host "⚠️  Nessun menu nel database" -ForegroundColor Yellow
            Write-Host "   Devi eseguire: .\setup_ristorante_completo.ps1" -ForegroundColor Cyan
        }
        
        $deployOK = $true
        
    } catch {
        # Deploy non ancora completato, continua
    }
}

if (-not $deployOK) {
    Write-Host ""
    Write-Host "⚠️  Timeout raggiunto" -ForegroundColor Red
    Write-Host "   Il deploy potrebbe richiedere più tempo" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Verifica manualmente su Railway:" -ForegroundColor Cyan
    Write-Host "https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab" -ForegroundColor White
}

Write-Host ""
Write-Host "🏁 Monitoraggio terminato" -ForegroundColor Cyan
