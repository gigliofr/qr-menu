# Aspetta il deploy e poi monitora i log dal browser Railway

$baseUrl = "https://qr-menu-staging.up.railway.app"

Write-Host "⏳ Aspettando deploy Railway..." -ForegroundColor Cyan
Write-Host "   Commit: 72b67f4 (Debug logs)" -ForegroundColor Gray
Write-Host ""

$deployed = $false
$attempts = 0
$maxAttempts = 30  # 5 minuti

while (-not $deployed -and $attempts -lt $maxAttempts) {
    $attempts++
    Start-Sleep -Seconds 10
    
    try {
        $health = Invoke-RestMethod -Uri "$baseUrl/health" -TimeoutSec 5 -ErrorAction Stop
        
        # Se l'app risponde, probabilmente il deploy è completato
        if ($attempts -gt 2) {
            Write-Host "[✓] App online dopo $($attempts * 10) secondi" -ForegroundColor Green
            $deployed = $true
        } else {
            Write-Host "[$attempts/$maxAttempts] Waiting..." -ForegroundColor Gray
        }
    } catch {
        Write-Host "[$attempts/$maxAttempts] App non disponibile (rebuild in corso)..." -ForegroundColor Yellow
    }
}

if ($deployed) {
    Write-Host ""
    Write-Host "✅ Deploy completato!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🧪 Eseguo test con i nuovi log..." -ForegroundColor Cyan
    
    # Login
    $loginBody = @{username = "trattoria_roma"; password = "RomaTest2026!"} | ConvertTo-Json
    $loginResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginResp.data.token
    $headers = @{"Authorization" = "Bearer $token"}
    
    # Chiama GET /menus per triggherare i log
    Write-Host "Chiamando GET /menus..." -ForegroundColor Gray
    $menusResp = Invoke-RestMethod -Uri "$baseUrl/api/v1/menus" -Headers $headers
    Write-Host "Menu trovati: $($menusResp.data.Count)" -ForegroundColor Yellow
    
    Write-Host ""
    Write-Host "📊 Per vedere i log di debug:" -ForegroundColor Cyan
    Write-Host "   1. Apri: https://railway.com/project/6c597b1a-4973-48af-bd4e-818e88568bab" -ForegroundColor White
    Write-Host "   2. Seleziona il servizio 'qr-menu'" -ForegroundColor White
    Write-Host "   3. Tab 'Deployments' → Ultimo deploy → 'View Logs'" -ForegroundColor White
    Write-Host ""
    Write-Host "🔍 Cerca nei log:" -ForegroundColor Yellow
    Write-Host "   - '🔍 GetMenusByRestaurantID' per vedere la query" -ForegroundColor Gray
    Write-Host "   - '✅ Trovati' per vedere i risultati" -ForegroundColor Gray
    Write-Host "   - 'GetMenusHandler chiamato' per vedere il restaurant_id" -ForegroundColor Gray
    Write-Host ""
    
    if ($menusResp.data.Count -gt 0) {
        Write-Host "✅ FIX FUNZIONANTE! Menu visibili!" -ForegroundColor Green
    } else {
        Write-Host "❌ Menu ancora non visibili - controlla i log!" -ForegroundColor Red
    }
} else {
    Write-Host ""
    Write-Host "⚠️  Timeout - deploy potrebbe richiedere più tempo" -ForegroundColor Yellow
    Write-Host "   Controlla manualmente su Railway" -ForegroundColor White
}

Write-Host ""
